-- +migrate Down
ALTER TABLE test_table ADD COLUMN description TEXT;
