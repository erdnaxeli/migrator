# Migrator

Migrator is a golang library to apply migrations on a database.

# Why another tools?

Existing tools provide a lot of features but are complex and come with many dependencies.
The goal of this tool is to be simple, and without any dependencies.

# Usage

Create a migration like this:
```sql
-- +migrate Up
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
```

```go
package main

import (
    "database.sql"
    "os"

    "github.com/erdnaxeli/migrator"
    _ "modernc.org/sqlite"
)

func main() {
    // Note: in this example errors are not checked, but you should do it
    db, _ := sql.Open("sqlite", "db.sqlite")
    defer db.Close()

    directory := os.DirFS("migrations")
    migrator, _ := migrator.New(db, directory)

    migrator.Migrate()
}
```

# Features

Current features:
* Apply up migrations.
* Multiple statements in a migration.
* Each migration is applied in his own transaction. If one migration fails, nothing is applied and it stops.
* Support any database compatible with `sql.DB`.
* Support any migrations source compatible with `fs.FS`

No implemented:
* A CLI.
* Down migrations.
* Code migrations.
