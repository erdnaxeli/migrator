package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"

	"github.com/erdnaxeli/migrator"
)

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

	directory := os.DirFS("migrations")
	migrator, err := migrator.New(db, directory)
	if err != nil {
		return nil, err
	}

	err = migrator.Migrate()
	if err != nil {
		return nil, err
	}

	return db, nil
}
