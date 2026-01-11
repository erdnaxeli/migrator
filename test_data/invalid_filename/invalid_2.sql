-- +migrate Up
CREATE TABLE some_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
