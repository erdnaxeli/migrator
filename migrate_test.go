package migrator_test

import (
	"database/sql"
	"embed"
	"io/fs"
	"testing"

	"github.com/erdnaxeli/migrator"
)

//go:embed test_data/migration_rollback/*.sql
var migrationRollbackFS embed.FS

func TestMigrate_OK(t *testing.T) {
	// apply all migrations successfully

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(migrationsOKFS, "test_data/migrations_ok")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	migrator, err := migrator.New(db, subFS)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	err = migrator.Migrate()
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

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	subFS, err := fs.Sub(migrationRollbackFS, "test_data/migration_rollback")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	migrator, err := migrator.New(db, subFS)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	err = migrator.Migrate()
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

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("failed to create schema_migrations table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO schema_migrations (version) VALUES (2)`)
	if err != nil {
		t.Fatalf("failed to insert current version: %v", err)
	}

	_, err = db.Exec(
		`CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT, description TEXT)`,
	)
	if err != nil {
		t.Fatalf("failed to create test_table: %v", err)
	}

	subFS, err := fs.Sub(migrationsOKFS, "test_data/migrations_ok")
	if err != nil {
		t.Fatalf("failed to get sub filesystem: %v", err)
	}

	migrator, err := migrator.New(db, subFS)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}

	err = migrator.Migrate()
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
