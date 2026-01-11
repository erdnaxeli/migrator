package migrator_test

import (
	"database/sql"
	"io/fs"
	"testing"

	"github.com/erdnaxeli/migrator"
)

func TestVersion(t *testing.T) {
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

	version, err := migrator.Version()
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if version != 0 {
		t.Fatalf("expected initial version 0, got: %d", version)
	}

	err = migrator.Migrate()
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	version, err = migrator.Version()
	if err != nil {
		t.Fatalf("failed to get current version: %v", err)
	}

	if version != 4 {
		t.Fatalf("expected version 3 after migrations, got: %d", version)
	}
}
