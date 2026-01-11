package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"

	_ "modernc.org/sqlite"

	"github.com/erdnaxeli/migrator"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	_, err := openDatabaseAndMigrate()
	if err != nil {
		log.Fatal(err)
	}
}

func openDatabaseAndMigrate() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	// The migration files are embedded in a folder named "migrations",
	// so we need to create a sub filesystem starting from there.
	// Else it will list the root folder and not find any .sql files.
	subFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	migrator, err := migrator.New(db, subFS)
	if err != nil {
		return nil, err
	}

	err = migrator.Migrate()
	if err != nil {
		return nil, err
	}

	return db, nil
}
