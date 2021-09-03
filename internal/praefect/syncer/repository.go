package syncer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
)

// StateOwner performs check for the existence of the repositories.
type StateOwner interface {
	// Exists checks if repositories exist and return a boolean flag associated with requested entry.
	Exists(ctx context.Context, paths []datastore.RepositoryClusterPath) (map[datastore.RepositoryClusterPath]bool, error)
}

// Acquirer allows to acquire storage for processing so no any other Acquirer can acquire until it is released.
type Acquirer interface {
	// Populate adds provided storage into the pool of entries to acquire.
	Populate(ctx context.Context, virtualStorage, storage string) error
	// AcquireNextStorage acquires next storage based on the inactive time.
	AcquireNextStorage(ctx context.Context, inactive, updatePeriod time.Duration) (*datastore.ClusterPath, func() error, error)
}

// Action is a procedure to be executed with knowledge of repository existence.
type Action interface {
	// Perform runs actual action depending on the repository existence.
	Perform(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error
}

// RepositoryExistence scans healthy gitalies nodes for the repositories, verifies if
// found repositories are known for the praefect and runs a special action.
type RepositoryExistence struct {
	cfg           config.Config
	logger        logrus.FieldLogger
	healthChecker praefect.HealthChecker
	conns         praefect.Connections
	stateOwner    StateOwner
	acquirer      Acquirer
	action        Action
}

// NewRepositoryExistence returns instance of the RepositoryExistence.
func NewRepositoryExistence(cfg config.Config, logger logrus.FieldLogger, healthChecker praefect.HealthChecker, conns praefect.Connections, stateOwner StateOwner, acquirer Acquirer, action Action) *RepositoryExistence {
	return &RepositoryExistence{
		cfg:           cfg,
		logger:        logger.WithField("component", "syncer.repository_existence"),
		healthChecker: healthChecker,
		conns:         conns,
		stateOwner:    stateOwner,
		acquirer:      acquirer,
		action:        action,
	}
}

// Run scans healthy gitalies nodes for the repositories, verifies if
// found repositories are known for the praefect and runs a special action.
// It runs on each tick of the provided ticker and finishes with context cancellation.
func (gs *RepositoryExistence) Run(ctx context.Context, ticker helper.Ticker) error {
	gs.logger.Info("started")
	defer gs.logger.Info("completed")

	defer ticker.Stop()

	for {
		ticker.Reset()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C():
			gs.run(ctx)
		}
	}
}

func (gs *RepositoryExistence) run(ctx context.Context) {
	for virtualStorage, nodes := range gs.healthChecker.HealthyNodes() {
		for _, storage := range nodes {
			if err := gs.acquirer.Populate(ctx, virtualStorage, storage); err != nil {
				gs.loggerWith(virtualStorage, storage).WithError(err).Error("failure on populating database")
			}

			gs.performAction(ctx)
		}
	}
}

func (gs *RepositoryExistence) performAction(ctx context.Context) {
	clusterPath, release, err := gs.acquirer.AcquireNextStorage(ctx, gs.cfg.Sync.RunInterval.Duration(), gs.cfg.Sync.LivenessInterval.Duration())
	if err != nil {
		gs.logger.WithError(err).Error("unable to acquire next storage to verify")
		return
	}

	logger := gs.logger
	defer func() {
		if err := release(); err != nil {
			logger.WithError(err).Error("failed to release storage acquired to verify")
		}
	}()

	if clusterPath == nil {
		gs.logger.Debug("no storages to verify")
		return
	}

	logger = gs.loggerWith(clusterPath.VirtualStorage, clusterPath.Storage)
	err = gs.execOnRepositories(ctx, clusterPath.VirtualStorage, clusterPath.Storage, func(paths []datastore.RepositoryClusterPath) {
		existence, err := gs.stateOwner.Exists(ctx, paths)
		if err != nil {
			logger.WithError(err).WithField("repositories", paths).Error("failed to check existence")
			return
		}

		if err := gs.action.Perform(ctx, existence); err != nil {
			logger.WithError(err).WithField("existence", existence).Error("perform action")
			return
		}
	})
	if err != nil {
		logger.WithError(err).Error("failed to exec action on repositories")
		return
	}
}

func (gs *RepositoryExistence) loggerWith(virtualStorage, storage string) logrus.FieldLogger {
	return gs.logger.WithFields(logrus.Fields{"virtual_storage": virtualStorage, "storage": storage})
}

func (gs *RepositoryExistence) execOnRepositories(ctx context.Context, virtualStorage, storage string, action func([]datastore.RepositoryClusterPath)) error {
	gclient, err := gs.getInternalGitalyClient(virtualStorage, storage)
	if err != nil {
		return fmt.Errorf("setup gitaly client: %w", err)
	}

	resp, err := gclient.WalkRepos(ctx, &gitalypb.WalkReposRequest{StorageName: storage})
	if err != nil {
		return fmt.Errorf("unable to walk repos: %w", err)
	}

	batch := make([]datastore.RepositoryClusterPath, 0, gs.cfg.Sync.RepositoriesInBatch)
	for {
		res, err := resp.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("failure on walking repos: %w", err)
			}
			break
		}

		batch = append(batch, datastore.RepositoryClusterPath{
			ClusterPath: datastore.ClusterPath{
				VirtualStorage: virtualStorage,
				Storage:        storage,
			},
			RelativePath: res.RelativePath,
		})

		if len(batch) == cap(batch) {
			action(batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		action(batch)
	}
	return nil
}

func (gs *RepositoryExistence) getInternalGitalyClient(virtualStorage, storage string) (gitalypb.InternalGitalyClient, error) {
	conn, found := gs.conns[virtualStorage][storage]
	if !found {
		return nil, fmt.Errorf("no connection to the gitaly node %q/%q", virtualStorage, storage)
	}
	return gitalypb.NewInternalGitalyClient(conn), nil
}
