package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaMigrationLockID int64 = 2026072201

type migration struct {
	version int64
	name    string
	sql     string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationFS fs.FS, logger *slog.Logger) error {
	migrations, err := readMigrations(migrationFS)
	if err != nil {
		return err
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin database migrations: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", schemaMigrationLockID); err != nil {
		return fmt.Errorf("lock database migrations: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	appliedVersions, err := findAppliedVersions(ctx, tx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if appliedVersions[migration.version] {
			continue
		}

		if logger != nil {
			logger.Info("applying database migration", "migration", migration.name)
		}

		if _, err := tx.Exec(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.name, err)
		}

		if _, err := tx.Exec(
			ctx,
			"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			migration.version,
			migration.name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", migration.name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit database migrations: %w", err)
	}

	return nil
}

func readMigrations(migrationFS fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version, ok := parseMigrationVersion(entry.Name())
		if !ok {
			continue
		}

		content, err := fs.ReadFile(migrationFS, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration{
			version: version,
			name:    entry.Name(),
			sql:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].version == migrations[j].version {
			return migrations[i].name < migrations[j].name
		}

		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func parseMigrationVersion(name string) (int64, bool) {
	versionPart, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, false
	}

	version, err := strconv.ParseInt(versionPart, 10, 64)
	if err != nil {
		return 0, false
	}

	return version, true
}

func findAppliedVersions(ctx context.Context, tx pgx.Tx) (map[int64]bool, error) {
	rows, err := tx.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("find applied migrations: %w", err)
	}
	defer rows.Close()

	appliedVersions := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}

		appliedVersions[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}

	return appliedVersions, nil
}
