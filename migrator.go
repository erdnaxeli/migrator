// Package migrator implements database migration functionalities.
package migrator

import (
	"database/sql"
	"io/fs"
	"regexp"
)

// FilenameRgx is the regular expression to match migration filenames.
var FilenameRgx = regexp.MustCompile(`^(\d+)_(.*)\.sql$`)

// Migrator is the interface to manage database migrations.
type Migrator interface {
	// Migrate applies all pending database migrations.
	Migrate() error

	// Version returns the current version of the database schema.
	Version() (int, error)
}

type migrator struct {
	db *sql.DB

	migrations     []Migration
	currentVersion int
	lastVersion    int
}

// Migration represents a database migration.
type Migration struct {
	version int
	name    string
	upSQL   []string
}

// New creates a new Migrator instance.
//
// It loads migrations from the provided fs.FS and checks the current database version.
// If the schema_migrations table does not exist, it creates it.
//
// It can returns the following errors:
//   - InvalidMigrationFilenameError
//   - InvalidMigrationFileError
//   - EmptyMigrationError
//   - DuplicateMigrationVersionError
//   - MissingMigrationVersionError
func New(db *sql.DB, fs fs.FS) (Migrator, error) {
	migrations, err := loadMigrations(fs)
	if err != nil {
		return nil, err
	}

	lastVersion, err := validateMigrations(migrations)
	if err != nil {
		return nil, err
	}

	currentVersion, err := getCurrentDBVersion(db)
	if err != nil {
		return nil, err
	}

	if currentVersion > lastVersion {
		return nil, InvalidCurrentVersionError{Version: currentVersion}
	}

	return &migrator{
		db:             db,
		migrations:     migrations,
		lastVersion:    lastVersion,
		currentVersion: currentVersion,
	}, nil
}
