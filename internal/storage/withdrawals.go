package storage

import (
	"context"
	"time"

	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
)

func (s *Storage) Withdraw(ctx context.Context, w entity.Withdraw) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	balanceQuery := `
		UPDATE balances 
		SET current = current - $1,
			withdrawn = withdrawn + $1
		WHERE user_id = $2`

	_, err = tx.ExecContext(timeoutCtx, balanceQuery, w.Sum, w.UserID)
	if err != nil {
		return err
	}

	withdrawQuery := `
		INSERT INTO withdrawals 
		    (user_id, 
		     order_id, 
		     sum)
		VALUES ($1, $2, $3)`

	_, err = tx.ExecContext(timeoutCtx, withdrawQuery, w.UserID, w.OrderID, w.Sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetUserWithdrawals(ctx context.Context, userID int) ([]entity.Withdraw, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		SELECT order_id, 
		       sum, 
		       processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at ASC`

	rows, err := s.db.QueryContext(timeoutCtx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var withdrawals []entity.Withdraw
	for rows.Next() {
		var w entity.Withdraw

		err := rows.Scan(&w.OrderID, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
