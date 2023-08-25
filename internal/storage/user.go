package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
)

var ErrAlreadyExist = errors.New("login already exist in database")

func (s *Storage) SaveUser(ctx context.Context, user entity.User) (int, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id`

	var userID int
	err := s.db.QueryRowContext(timeoutCtx, query, user.Login, user.PasswordHash).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == uniqueViolationErrCode {
				return 0, ErrAlreadyExist
			}
		}
		return 0, err
	}

	return userID, nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (*entity.User, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*contextTimeoutSeconds)
	defer cancel()

	query := `
		SELECT id, login, password
		FROM users
		WHERE login = $1`

	row := s.db.QueryRowContext(timeoutCtx, query, login)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash); err != nil {
		return nil, err
	}

	return &user, nil
}
