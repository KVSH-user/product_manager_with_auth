-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR NOT NULL,
    password_hashed VARCHAR NOT NULL,
    created_at DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY,
    category_name VARCHAR NOT NULL
);

INSERT INTO category (category_name)
VALUES ('No category');

CREATE TABLE IF NOT EXISTS good (
    id SERIAL PRIMARY KEY,
    good_name VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS good_category (
    good_id INT NOT NULL,
    category_id INT NOT NULL,
    PRIMARY KEY (good_id, category_id),
    FOREIGN KEY (good_id) REFERENCES good (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE
);

CREATE INDEX users_email_idx ON users (email);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS good_category;
DROP TABLE IF EXISTS good;
DROP TABLE IF EXISTS category;
-- +goose StatementEnd
