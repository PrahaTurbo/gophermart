package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/PrahaTurbo/gophermart/internal/client"
	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/models"
	"github.com/PrahaTurbo/gophermart/internal/storage"
	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateUser(ctx context.Context, userReq models.UserRequest) (int, error)
	LoginUser(ctx context.Context, userReq models.UserRequest) (int, error)

	ProcessOrder(ctx context.Context, orderID string) error
	GetUserOrders(ctx context.Context) ([]models.OrderResponse, error)

	GetBalance(ctx context.Context) (*models.BalanceResponse, error)

	Withdraw(ctx context.Context, req models.WithdrawRequest) error
	GetUserWithdrawals(ctx context.Context) ([]models.WithdrawalsResponse, error)
}

type service struct {
	log                logger.Logger
	storage            storage.Repository
	accrualClient      *client.AccrualClient
	accrualUpdaterChan chan entity.Order
}

func NewService(
	storage storage.Repository,
	accrualClient *client.AccrualClient,
	logger logger.Logger,
) Service {
	s := service{
		log:                logger,
		storage:            storage,
		accrualClient:      accrualClient,
		accrualUpdaterChan: make(chan entity.Order, 20),
	}

	s.log.Info().Msg("starting accrual updater")
	go s.startAccrualUpdater(time.Second*3, 50)

	return &s
}

func (s *service) CreateUser(ctx context.Context, userReq models.UserRequest) (int, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error().Err(err).Str("login", userReq.Login).Msg("failed to create hash from password")
		return 0, err
	}
	user := entity.User{
		Login:        userReq.Login,
		PasswordHash: string(passHash),
	}

	userID, err := s.storage.SaveUser(ctx, user)
	if err != nil {
		s.log.Error().Err(err).Str("login", user.Login).Msg("failed to save user in database")
		return 0, err
	}

	if err := s.storage.CreateBalance(ctx, userID); err != nil {
		s.log.Error().Err(err).Str("login", user.Login).Msg("failed to create balance for user")
		return 0, err
	}

	s.log.Info().Int("user", userID).Msg("user was created")
	return userID, nil
}

func (s *service) LoginUser(ctx context.Context, userReq models.UserRequest) (int, error) {
	savedUser, err := s.storage.GetUser(ctx, userReq.Login)
	if err != nil {
		s.log.Error().Err(err).Str("login", userReq.Login).Msg("cannot find user in database")
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte(userReq.Password)); err != nil {
		s.log.Error().Err(err).Str("login", userReq.Login).Msg("hash and password mismatch")
		return 0, err
	}

	s.log.Info().Int("user", savedUser.ID).Msg("user logged in")
	return savedUser.ID, nil
}

func (s *service) ProcessOrder(ctx context.Context, orderID string) error {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to extracrt user from context")
		return err
	}

	isValidLuhn := validateLuhn(orderID)
	if !isValidLuhn {
		s.log.Error().Err(err).Str("order", orderID).Send()
		return ErrInvalidOrderID
	}

	order, err := s.storage.GetOrder(ctx, orderID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.log.Error().Err(err).Str("order", orderID).Msg("failed to get order for provided order id")
			return err
		}
	}

	if order != nil {
		if order.UserID != userID {
			s.log.Info().Str("order", orderID).Int("user", userID).Msg(ErrOrderByAnotherUser.Error())
			return ErrOrderByAnotherUser
		}

		s.log.Info().Str("order", orderID).Int("user", userID).Msg(ErrOrderByCurrentUser.Error())
		return ErrOrderByCurrentUser
	}

	order = &entity.Order{
		ID:     orderID,
		UserID: userID,
		Status: OrderNew,
	}

	if saveErr := s.storage.SaveOrder(ctx, *order); saveErr != nil {
		s.log.Error().Err(err).Any("order", order).Msg("failed to save order")
		return saveErr
	}

	s.accrualUpdaterChan <- *order

	s.log.Info().Any("order", order).Msg("order was placed in update channel")

	return nil
}

func (s *service) GetUserOrders(ctx context.Context) ([]models.OrderResponse, error) {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to extracrt user from context")
		return nil, err
	}

	orders, err := s.storage.GetUserOrders(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Int("user", userID).Msg("failed to get user's orders")
		return nil, err
	}

	resp := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		o := models.OrderResponse{
			ID:         order.ID,
			Status:     order.Status,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}

		if order.Status == OrderProcessed {
			o.Accrual = amountToFloat64(order.Accrual)
		}

		resp[i] = o
	}

	return resp, nil
}

func (s *service) GetBalance(ctx context.Context) (*models.BalanceResponse, error) {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to extracrt user from context")
		return nil, err
	}

	balance, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Int("user", userID).Msg("failed to get user's balance")
		return nil, err
	}

	resp := &models.BalanceResponse{
		Current:   amountToFloat64(balance.Current),
		Withdrawn: amountToFloat64(balance.Withdrawn),
	}

	return resp, nil
}

func (s *service) Withdraw(ctx context.Context, req models.WithdrawRequest) error {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to extracrt user from context")
		return err
	}

	isValidLuhn := validateLuhn(req.Order)
	if !isValidLuhn {
		s.log.Error().Err(err).Str("order", req.Order).Send()
		return ErrInvalidOrderID
	}

	withdraw := entity.Withdraw{
		UserID:  userID,
		OrderID: req.Order,
		Sum:     amountToInt(req.Sum),
	}

	balance, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Int("user", userID).Msg("failed to get user's balance")
		return err
	}

	if balance.Current < withdraw.Sum {
		s.log.Error().Err(err).Int("user", userID).Send()
		return ErrBalanceNotEnough
	}

	if err := s.storage.Withdraw(ctx, withdraw); err != nil {
		s.log.Error().Err(err).Int("user", userID).Msg("failed to withdraw from balance")
		return err
	}

	s.log.Info().Int("user", userID).Int("sum", withdraw.Sum).Msg("funds were withdrawn from user's balance")

	return nil
}

func (s *service) GetUserWithdrawals(ctx context.Context) ([]models.WithdrawalsResponse, error) {
	userID, err := extractUserIDFromCtx(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to extracrt user from context")
		return nil, err
	}

	withdrawals, err := s.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).Int("user", userID).Msg("failed to get user's withdrawals")
		return nil, err
	}

	resp := make([]models.WithdrawalsResponse, len(withdrawals))
	for i, withdraw := range withdrawals {
		w := models.WithdrawalsResponse{
			Order:       withdraw.OrderID,
			Sum:         amountToFloat64(withdraw.Sum),
			ProcessedAt: withdraw.ProcessedAt.Format(time.RFC3339),
		}

		resp[i] = w
	}

	return resp, nil
}

func (s *service) startAccrualUpdater(interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)

	var orders []entity.Order

	for {
		select {
		case order := <-s.accrualUpdaterChan:
			orders = append(orders, order)

			if len(orders) >= batchSize {
				s.handleAccrualUpdater(orders)
				orders = nil
			}
		case <-ticker.C:
			if len(orders) > 0 {
				s.handleAccrualUpdater(orders)
				orders = nil
			}
		}
	}
}

func (s *service) handleAccrualUpdater(orders []entity.Order) {
	const maxRetryCount = 3

	for _, order := range orders {
		go func(order entity.Order) {
			for retryCount := 0; retryCount < maxRetryCount; retryCount++ {
				resp, err := s.accrualClient.GetAccrual(order.ID)
				if err != nil {
					s.log.Error().Err(err).Str("order_id", order.ID).Int("attempt", retryCount+1).Msg("failed to get accrual information")
					continue
				}

				order.Accrual = amountToInt(resp.Accrual)
				order.Status = resp.Status

				if err := s.storage.UpdateOrder(order); err != nil {
					s.log.Error().Err(err).Str("order_id", order.ID).Msg("failed to update order")
				}

				switch order.Status {
				case OrderProcessed:
					if err := s.storage.UpdateBalance(order.Accrual, order.UserID); err != nil {
						s.log.Error().Err(err).Str("order_id", order.ID).Msg("failed to update balance")
					}
				case OrderRegistered, OrderProcessing:
					s.accrualUpdaterChan <- order
				}

				break
			}
		}(order)
	}
}
