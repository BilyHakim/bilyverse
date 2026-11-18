-- +goose Up
CREATE TABLE site_content (
    id boolean PRIMARY KEY DEFAULT true CHECK (id),
    revision integer NOT NULL DEFAULT 0,
    fields jsonb NOT NULL DEFAULT '{}'
);
INSERT INTO site_content DEFAULT VALUES;

-- +goose Down
DROP TABLE site_content;
