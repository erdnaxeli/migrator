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

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

//go:embed test_data/no_migrations/empty
var noMigrationsRootFS embed.FS
var noMigrationsFS = Must(fs.Sub(noMigrationsRootFS, "test_data/no_migrations"))

//go:embed test_data/empty_migration/*.sql
var emptyMigrationRootFS embed.FS
var emptyMigrationFS = Must(fs.Sub(emptyMigrationRootFS, "test_data/empty_migration"))

//go:embed test_data/invalid_filename/*.sql
var invalidFilenameRootFS embed.FS
var invalidFilenameFS = Must(fs.Sub(invalidFilenameRootFS, "test_data/invalid_filename"))

//go:embed test_data/duplicated_version/*.sql
var duplicatedVersionRootFS embed.FS
var duplicatedVersionFS = Must(fs.Sub(duplicatedVersionRootFS, "test_data/duplicated_version"))

//go:embed test_data/missing_version/*.sql
var missingVersionRootFS embed.FS
var missingVersionFS = Must(fs.Sub(missingVersionRootFS, "test_data/missing_version"))

//go:embed test_data/invalid_migration_1/*.sql
var invalidMigration1RootFS embed.FS
var invalidMigration1FS = Must(fs.Sub(invalidMigration1RootFS, "test_data/invalid_migration_1"))

//go:embed test_data/invalid_migration_2/*.sql
var invalidMigration2RootFS embed.FS
var invalidMigration2FS = Must(fs.Sub(invalidMigration2RootFS, "test_data/invalid_migration_2"))

//go:embed test_data/migrations_ok/*.sql
var migrationsOKRootFS embed.FS
var migrationsOKFS = Must(fs.Sub(migrationsOKRootFS, "test_data/migrations_ok"))

func getDB(t *testing.T, stmts ...string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			t.Fatalf("failed to execute statement '%s': %v", stmt, err)
		}
	}

	return db
}

func getMigrator(t *testing.T, db *sql.DB, migrations fs.FS) migrator.Migrator {
	t.Helper()

	migrator, err := migrator.New(db, migrations)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	return migrator
}

func getDBAndMigrator(
	t *testing.T,
	migrations fs.FS,
	stmts ...string,
) (*sql.DB, migrator.Migrator) {
	t.Helper()

	db := getDB(t, stmts...)
	migrator := getMigrator(t, db, migrations)

	return db, migrator
}

func TestNew_EmptyState(t *testing.T) {
	// no migrations, no schema_migrations table
	t.Parallel()

	db, _ := getDBAndMigrator(t, noMigrationsFS)
	defer db.Close()

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
	t.Parallel()

	db, _ := getDBAndMigrator(
		t,
		noMigrationsFS,
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`,
	)
	defer db.Close()

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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, emptyMigrationFS)
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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, invalidFilenameFS)
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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, duplicatedVersionFS)
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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, missingVersionFS)
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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, invalidMigration1FS)
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
	t.Parallel()

	db := getDB(t)
	defer db.Close()

	_, err := migrator.New(db, invalidMigration2FS)
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
	t.Parallel()

	db := getDB(
		t,
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`,
		`INSERT INTO schema_migrations (version) VALUES (5)`,
	)
	defer db.Close()

	_, err := migrator.New(db, migrationsOKFS)
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
