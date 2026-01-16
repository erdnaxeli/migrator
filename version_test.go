package migrator_test

import (
	"testing"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	db, migrator := getDBAndMigrator(t, migrationsOKFS)
	defer db.Close()

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
