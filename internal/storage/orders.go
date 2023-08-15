package storage

import (
	"context"
	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
	"time"
)

func (s *Storage) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		SELECT order_id, 
		       user_id, 
		       accrual, 
		       status
		FROM orders
		WHERE order_id = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, orderID)

	var order entity.Order
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Accrual,
		&order.Status)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *Storage) SaveOrder(ctx context.Context, order entity.Order) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		INSERT INTO orders 
		    (order_id, 
		     user_id, 
		     accrual, 
		     status)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.ExecContext(timeoutCtx, query, order.ID, order.UserID, order.Accrual, order.Status)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateOrder(order entity.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		UPDATE orders 
		SET accrual = $1,
		    status = $2
		WHERE order_id = $3`

	_, err := s.db.ExecContext(ctx, query, order.Accrual, order.Status, order.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		SELECT order_id, 
		       accrual, 
		       status, 
		       uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at ASC`

	rows, err := s.db.QueryContext(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order

		err := rows.Scan(
			&order.ID,
			&order.Accrual,
			&order.Status,
			&order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
