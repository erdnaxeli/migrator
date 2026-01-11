-- +migrate Up
CREATE TABLE another_test_table (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
