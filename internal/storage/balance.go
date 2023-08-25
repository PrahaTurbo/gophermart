package storage

import (
	"context"
	"time"

	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
)

func (s *Storage) CreateBalance(ctx context.Context, userID int) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		INSERT INTO balances 
		    (user_id)
		VALUES ($1)`

	_, err := s.db.ExecContext(timeoutCtx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID int) (*entity.Balance, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		SELECT current, withdrawn
		FROM balances
		WHERE user_id = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, userID)

	var balance entity.Balance
	err := row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (s *Storage) UpdateBalance(amount int, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		UPDATE balances 
		SET current = current + $1
		WHERE user_id = $2`

	_, err := s.db.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return err
	}

	return nil
}
