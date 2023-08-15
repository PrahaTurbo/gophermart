package main

import (
	"github.com/PrahaTurbo/gophermart/config"
	"github.com/PrahaTurbo/gophermart/internal/app"
	"github.com/PrahaTurbo/gophermart/internal/client"
	"github.com/PrahaTurbo/gophermart/internal/logger"
	"github.com/PrahaTurbo/gophermart/internal/service"
	"github.com/PrahaTurbo/gophermart/internal/storage"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	c := config.Load()
	log := logger.NewLogger()

	db, err := storage.SetupDB(c.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup database")
	}
	defer db.Close()

	storage := storage.NewStorage(db, log)
	accrualClient := client.NewAccrualClient(c.AccrualSysAddr)
	service := service.NewService(storage, accrualClient, log)
	application := app.NewApp(c.JWTSecret, service, log)

	log.Info().Str("address", c.RunAddr).Msg("server is running")
	if err := http.ListenAndServe(c.RunAddr, application.Router()); err != nil {
		log.Fatal().Err(err)
	}
}
