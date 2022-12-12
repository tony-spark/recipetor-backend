package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/config"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage/mongodb"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting ingredient service...")

	err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load config")
	}

	_, err = mongodb.NewStorage(config.Config.Mongo.DSN, config.Config.Mongo.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize storage")
	}
	log.Info().Msg("connected to MongoDB")

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	log.Info().Msg("ingredient service interrupted via system signal")
}
