package migrator_test

import (
	"database/sql"
	"embed"
	"io/fs"
	"testing"
)

//go:embed test_data/migration_rollback/*.sql
var migrationRollbackRootFS embed.FS
var migrationRollbackFS = Must(fs.Sub(migrationRollbackRootFS, "test_data/migration_rollback"))

func TestMigrate_OK(t *testing.T) {
	// apply all migrations successfully
	t.Parallel()

	db, migrator := getDBAndMigrator(t, migrationsOKFS)
	defer db.Close()

	err := migrator.Migrate()
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	version, err := migrator.Version()
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if version != 4 {
		t.Fatalf("expected version 4, got: %d", version)
	}

	// check that test_table exists
	rows, err := db.Query(`SELECT id, name, description FROM test_table`)
	if err != nil {
		t.Fatalf("expected test_table to exist, got error: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("expected two rows in test_table, got none")
	}

	var id int
	var name string
	var description sql.NullString
	err = rows.Scan(&id, &name, &description)
	if err != nil {
		t.Fatalf("failed to scan first row: %v", err)
	}

	if id != 1 || name != "Test Name 1" || description.Valid {
		t.Fatalf(
			"unexpected first row values: id=%d, name=%s, description=%s",
			id,
			name,
			description.String,
		)
	}

	if !rows.Next() {
		t.Fatalf("expected two rows in test_table, got only one")
	}

	err = rows.Scan(&id, &name, &description)
	if err != nil {
		t.Fatalf("failed to scan second row: %v", err)
	}

	if id != 2 || name != "Test Name 2" || description.Valid {
		t.Fatalf(
			"unexpected second row values: id=%d, name=%s, description=%s",
			id,
			name,
			description.String,
		)
	}

	if rows.Next() {
		t.Fatalf("expected only two rows in test_table, got more")
	}

	// check that another_test_table exists
	_, err = db.Exec(`SELECT id, name FROM another_test_table`)
	if err != nil {
		t.Fatalf("expected another_test_table to exist, got error: %v", err)
	}
}

func TestMigrate_Rollback(t *testing.T) {
	// migration fails and rolls back
	t.Parallel()

	db, migrator := getDBAndMigrator(t, migrationRollbackFS)
	defer db.Close()

	err := migrator.Migrate()
	if err == nil {
		t.Fatalf("expected migration to fail, but it succeeded")
	}

	version, err := migrator.Version()
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if version != 2 {
		t.Fatalf("expected version 2 after rollback, got: %d", version)
	}

	// check that test_table exists
	_, err = db.Exec(`SELECT id, name, description FROM test_table`)
	if err != nil {
		t.Fatalf("expected test_table to exist, got error: %v", err)
	}

	// check that another_test_table does not exist
	_, err = db.Exec(`SELECT id, name FROM another_test_table`)
	if err == nil {
		t.Fatalf("expected another_test_table to not exist, but it does")
	}
}

func TestMigrate_FromState(t *testing.T) {
	// start from version 2, apply remaining migrations
	t.Parallel()

	db, migrator := getDBAndMigrator(
		t,
		migrationsOKFS,
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`,
		`INSERT INTO schema_migrations (version) VALUES (2)`,
		`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, description TEXT)`,
	)
	defer db.Close()

	err := migrator.Migrate()
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	version, err := migrator.Version()
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if version != 4 {
		t.Fatalf("expected version 4, got: %d", version)
	}

	// check that test_tabel exists
	rows, err := db.Query(`SELECT id, name, description FROM test_table`)
	if err != nil {
		t.Fatalf("expected test_table to exist, got error: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("expected two rows in test_table, got none")
	}

	var id int
	var name string
	var description sql.NullString
	err = rows.Scan(&id, &name, &description)
	if err != nil {
		t.Fatalf("failed to scan first row: %v", err)
	}

	if id != 1 || name != "Test Name 1" || description.Valid {
		t.Fatalf(
			"unexpected first row values: id=%d, name=%s, description=%s",
			id,
			name,
			description.String,
		)
	}

	if !rows.Next() {
		t.Fatalf("expected two rows in test_table, got only one")
	}

	err = rows.Scan(&id, &name, &description)
	if err != nil {
		t.Fatalf("failed to scan second row: %v", err)
	}

	if id != 2 || name != "Test Name 2" || description.Valid {
		t.Fatalf(
			"unexpected second row values: id=%d, name=%s, description=%s",
			id,
			name,
			description.String,
		)
	}

	if rows.Next() {
		t.Fatalf("expected only two rows in test_table, got more")
	}

	// check that another_test_table exists
	_, err = db.Exec(`SELECT id, name FROM another_test_table`)
	if err != nil {
		t.Fatalf("expected another_test_table to exist, got error: %v", err)
	}
}
