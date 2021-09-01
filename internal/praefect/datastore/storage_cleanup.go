package datastore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"gitlab.com/gitlab-org/gitaly/v14/internal/helper"
)

// RepositoryClusterPath identifies location of the repository in the cluster.
type RepositoryClusterPath struct {
	ClusterPath
	// RelativePath relative path to the repository on the disk.
	RelativePath string
}

// NewRepositoryClusterPath initializes and returns RepositoryClusterPath.
func NewRepositoryClusterPath(virtualStorage, storage, relativePath string) RepositoryClusterPath {
	return RepositoryClusterPath{
		ClusterPath: ClusterPath{
			VirtualStorage: virtualStorage,
			Storage:        storage,
		},
		RelativePath: relativePath,
	}
}

// ClusterPath represents path on the cluster to the storage.
type ClusterPath struct {
	// VirtualStorage is the name of the virtual storage.
	VirtualStorage string
	// Storage is the name of the gitaly storage.
	Storage string
}

// NewStorageCleanup initialises and returns a new instance of the StorageCleanup.
func NewStorageCleanup(db *sql.DB) *StorageCleanup {
	return &StorageCleanup{db: db}
}

// StorageCleanup provides methods on the database for the repository cleanup operation.
type StorageCleanup struct {
	db *sql.DB
}

// Populate adds storage to the set, so it can be acquired afterwards.
func (ss *StorageCleanup) Populate(ctx context.Context, virtualStorage, storage string) error {
	if _, err := ss.db.ExecContext(
		ctx,
		`
		INSERT INTO storage_cleanups (virtual_storage, storage) VALUES ($1, $2)
		ON CONFLICT (virtual_storage, storage) DO NOTHING`,
		virtualStorage, storage,
	); err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

// AcquireNextStorage picks up the next storage for the processing.
// Once acquired no other call to the same method will return the same storage, so it
// works as exclusive lock on that entry.
// Once processing is done the returned function needs to be called to release
// acquired storage. It updates last_run column of the entry on execution.
func (ss *StorageCleanup) AcquireNextStorage(ctx context.Context, inactive, updatePeriod time.Duration) (*ClusterPath, func() error, error) {
	var entry ClusterPath
	if err := ss.db.QueryRowContext(
		ctx,
		`UPDATE storage_cleanups
			SET triggered_at = (NOW() AT TIME ZONE 'UTC')
			WHERE (virtual_storage, storage) IN (
				SELECT virtual_storage, storage
				FROM storage_cleanups
				WHERE
					COALESCE(NOW() AT TIME ZONE 'UTC' - last_run, INTERVAL '1 MILLISECOND' * $1) >= INTERVAL '1 MILLISECOND' * $1
					AND COALESCE (NOW() AT TIME ZONE 'UTC' - triggered_at, INTERVAL '1 MILLISECOND' * $2) >= INTERVAL '1 MILLISECOND' * $2
				ORDER BY last_run NULLS FIRST, virtual_storage, storage
				LIMIT 1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING virtual_storage, storage`,
		inactive.Milliseconds(), updatePeriod.Milliseconds(),
	).Scan(&entry.VirtualStorage, &entry.Storage); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("scan: %w", err)
		}
		return nil, func() error { return nil }, nil
	}

	stop := make(chan struct{})
	go func(stop <-chan struct{}) {
		trigger := helper.NewTimerTicker(updatePeriod - 100*time.Millisecond)
		defer trigger.Stop()
		trigger.Reset()

		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-trigger.C():
				if _, err := ss.db.ExecContext(
					ctx,
					`UPDATE storage_cleanups
					SET triggered_at = (NOW() AT TIME ZONE 'UTC')
					WHERE virtual_storage = $1 AND storage = $2`,
					entry.VirtualStorage, entry.Storage,
				); err != nil {
					return
				}
				trigger.Reset()
			}
		}
	}(stop)

	return &entry, func() error {
		close(stop)

		if _, err := ss.db.ExecContext(
			ctx,
			`UPDATE storage_cleanups
			SET last_run = NOW(), triggered_at = NULL
			WHERE virtual_storage = $1 AND storage = $2`,
			entry.VirtualStorage, entry.Storage,
		); err != nil {
			return fmt.Errorf("update storage_cleanups: %w", err)
		}
		return nil
	}, nil
}

// Exists checks if each repository exists in the database by querying repositories and
// storage_repositories tables. The returned map contains a existence flag for each repository.
func (ss *StorageCleanup) Exists(ctx context.Context, paths []RepositoryClusterPath) (map[RepositoryClusterPath]bool, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	params := make([]interface{}, 0, len(paths))
	placeholders := bytes.NewBufferString("(")
	res := make(map[RepositoryClusterPath]bool, len(paths))
	for i, path := range paths {
		res[path] = false
		params = append(params, pq.StringArray{path.VirtualStorage, path.RelativePath, path.Storage})
		if i > 0 {
			placeholders.WriteString(",")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
	}
	placeholders.WriteString(")")
	rows, err := ss.db.QueryContext(
		ctx,
		`SELECT virtual_storage, relative_path, storage
		FROM repositories r
		JOIN storage_repositories sr USING (virtual_storage, relative_path)
		WHERE ARRAY[virtual_storage, relative_path, storage] IN `+placeholders.String(),
		params...,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for i := 0; rows.Next(); i++ {
		var curr RepositoryClusterPath
		if err := rows.Scan(&curr.VirtualStorage, &curr.RelativePath, &curr.Storage); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res[curr] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loop: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}
	return res, nil
}
