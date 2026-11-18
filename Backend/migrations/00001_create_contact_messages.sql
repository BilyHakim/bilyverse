-- +goose Up
CREATE TABLE contact_messages (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name varchar(100) NOT NULL CHECK (length(btrim(name)) > 0),
    email varchar(254) NOT NULL CHECK (length(btrim(email)) > 0),
    message varchar(3000) NOT NULL CHECK (length(btrim(message)) > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE contact_messages;
