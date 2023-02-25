package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tony-spark/recipetor-backend/recipe-service/internal/config"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/service"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage/mongodb"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting recipe service...")

	err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load config")
	}

	stor, err := mongodb.NewStorage(config.Config.Mongo.DSN, config.Config.Mongo.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize storage")
	}
	log.Info().Msg("connected to MongoDB")

	_ = service.NewService(stor)

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	log.Info().Msg("recipe service interrupted via system signal")
}
