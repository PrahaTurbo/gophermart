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

		r.Post("/api/user/orders", a.ordersHandler)
		r.Get("/api/user/orders", a.getOrdersHandler)
		r.Get("/api/user/balance", a.balanceHandler)
		r.Post("/api/user/balance/withdraw", a.withdrawHandler)
		r.Get("/api/user/withdrawals", a.withdrawalsHandler)
	})

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", a.registerHandler)
		r.Post("/api/user/login", a.loginHandler)
	})

	return r
}
