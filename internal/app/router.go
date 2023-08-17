package app

import (
	"github.com/PrahaTurbo/gophermart/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *application) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(a.log.RequestLogger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(auth.Auth(a.jwtSecret))

		r.Post("/api/user/orders", a.processOrderHandler)
		r.Get("/api/user/orders", a.getOrdersHandler)
		r.Get("/api/user/balance", a.getBalanceHandler)
		r.Post("/api/user/balance/withdraw", a.withdrawHandler)
		r.Get("/api/user/withdrawals", a.getWithdrawalsHandler)
	})

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", a.registerUserHandler)
		r.Post("/api/user/login", a.loginUserHandler)
	})

	return r
}
