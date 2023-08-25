-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    user_id INT,
    order_id VARCHAR(255) UNIQUE,
    accrual INT,
    status VARCHAR(20),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
