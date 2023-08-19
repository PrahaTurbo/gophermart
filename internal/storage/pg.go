package storage

import (
	"context"
	"database/sql"

	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

const contextTimeoutSeconds = 3

const uniqueViolationErrCode = "23505"

type Repository interface {
	SaveUser(ctx context.Context, user entity.User) (int, error)
	GetUser(ctx context.Context, login string) (*entity.User, error)

	SaveOrder(ctx context.Context, order entity.Order) error
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]entity.Order, error)
	UpdateOrder(order entity.Order) error

	CreateBalance(ctx context.Context, userID int) error
	GetBalance(ctx context.Context, userID int) (*entity.Balance, error)
	UpdateBalance(amount int, userID int) error

	Withdraw(ctx context.Context, w entity.Withdraw) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]entity.Withdraw, error)
}

type Storage struct {
	db     *sql.DB
	logger logger.Logger
}

func NewStorage(db *sql.DB, logger logger.Logger) Repository {
	s := &Storage{
		db:     db,
		logger: logger,
	}

	return s
}

func SetupDB(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("dsn is an empty string")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	err = goose.SetDialect("pgx")
	if err != nil {
		return nil, err
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		return nil, err
	}

	return db, nil
}
