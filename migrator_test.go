package migrator_test

import (
	// register sqlite driver
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/erdnaxeli/migrator"
)

//go:embed test_data/no_migrations/empty
var noMigrationsFS embed.FS

//go:embed test_data/empty_migration/*.sql
var emptyMigrationFS embed.FS

//go:embed test_data/invalid_filename/*.sql
var invalidFilenameFS embed.FS

//go:embed test_data/duplicated_version/*.sql
var duplicatedVersionFS embed.FS

//go:embed test_data/missing_version/*.sql
var missingVersionFS embed.FS

//go:embed test_data/invalid_migration_1/*.sql
var invalidMigration1FS embed.FS

//go:embed test_data/invalid_migration_2/*.sql
var invalidMigration2FS embed.FS

//go:embed test_data/migrations_ok/*.sql
var migrationsOKFS embed.FS

func TestNew_EmptyState(t *testing.T) {
	// no migrations, no schema_migrations table

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	_, err = migrator.New(db, noMigrationsFS)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	rows, err := db.Query(`SELECT * FROM schema_migrations`)
	if err != nil {
		t.Fatalf("expected schema_migrations table to be created, got error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		t.Fatalf("expected schema_migrations table to be empty, but found a row")
	}
}

func TestNew_TableExists(t *testing.T) {
	// no migrations, schema_migrations table exists

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("failed to create schema_migrations table: %v", err)
	}

	_, err = migrator.New(db, noMigrationsFS)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	rows, err := db.Query(`SELECT * FROM schema_migrations`)
	if err != nil {
		t.Fatalf("expected schema_migrations table to exist, got error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		t.Fatalf("expected schema_migrations table to be empty, but found a row")
	}
}

func TestNew_EmptyMigration(t *testing.T) {
	// one migration, empty up SQL

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(emptyMigrationFS, "test_data/empty_migration")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected EmptyMigrationError, got no error: %v", err)
	}

	var emptyErr migrator.EmptyMigrationError
	if !errors.As(err, &emptyErr) {
		t.Fatalf("expected EmptyMigrationError, got: %v", err)
	}

	if emptyErr.Filename != "2_empty.sql" {
		t.Fatalf("expected filename '2_empty.sql', got: %s", emptyErr.Filename)
	}
}

func TestNew_InvalidFilename(t *testing.T) {
	// one migration, invalid filename

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(invalidFilenameFS, "test_data/invalid_filename")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected InvalidMigrationFilenameError, got no error: %v", err)
	}

	var invalidErr migrator.InvalidMigrationFilenameError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("expected InvalidMigrationFilenameError, got: %v", err)
	}

	if invalidErr.Filename != "invalid_2.sql" {
		t.Fatalf("expected invalid filename 'invalid_2.sql', got: %s", invalidErr.Filename)
	}
}

func TestNew_DuplicatedVersion(t *testing.T) {
	// two migrations, same version

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(duplicatedVersionFS, "test_data/duplicated_version")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected DuplicateMigrationVersionError, got no error: %v", err)
	}

	var dupErr migrator.DuplicateMigrationVersionError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected DuplicateMigrationVersionError, got: %v", err)
	}

	if dupErr.Version != 2 {
		t.Fatalf("expected duplicated version '2', got: %d", dupErr.Version)
	}
}

func TestNew_MissingVersion(t *testing.T) {
	// two migrations, missing version 2

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(missingVersionFS, "test_data/missing_version")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected MissingMigrationVersionError, got no error: %v", err)
	}

	var missErr migrator.MissingMigrationVersionError
	if !errors.As(err, &missErr) {
		t.Fatalf("expected MissingMigrationVersionError, got: %v", err)
	}

	if missErr.Version != 2 {
		t.Fatalf("expected missing version '2', got: %d", missErr.Version)
	}
}

func TestNew_InvalidMigration_1(t *testing.T) {
	// one migration, invalid format

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(invalidMigration1FS, "test_data/invalid_migration_1")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected InvalidMigrationFileError, got no error: %v", err)
	}

	var invalidErr migrator.InvalidMigrationFileError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("expected InvalidMigrationFileError, got: %v", err)
	}

	if invalidErr.Filename != "2_invalid.sql" {
		t.Fatalf("expected filename '2_invalid.sql', got: %s", invalidErr.Filename)
	}
}

func TestNew_InvalidMigration_2(t *testing.T) {
	// one migration, invalid format

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(invalidMigration2FS, "test_data/invalid_migration_2")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected InvalidMigrationFileError, got no error: %v", err)
	}

	var invalidErr migrator.InvalidMigrationFileError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("expected InvalidMigrationFileError, got: %v", err)
	}

	if invalidErr.Filename != "2_invalid.sql" {
		t.Fatalf("expected filename '2_invalid.sql', got: %s", invalidErr.Filename)
	}
}

func TestNew_InvalidCurrentVersion(t *testing.T) {
	// four migrations, current version is 5
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("failed to create schema_migrations table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO schema_migrations (version) VALUES (5)`)
	if err != nil {
		t.Fatalf("failed to insert current version: %v", err)
	}

	subFS, err := fs.Sub(migrationsOKFS, "test_data/migrations_ok")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	_, err = migrator.New(db, subFS)
	if err == nil {
		t.Fatalf("expected error due to invalid current version, got no error")
	}

	var invalidErr migrator.InvalidCurrentVersionError
	if !errors.As(err, &invalidErr) {
		t.Fatalf("expected InvalidCurrentVersionError, got: %v", err)
	}

	if invalidErr.Version != 5 {
		t.Fatalf("expected invalid current version '5', got: %d", invalidErr.Version)
	}
}
