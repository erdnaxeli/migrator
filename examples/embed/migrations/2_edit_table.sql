-- +migrate Up
ALTER TABLE test_table ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX idx_test_table_name ON test_table(name);
