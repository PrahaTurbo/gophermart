package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/PrahaTurbo/gophermart/internal/auth"
	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/mocks"
	"github.com/PrahaTurbo/gophermart/internal/models"
	"github.com/PrahaTurbo/gophermart/internal/storage"
	"github.com/PrahaTurbo/gophermart/internal/storage/entity"
)

type badContextKey string

var errInternal = errors.New("internal error")

func Test_service_CreateUser(t *testing.T) {
	service := service{
		log: logger.NewLogger(),
	}

	type want struct {
		userID int
		err    error
	}

	tests := []struct {
		name    string
		userReq models.UserRequest
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully return user id",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password",
			},
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						SaveUser(gomock.Any(), gomock.Any()).
						Return(1, nil),
					s.EXPECT().
						CreateBalance(gomock.Any(), 1).
						Return(nil),
				)
			},
			want: want{
				userID: 1,
				err:    nil,
			},
		},
		{
			name: "should return error that user already exist",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password",
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					SaveUser(gomock.Any(), gomock.Any()).
					Return(0, storage.ErrAlreadyExist)
			},
			want: want{
				userID: 0,
				err:    storage.ErrAlreadyExist,
			},
		},
		{
			name: "should return error if failed to create balance",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password",
			},
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						SaveUser(gomock.Any(), gomock.Any()).
						Return(1, nil),
					s.EXPECT().
						CreateBalance(gomock.Any(), 1).
						Return(errInternal),
				)
			},
			want: want{
				userID: 0,
				err:    errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			result, err := service.CreateUser(context.Background(), tt.userReq)
			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.userID, result)
		})
	}
}

func Test_service_LoginUser(t *testing.T) {
	service := service{
		log: logger.NewLogger(),
	}

	type want struct {
		userID int
		err    error
	}

	tests := []struct {
		name    string
		userReq models.UserRequest
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully return user id",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password",
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUser(gomock.Any(), "test").
					Return(&entity.User{
						ID:           1,
						Login:        "test",
						PasswordHash: genHashString("test_password"),
					}, nil)
			},
			want: want{
				userID: 1,
				err:    nil,
			},
		},
		{
			name: "should return error if password and has mismatch",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password_2",
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUser(gomock.Any(), "test").
					Return(&entity.User{
						ID:           1,
						Login:        "test",
						PasswordHash: genHashString("test_password"),
					}, nil)
			},
			want: want{
				userID: 0,
				err:    bcrypt.ErrMismatchedHashAndPassword,
			},
		},
		{
			name: "should return error if can't find user",
			userReq: models.UserRequest{
				Login:    "test",
				Password: "test_password_2",
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUser(gomock.Any(), "test").
					Return(nil, errInternal)
			},
			want: want{
				userID: 0,
				err:    errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			result, err := service.LoginUser(context.Background(), tt.userReq)
			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.userID, result)
		})
	}
}

func Test_service_ProcessOrder(t *testing.T) {
	service := service{
		log:                logger.NewLogger(),
		accrualUpdaterChan: make(chan entity.Order, 20),
	}

	type want struct {
		orderID string
		err     error
	}

	tests := []struct {
		name    string
		orderID string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name:    "should successfully accept an order",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						GetOrder(gomock.Any(), "12345678903").
						Return(nil, sql.ErrNoRows),
					s.EXPECT().
						SaveOrder(gomock.Any(), entity.Order{
							ID:     "12345678903",
							UserID: 1,
							Status: OrderNew,
						}).
						Return(nil),
				)
			},
			want: want{
				orderID: "12345678903",
				err:     nil,
			},
		},
		{
			name:    "should return error if order id doesn't match luhn algorithm",
			orderID: "1234567890",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrInvalidOrderID,
			},
		},
		{
			name:    "should return error if can't get order",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetOrder(gomock.Any(), "12345678903").
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name:    "should return error if order was added by another user",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetOrder(gomock.Any(), "12345678903").
					Return(&entity.Order{
						ID:     "12345678903",
						UserID: 2,
					}, nil)
			},
			want: want{
				err: ErrOrderByAnotherUser,
			},
		},
		{
			name:    "should return error if order was added by current user",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetOrder(gomock.Any(), "12345678903").
					Return(&entity.Order{
						ID:     "12345678903",
						UserID: 1,
					}, nil)
			},
			want: want{
				err: ErrOrderByCurrentUser,
			},
		},
		{
			name:    "should return error if can't save order",
			orderID: "12345678903",
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						GetOrder(gomock.Any(), "12345678903").
						Return(nil, sql.ErrNoRows),
					s.EXPECT().
						SaveOrder(gomock.Any(), entity.Order{
							ID:     "12345678903",
							UserID: 1,
							Status: OrderNew,
						}).
						Return(errInternal),
				)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)

			if tt.want.err == ErrExtractFromContext {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			err := service.ProcessOrder(ctx, tt.orderID)

			if err == nil {
				res := <-service.accrualUpdaterChan
				assert.Equal(t, tt.want.orderID, res.ID)
			}

			assert.Equal(t, tt.want.err, err)
		})
	}
}

func Test_service_GetUserOrders(t *testing.T) {
	now := time.Now()

	service := service{
		log: logger.NewLogger(),
	}

	type want struct {
		ordersResp []models.OrderResponse
		err        error
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully return user orders",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUserOrders(gomock.Any(), 1).
					Return([]entity.Order{
						{
							ID:         "12345678903",
							Status:     OrderProcessed,
							UploadedAt: now,
							Accrual:    13400,
						},
					}, nil)
			},
			want: want{
				ordersResp: []models.OrderResponse{
					{
						ID:         "12345678903",
						Accrual:    134,
						Status:     OrderProcessed,
						UploadedAt: now.Format(time.RFC3339),
					},
				},
				err: nil,
			},
		},
		{
			name: "shouldn't return accrual if status not processed",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUserOrders(gomock.Any(), 1).
					Return([]entity.Order{
						{
							ID:         "12345678903",
							Status:     OrderNew,
							UploadedAt: now,
							Accrual:    13400,
						},
					}, nil)
			},
			want: want{
				ordersResp: []models.OrderResponse{
					{
						ID:         "12345678903",
						Status:     OrderNew,
						UploadedAt: now.Format(time.RFC3339),
					},
				},
				err: nil,
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if can't get user orders",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUserOrders(gomock.Any(), 1).
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)

			if tt.want.err == ErrExtractFromContext {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			result, err := service.GetUserOrders(ctx)

			if err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.ordersResp, result)
		})
	}
}

func Test_service_GetBalance(t *testing.T) {
	service := service{
		log: logger.NewLogger(),
	}

	type want struct {
		balanceResp *models.BalanceResponse
		err         error
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully return balance response",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetBalance(gomock.Any(), 1).
					Return(&entity.Balance{
						Current:   13400,
						Withdrawn: 1300,
					}, nil)
			},
			want: want{
				balanceResp: &models.BalanceResponse{
					Current:   134,
					Withdrawn: 13,
				},
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if can't get user balance",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetBalance(gomock.Any(), 1).
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)

			if tt.want.err == ErrExtractFromContext {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			result, err := service.GetBalance(ctx)

			if err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.balanceResp, result)
		})
	}
}

func Test_service_Withdraw(t *testing.T) {
	service := service{
		log:                logger.NewLogger(),
		accrualUpdaterChan: make(chan entity.Order, 20),
	}

	type want struct {
		err error
	}

	tests := []struct {
		name    string
		req     models.WithdrawRequest
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully accept an order",
			req: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   13,
			},
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						GetBalance(gomock.Any(), 1).
						Return(&entity.Balance{
							Current:   13400,
							Withdrawn: 1300,
						}, nil),
					s.EXPECT().
						Withdraw(gomock.Any(), entity.Withdraw{
							UserID:  1,
							OrderID: "12345678903",
							Sum:     1300,
						}).
						Return(nil),
				)
			},
		},
		{
			name: "should return error if order id doesn't match luhn algorithm",
			req: models.WithdrawRequest{
				Order: "123",
				Sum:   13,
			},
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrInvalidOrderID,
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if balance lower than withdraw sum",
			req: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   130,
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetBalance(gomock.Any(), 1).
					Return(&entity.Balance{
						Current:   1000,
						Withdrawn: 1300,
					}, nil)
			},
			want: want{
				err: ErrBalanceNotEnough,
			},
		},
		{
			name: "should return error if can't get user balance",
			req: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   130,
			},
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetBalance(gomock.Any(), 1).
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
		{
			name: "should return error if can't withdraw from balance",
			req: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   13,
			},
			prepare: func(s *mocks.MockRepository) {
				gomock.InOrder(
					s.EXPECT().
						GetBalance(gomock.Any(), 1).
						Return(&entity.Balance{
							Current:   13400,
							Withdrawn: 1300,
						}, nil),
					s.EXPECT().
						Withdraw(gomock.Any(), entity.Withdraw{
							UserID:  1,
							OrderID: "12345678903",
							Sum:     1300,
						}).
						Return(errInternal),
				)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)

			if tt.want.err == ErrExtractFromContext {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			err := service.Withdraw(ctx, tt.req)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func Test_service_GetUserWithdrawals(t *testing.T) {
	now := time.Now()

	service := service{
		log: logger.NewLogger(),
	}

	type want struct {
		withdrawals []models.WithdrawalsResponse
		err         error
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockRepository)
		want    want
	}{
		{
			name: "should successfully return user withdrawals",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUserWithdrawals(gomock.Any(), 1).
					Return([]entity.Withdraw{
						{
							OrderID:     "12345678903",
							Sum:         1000,
							ProcessedAt: now,
						},
					}, nil)
			},
			want: want{
				withdrawals: []models.WithdrawalsResponse{
					{
						Order:       "12345678903",
						Sum:         10,
						ProcessedAt: now.Format(time.RFC3339),
					},
				},
				err: nil,
			},
		},
		{
			name:    "should return error if can't extract user id from context",
			prepare: func(s *mocks.MockRepository) {},
			want: want{
				err: ErrExtractFromContext,
			},
		},
		{
			name: "should return error if can't get user withdrawals",
			prepare: func(s *mocks.MockRepository) {
				s.EXPECT().
					GetUserWithdrawals(gomock.Any(), 1).
					Return(nil, errInternal)
			},
			want: want{
				err: errInternal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			storage := mocks.NewMockRepository(ctrl)

			tt.prepare(storage)
			service.storage = storage

			ctx := context.WithValue(context.Background(), auth.UserIDKey, 1)

			if tt.want.err == ErrExtractFromContext {
				var k badContextKey = "bad_key"
				ctx = context.WithValue(context.Background(), k, 1)
			}

			result, err := service.GetUserWithdrawals(ctx)

			if err != nil {
				assert.Equal(t, tt.want.err, err)
			}

			assert.Equal(t, tt.want.withdrawals, result)
		})
	}
}

func genHashString(s string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)

	return string(hash)
}
