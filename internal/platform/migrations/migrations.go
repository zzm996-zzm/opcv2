package migrations

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Result struct {
	Applied        []int64
	CurrentVersion int64
}

type migration struct {
	version int64
	name    string
	sql     string
}

func Run(ctx context.Context, db DB, dir string) (Result, error) {
	return RunFS(ctx, db, os.DirFS(dir))
}

func RunFS(ctx context.Context, db DB, fsys fs.FS) (Result, error) {
	migrations, err := loadMigrations(fsys)
	if err != nil {
		return Result{}, err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("begin migration transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	current, dirty, err := currentVersion(ctx, tx)
	if err != nil {
		return Result{}, err
	}
	if dirty {
		return Result{}, fmt.Errorf("database schema migration is dirty at version %d", current)
	}

	result := Result{CurrentVersion: current}
	for _, migration := range migrations {
		if migration.version <= current {
			continue
		}
		if err := applyMigration(ctx, tx, migration); err != nil {
			return Result{}, err
		}
		result.Applied = append(result.Applied, migration.version)
		result.CurrentVersion = migration.version
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit migrations: %w", err)
	}
	committed = true
	return result, nil
}

func currentVersion(ctx context.Context, tx pgx.Tx) (int64, bool, error) {
	if _, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint not null primary key,
    dirty boolean not null
)`); err != nil {
		return 0, false, fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	var version int64
	var dirty bool
	err := tx.QueryRow(ctx, `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version, &dirty)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("read schema version: %w", err)
	}
	return version, dirty, nil
}

func applyMigration(ctx context.Context, tx pgx.Tx, migration migration) error {
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations`); err != nil {
		return fmt.Errorf("clear schema version for %s: %w", migration.name, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, dirty) VALUES ($1, TRUE)`, migration.version); err != nil {
		return fmt.Errorf("mark migration %s dirty: %w", migration.name, err)
	}
	if _, err := tx.Exec(ctx, migration.sql); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.name, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE schema_migrations SET dirty = FALSE WHERE version = $1`, migration.version); err != nil {
		return fmt.Errorf("mark migration %s clean: %w", migration.name, err)
	}
	return nil
}

func loadMigrations(fsys fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		version, err := migrationVersion(name)
		if err != nil {
			return nil, err
		}
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}
		migrations = append(migrations, migration{
			version: version,
			name:    name,
			sql:     strings.TrimSpace(string(content)),
		})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}

func migrationVersion(name string) (int64, error) {
	raw, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("migration %s is missing numeric prefix", name)
	}
	version, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("migration %s has invalid numeric prefix: %w", name, err)
	}
	return version, nil
}
