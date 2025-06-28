package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
)

//go:embed migrations
var embeddedMigrations embed.FS

const createMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version int not null,
    hash blob not null
);`

func migrateDB(ctx context.Context, db *sql.DB) error {

	migs, err := loadMigrations(embeddedMigrations)
	if err != nil {
		return err
	}
	if err = migs.valid(); err != nil {
		return fmt.Errorf("invalid migrations: %w", err)
	}
	_, err = db.ExecContext(ctx, createMigrationsTable)
	if err != nil {
		return err
	}
	oldMigs, err := loadMigrationHistory(ctx, db)
	if err != nil {
		return err
	}
	for idx, mig := range migs {
		if idx >= len(oldMigs) {
			_, err = db.ExecContext(ctx, mig.content)
			if err != nil {
				return err
			}
			_, err = db.ExecContext(ctx, "INSERT INTO schema_migrations (version, hash) VALUES ($1, $2);", mig.version, mig.hash)
			if err != nil {
				return err
			}
		} else {
			other := oldMigs[idx]
			if other.version != mig.version || !bytes.Equal(other.hash, mig.hash) {
				return fmt.Errorf("migration mismatch between old %v, new %v", other, mig)
			}
		}
	}
	return nil
}

func loadMigrationHistory(ctx context.Context, db *sql.DB) (migrations, error) {

	rows, err := db.QueryContext(ctx, "SELECT version, hash FROM schema_migrations ORDER BY version;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var migs migrations
	for rows.Next() {
		var mig migration
		err = rows.Scan(&mig.version, &mig.hash)
		if err != nil {
			return nil, err
		}
		migs = append(migs, mig)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return migs, nil
}

type migrations []migration

func (m migrations) valid() error {

	if len(m) == 0 {
		return errors.New("no migrations")
	}
	versions := map[int]struct{}{}
	for _, mig := range m {
		if _, ok := versions[mig.version]; ok {
			return fmt.Errorf("version number clash on %d", mig.version)
		}
		versions[mig.version] = struct{}{}
	}
	return nil
}

type migration struct {
	version int
	content string
	hash    []byte
}

var migrationRe = regexp.MustCompile("(\\d+)_([^.]+)\\.sql")

func loadMigrations(dir embed.FS) (migrations, error) {

	var migs migrations
	err := fs.WalkDir(dir, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		submatch := migrationRe.FindStringSubmatch(filepath.Base(path))
		// text, group1, group2
		if len(submatch) != 3 {
			return nil
		}
		version, err := strconv.ParseInt(submatch[1], 10, 64)
		if err != nil {
			return err
		}
		content, err := dir.ReadFile(path)
		if err != nil {
			return err
		}
		hashArr := sha256.Sum256(content)
		migs = append(migs, migration{
			version: int(version),
			content: string(content),
			hash:    hashArr[:],
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(migs, func(a, b migration) int {
		return a.version - b.version
	})

	return migs, nil
}
