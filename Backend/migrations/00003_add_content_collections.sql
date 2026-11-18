-- +goose Up
ALTER TABLE site_content ADD COLUMN collections jsonb;

-- +goose Down
ALTER TABLE site_content DROP COLUMN collections;
