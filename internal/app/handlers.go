package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/PrahaTurbo/gophermart/internal/auth"
	"github.com/PrahaTurbo/gophermart/internal/models"
	"github.com/PrahaTurbo/gophermart/internal/service"
	"github.com/PrahaTurbo/gophermart/internal/storage"
	"github.com/pkg/errors"
)

func (a *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		a.log.Error().Err(err).Msg("cannot unmarshal response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Password == "" || user.Login == "" {
		a.log.Error().Msg("password or login is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := a.service.CreateUser(r.Context(), user)
	if errors.Is(err, storage.ErrAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie, err := auth.CreateJWTAuthCookie(userID, a.jwtSecret)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to create jwt auth cookie")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (a *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.UserRequest

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		a.log.Error().Err(err).Msg("cannot unmarshal response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Password == "" || user.Login == "" {
		a.log.Error().Msg("password or login is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := a.service.LoginUser(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cookie, err := auth.CreateJWTAuthCookie(userID, a.jwtSecret)
	if err != nil {
		a.log.Error().Err(err).Msg("failed to create jwt auth cookie")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (a *application) processOrderHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.service.ProcessOrder(r.Context(), string(body))

	if errors.Is(err, service.ErrInvalidOrderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if errors.Is(err, service.ErrOrderByAnotherUser) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if errors.Is(err, service.ErrOrderByCurrentUser) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (a *application) getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := a.service.GetUserOrders(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (a *application) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	balance, err := a.service.GetBalance(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *application) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	var withdrawReq models.WithdrawRequest

	if err := json.NewDecoder(r.Body).Decode(&withdrawReq); err != nil {
		a.log.Error().Err(err).Msg("cannot unmarshal response")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := a.service.Withdraw(r.Context(), withdrawReq)

	if errors.Is(err, service.ErrInvalidOrderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if errors.Is(err, service.ErrBalanceNotEnough) {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *application) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	withdrawals, err := a.service.GetUserWithdrawals(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
