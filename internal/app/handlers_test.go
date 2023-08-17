package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PrahaTurbo/gophermart/internal/auth"
	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/mocks"
	"github.com/PrahaTurbo/gophermart/internal/models"
	"github.com/PrahaTurbo/gophermart/internal/service"
	"github.com/PrahaTurbo/gophermart/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_application_registerUserHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode int
		cookie     *http.Cookie
	}

	tests := []struct {
		name        string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should successfully register user",
			requestBody: `{"login": "test", "password": "test_password"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					CreateUser(gomock.Any(), models.UserRequest{
						Login:    "test",
						Password: "test_password",
					}).
					Return(1, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				cookie: &http.Cookie{
					Name: auth.JWTTokenCookieName,
				},
			},
		},
		{
			name:        "should return 204 if user already exist",
			requestBody: `{"login": "test", "password": "test_password"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					CreateUser(gomock.Any(), models.UserRequest{
						Login:    "test",
						Password: "test_password",
					}).
					Return(0, storage.ErrAlreadyExist)
			},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name:        "should return 500 if can't save user",
			requestBody: `{"login": "test", "password": "test_password"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					CreateUser(gomock.Any(), models.UserRequest{
						Login:    "test",
						Password: "test_password",
					}).
					Return(0, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:        "should return 400 http error if login or password is empty",
			requestBody: `{"login": "", "password": "test_password"}`,
			prepare:     func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "should return 400 http error if body fields are incorrect",
			requestBody: `{"email": "test", "password": "test_password"}`,
			prepare:     func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)

			app.service = service

			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, "/api/user/register", reader)

			w := httptest.NewRecorder()
			app.registerUserHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.cookie != nil {
				cookie := result.Cookies()[0]
				assert.Equal(t, tt.want.cookie.Name, cookie.Name)
				assert.NotEmpty(t, cookie.Value)
			}
		})
	}
}

func Test_application_loginUserHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode int
		cookie     *http.Cookie
	}

	tests := []struct {
		name        string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should successfully login user",
			requestBody: `{"login": "test", "password": "test_password"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					LoginUser(gomock.Any(), models.UserRequest{
						Login:    "test",
						Password: "test_password",
					}).
					Return(1, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				cookie: &http.Cookie{
					Name: auth.JWTTokenCookieName,
				},
			},
		},
		{
			name:        "should return 401 http error when login and password don't match",
			requestBody: `{"login": "test", "password": "test_password"}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					LoginUser(gomock.Any(), models.UserRequest{
						Login:    "test",
						Password: "test_password",
					}).
					Return(0, errors.New("Unathorized"))
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:        "should return 400 http error if login or password is empty",
			requestBody: `{"login": "", "password": "test_password"}`,
			prepare:     func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "should return 400 http error if body fields are incorrect",
			requestBody: `{"email": "test", "password": "test_password"}`,
			prepare:     func(s *mocks.MockService) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)

			app.service = service

			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, "/api/user/login", reader)

			w := httptest.NewRecorder()
			app.loginUserHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.cookie != nil {
				cookie := result.Cookies()[0]
				assert.Equal(t, tt.want.cookie.Name, cookie.Name)
				assert.NotEmpty(t, cookie.Value)
			}
		})
	}
}

func Test_application_processOrderHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should successfully accept order",
			requestBody: "12345678903",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					ProcessOrder(gomock.Any(), "12345678903").
					Return(nil)
			},
			want: want{
				statusCode: http.StatusAccepted,
			},
		},
		{
			name:        "should return 422 when order id doesn't match luhn algorithm",
			requestBody: "1234567890",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					ProcessOrder(gomock.Any(), "1234567890").
					Return(service.ErrInvalidOrderID)
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name:        "should return 409 when order id was uploaded by another user",
			requestBody: "12345678903",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					ProcessOrder(gomock.Any(), "12345678903").
					Return(service.ErrOrderByAnotherUser)
			},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name:        "should return 200 when order id was uploaded by сurrent user",
			requestBody: "12345678903",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					ProcessOrder(gomock.Any(), "12345678903").
					Return(service.ErrOrderByCurrentUser)
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "should return 500 when internal server error",
			requestBody: "12345678903",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					ProcessOrder(gomock.Any(), "12345678903").
					Return(errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)

			app.service = service

			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, "/api/user/orders", reader)

			w := httptest.NewRecorder()
			app.processOrderHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

func Test_application_getOrdersHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockService)
		want    want
	}{
		{
			name: "should successfully return list of orders",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserOrders(gomock.Any()).
					Return([]models.OrderResponse{
						{
							ID:         "123",
							Accrual:    23,
							Status:     "status",
							UploadedAt: "date",
						},
					}, nil)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "should return 204 if list of orders is empty",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserOrders(gomock.Any()).
					Return([]models.OrderResponse{}, nil)
			},
			want: want{
				statusCode: http.StatusNoContent,
			},
		},
		{
			name: "should return 500 when internal error",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserOrders(gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.service = service

			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

			w := httptest.NewRecorder()
			app.getOrdersHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			}
		})
	}
}

func Test_application_getBalanceHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockService)
		want    want
	}{
		{
			name: "should successfully return balance",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetBalance(gomock.Any()).
					Return(&models.BalanceResponse{
						Current:   100,
						Withdrawn: 75,
					}, nil)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "should return 500 when internal error",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetBalance(gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.service = service

			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)

			w := httptest.NewRecorder()
			app.getBalanceHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			}
		})
	}
}

func Test_application_withdrawHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		requestBody string
		prepare     func(s *mocks.MockService)
		want        want
	}{
		{
			name:        "should successfully withdraw from balance",
			requestBody: `{"order": "12345678903", "sum": 12}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					Withdraw(gomock.Any(), models.WithdrawRequest{
						Order: "12345678903",
						Sum:   12,
					}).
					Return(nil)
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "should return 422 when order id doesn't match luhn algorithm",
			requestBody: `{"order": "1234567890", "sum": 12}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					Withdraw(gomock.Any(), models.WithdrawRequest{
						Order: "1234567890",
						Sum:   12,
					}).
					Return(service.ErrInvalidOrderID)
			},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name:        "should return 402 when balance is lower than withdraw sum",
			requestBody: `{"order": "12345678903", "sum": 12}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					Withdraw(gomock.Any(), models.WithdrawRequest{
						Order: "12345678903",
						Sum:   12,
					}).
					Return(service.ErrBalanceNotEnough)
			},
			want: want{
				statusCode: http.StatusPaymentRequired,
			},
		},
		{
			name:        "should return 500 when internal error",
			requestBody: `{"order": "12345678903", "sum": 12}`,
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					Withdraw(gomock.Any(), models.WithdrawRequest{
						Order: "12345678903",
						Sum:   12,
					}).
					Return(errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)

			app.service = service

			reader := strings.NewReader(tt.requestBody)
			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", reader)

			w := httptest.NewRecorder()
			app.withdrawHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

func Test_application_getWithdrawalsHandler(t *testing.T) {
	app := application{
		log: logger.NewLogger(),
	}

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name    string
		prepare func(s *mocks.MockService)
		want    want
	}{
		{
			name: "should successfully return list of orders",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserWithdrawals(gomock.Any()).
					Return([]models.WithdrawalsResponse{
						{
							Order:       "123",
							Sum:         23,
							ProcessedAt: "date",
						},
					}, nil)
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "should return 204 if list of withdrawals is empty",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserWithdrawals(gomock.Any()).
					Return([]models.WithdrawalsResponse{}, nil)
			},
			want: want{
				statusCode: http.StatusNoContent,
			},
		},
		{
			name: "should return 500 when internal error",
			prepare: func(s *mocks.MockService) {
				s.EXPECT().
					GetUserWithdrawals(gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := mocks.NewMockService(ctrl)

			tt.prepare(service)
			app.service = service

			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

			w := httptest.NewRecorder()
			app.getWithdrawalsHandler(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			}
		})
	}
}
