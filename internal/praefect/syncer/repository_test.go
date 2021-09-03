package syncer

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/backchannel"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/gittest"
	gconfig "gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/service/setup"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore/glsql"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/service/transaction"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testserver"
)

func TestExistencer_Run(t *testing.T) {
	t.Parallel()

	const (
		repo1RelPath = "repo-1.git"
		repo2RelPath = "repo-2.git"
		repo3RelPath = "repo-3.git"

		storage1 = "gitaly-1"
		storage2 = "gitaly-2"
		storage3 = "gitaly-3"

		virtualStorage = "praefect"
	)

	g1Cfg := testcfg.Build(t, testcfg.WithStorages(storage1))
	g2Cfg := testcfg.Build(t, testcfg.WithStorages(storage2))
	g3Cfg := testcfg.Build(t, testcfg.WithStorages(storage3))

	g1Addr := testserver.RunGitalyServer(t, g1Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())
	g2Addr := testserver.RunGitalyServer(t, g2Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())
	g3Addr := testserver.RunGitalyServer(t, g3Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())

	db := glsql.NewDB(t)
	var database string
	require.NoError(t, db.QueryRow(`SELECT current_database()`).Scan(&database))
	dbConf := glsql.GetDBConfig(t, database)

	conf := config.Config{
		SocketPath: testhelper.GetTemporaryGitalySocketFileName(t),
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: virtualStorage,
				Nodes: []*config.Node{
					{Storage: g1Cfg.Storages[0].Name, Address: g1Addr},
					{Storage: g2Cfg.Storages[0].Name, Address: g2Addr},
					{Storage: g3Cfg.Storages[0].Name, Address: g3Addr},
				},
			},
		},
		DB: dbConf,
		Sync: config.Sync{
			RunInterval:         gconfig.Duration(1),
			LivenessInterval:    gconfig.Duration(1),
			RepositoriesInBatch: 2,
		},
	}

	gittest.CloneRepo(t, g1Cfg, g1Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo1RelPath})
	gittest.CloneRepo(t, g1Cfg, g1Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo2RelPath})
	gittest.CloneRepo(t, g1Cfg, g1Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo3RelPath})

	// second gitaly is missing repo-3.git repository
	gittest.CloneRepo(t, g2Cfg, g2Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo1RelPath})
	gittest.CloneRepo(t, g2Cfg, g2Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo2RelPath})

	// third gitaly has an extra repo-4.git repository
	gittest.CloneRepo(t, g3Cfg, g3Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo1RelPath})
	gittest.CloneRepo(t, g3Cfg, g3Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo2RelPath})
	gittest.CloneRepo(t, g3Cfg, g3Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo3RelPath})
	gittest.CloneRepo(t, g3Cfg, g3Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: "repo-4.git"})

	ctx, cancel := testhelper.Context()
	defer cancel()

	repoStore := datastore.NewPostgresRepositoryStore(db.DB, nil)
	for i, set := range []struct {
		relativePath string
		primary      string
		secondaries  []string
	}{
		{
			relativePath: repo1RelPath,
			primary:      storage1,
			secondaries:  []string{storage3},
		},
		{
			relativePath: repo2RelPath,
			primary:      storage1,
			secondaries:  []string{storage2, storage3},
		},
		{
			relativePath: repo3RelPath,
			primary:      storage1,
			secondaries:  []string{storage2, storage3},
		},
	} {
		require.NoError(t, repoStore.CreateRepository(ctx, int64(i), conf.VirtualStorages[0].Name, set.relativePath, set.primary, set.secondaries, nil, false, false))
	}

	logger, loggerHook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	entry := logger.WithContext(ctx)
	clientHandshaker := backchannel.NewClientHandshaker(entry, praefect.NewBackchannelServerFactory(entry, transaction.NewServer(nil)))
	nodeSet, err := praefect.DialNodes(ctx, conf.VirtualStorages, protoregistry.GitalyProtoPreregistered, nil, clientHandshaker)
	require.NoError(t, err)
	defer nodeSet.Close()

	storageCleanup := datastore.NewStorageCleanup(db.DB)

	var iteration int32
	existencer := NewRepositoryExistence(conf, logger, praefect.StaticHealthChecker{virtualStorage: []string{storage1, storage2, storage3}}, nodeSet.Connections(), storageCleanup, storageCleanup, actionStub{
		PerformMethod: func(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error {
			i := atomic.LoadInt32(&iteration)
			switch i {
			case 0:
				assert.Equal(
					t,
					map[datastore.RepositoryClusterPath]bool{
						datastore.NewRepositoryClusterPath(virtualStorage, storage1, repo1RelPath): true,
						datastore.NewRepositoryClusterPath(virtualStorage, storage1, repo2RelPath): true,
					},
					existence,
				)
			case 1:
				assert.Equal(
					t,
					map[datastore.RepositoryClusterPath]bool{datastore.NewRepositoryClusterPath(virtualStorage, storage1, repo3RelPath): true},
					existence,
				)
			case 2:
				assert.Equal(
					t,
					map[datastore.RepositoryClusterPath]bool{
						datastore.NewRepositoryClusterPath(virtualStorage, storage2, repo1RelPath): false,
						datastore.NewRepositoryClusterPath(virtualStorage, storage2, repo2RelPath): true,
					},
					existence,
				)
			case 3:
				assert.Equal(
					t,
					map[datastore.RepositoryClusterPath]bool{
						datastore.NewRepositoryClusterPath(virtualStorage, storage3, repo1RelPath): true,
						datastore.NewRepositoryClusterPath(virtualStorage, storage3, repo2RelPath): true,
					},
					existence,
				)
			case 4:
				assert.Equal(
					t,
					map[datastore.RepositoryClusterPath]bool{
						datastore.NewRepositoryClusterPath(virtualStorage, storage3, repo3RelPath): true,
						datastore.NewRepositoryClusterPath(virtualStorage, storage3, "repo-4.git"): false,
					},
					existence,
				)
			}
			atomic.AddInt32(&iteration, 1)
			return nil
		},
	})

	ticker := helper.NewManualTicker()
	done := make(chan struct{})
	go func() {
		defer close(done)
		assert.Equal(t, context.Canceled, existencer.Run(ctx, ticker))
	}()
	ticker.Tick() // triggers execution
	ticker.Tick()
	ticker.Tick() // blocks until first run will be completed
	require.Greater(t, atomic.LoadInt32(&iteration), int32(4))
	require.Len(t, loggerHook.AllEntries(), 1)
	require.Equal(
		t,
		map[string]interface{}{"Data": logrus.Fields{"component": "syncer.repository_existence"}, "Message": "started"},
		map[string]interface{}{"Data": loggerHook.AllEntries()[0].Data, "Message": loggerHook.AllEntries()[0].Message},
	)
	// Terminates the loop.
	cancel()
	<-done
	require.Equal(
		t,
		map[string]interface{}{"Data": logrus.Fields{"component": "syncer.repository_existence"}, "Message": "completed"},
		map[string]interface{}{"Data": loggerHook.LastEntry().Data, "Message": loggerHook.LastEntry().Message},
	)
}

func TestExistencer_Run_noAvailableStorages(t *testing.T) {
	t.Parallel()

	const (
		repo1RelPath = "repo-1.git"

		storage1 = "gitaly-1"
		storage2 = "gitaly-2"
		storage3 = "gitaly-3"

		virtualStorage = "praefect"
	)

	g1Cfg := testcfg.Build(t, testcfg.WithStorages(storage1))
	g2Cfg := testcfg.Build(t, testcfg.WithStorages(storage2))
	g3Cfg := testcfg.Build(t, testcfg.WithStorages(storage3))

	g1Addr := testserver.RunGitalyServer(t, g1Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())
	g2Addr := testserver.RunGitalyServer(t, g2Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())
	g3Addr := testserver.RunGitalyServer(t, g3Cfg, nil, setup.RegisterAll, testserver.WithDisablePraefect())

	db := glsql.NewDB(t)
	var database string
	require.NoError(t, db.QueryRow(`SELECT current_database()`).Scan(&database))
	dbConf := glsql.GetDBConfig(t, database)

	conf := config.Config{
		SocketPath: testhelper.GetTemporaryGitalySocketFileName(t),
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: virtualStorage,
				Nodes: []*config.Node{
					{Storage: g1Cfg.Storages[0].Name, Address: g1Addr},
					{Storage: g2Cfg.Storages[0].Name, Address: g2Addr},
					{Storage: g3Cfg.Storages[0].Name, Address: g3Addr},
				},
			},
		},
		DB: dbConf,
		Sync: config.Sync{
			RunInterval:         gconfig.Duration(time.Minute),
			LivenessInterval:    gconfig.Duration(time.Second),
			RepositoriesInBatch: 2,
		},
	}

	gittest.CloneRepo(t, g1Cfg, g1Cfg.Storages[0], gittest.CloneRepoOpts{RelativePath: repo1RelPath})

	ctx, cancel := testhelper.Context()
	defer cancel()

	repoStore := datastore.NewPostgresRepositoryStore(db.DB, nil)
	for i, set := range []struct {
		relativePath string
		primary      string
		secondaries  []string
	}{
		{
			relativePath: repo1RelPath,
			primary:      storage1,
			secondaries:  []string{storage3},
		},
	} {
		require.NoError(t, repoStore.CreateRepository(ctx, int64(i), conf.VirtualStorages[0].Name, set.relativePath, set.primary, set.secondaries, nil, false, false))
	}

	logger := testhelper.NewTestLogger(t)
	entry := logger.WithContext(ctx)
	clientHandshaker := backchannel.NewClientHandshaker(entry, praefect.NewBackchannelServerFactory(entry, transaction.NewServer(nil)))
	nodeSet, err := praefect.DialNodes(ctx, conf.VirtualStorages, protoregistry.GitalyProtoPreregistered, nil, clientHandshaker)
	require.NoError(t, err)
	defer nodeSet.Close()

	storageCleanup := datastore.NewStorageCleanup(db.DB)
	startSecond := make(chan struct{})
	releaseFirst := make(chan struct{})
	existencer := NewRepositoryExistence(conf, logger, praefect.StaticHealthChecker{virtualStorage: []string{storage1, storage2, storage3}}, nodeSet.Connections(), storageCleanup, storageCleanup, actionStub{
		PerformMethod: func(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error {
			assert.Equal(
				t,
				map[datastore.RepositoryClusterPath]bool{
					datastore.NewRepositoryClusterPath(virtualStorage, storage1, repo1RelPath): true,
				},
				existence,
			)
			// Block execution here until send instance completes it's execution.
			// It allows us to be sure the picked storage can't be picked once again by
			// another instance as well as that it works without problems if there is
			// nothing to pickup to process.
			close(startSecond)
			<-releaseFirst
			return nil
		},
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		logger, loggerHook := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		existencer := NewRepositoryExistence(conf, logger, praefect.StaticHealthChecker{virtualStorage: []string{storage1, storage2, storage3}}, nodeSet.Connections(), storageCleanup, storageCleanup, actionStub{
			PerformMethod: func(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error {
				assert.FailNow(t, "should not be triggered as there is no available storages to acquire")
				return nil
			},
		})

		ctx, cancel := testhelper.Context()
		defer cancel()
		ticker := helper.NewManualTicker()

		go func() {
			defer wg.Done()
			assert.Equal(t, context.Canceled, existencer.Run(ctx, ticker))
		}()
		<-startSecond
		ticker.Tick() // triggers execution
		ticker.Tick()
		ticker.Tick() // blocks until first run will be completed
		close(releaseFirst)
		cancel()
		wg.Wait()

		entries := loggerHook.AllEntries()
		require.Greater(t, len(entries), 2)
		require.Equal(
			t,
			map[string]interface{}{"Data": logrus.Fields{"component": "syncer.repository_existence"}, "Message": "no storages to verify"},
			map[string]interface{}{"Data": loggerHook.AllEntries()[1].Data, "Message": loggerHook.AllEntries()[1].Message},
		)
	}()

	ticker := helper.NewManualTicker()
	go func() {
		defer wg.Done()
		assert.Equal(t, context.Canceled, existencer.Run(ctx, ticker))
	}()
	ticker.Tick() // triggers execution
	ticker.Tick()
	ticker.Tick() // blocks until first run will be completed
	cancel()
	wg.Wait()
}

type actionStub struct {
	PerformMethod func(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error
}

func (as actionStub) Perform(ctx context.Context, existence map[datastore.RepositoryClusterPath]bool) error {
	if as.PerformMethod != nil {
		return as.PerformMethod(ctx, existence)
	}
	return nil
}
