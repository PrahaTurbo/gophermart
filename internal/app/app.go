package app

import (
	"github.com/go-chi/chi/v5"

	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/service"
)

type App interface {
	Router() chi.Router
}

type application struct {
	jwtSecret string
	service   service.Service
	log       logger.Logger
}

func NewApp(jwtSecret string, srv service.Service, logger logger.Logger) App {
	return &application{
		jwtSecret: jwtSecret,
		service:   srv,
		log:       logger,
	}
}
