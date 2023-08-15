-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals (
    user_id INT,
    order_id VARCHAR(255),
    sum INT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE withdrawals;
-- +goose StatementEnd
