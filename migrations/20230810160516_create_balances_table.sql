-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS balances (
    user_id INT,
    current INT DEFAULT 0,
    withdrawn INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE balances;
-- +goose StatementEnd
