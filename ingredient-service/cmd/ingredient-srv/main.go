package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/config"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/controller/kafka"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/service"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage/mongodb"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting ingredient service...")

	err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load config")
	}
	logLevel, err := zerolog.ParseLevel(config.Config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("unknown log level")
	}
	log.Logger.Level(logLevel)

	stor, err := mongodb.NewStorage(config.Config.Mongo.DSN, config.Config.Mongo.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize storage")
	}
	log.Info().Msg("connected to MongoDB")

	ingredientService := service.NewService(stor)

	controller, err := kafka.NewController(ingredientService, config.Config.Kafka.Brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka controller")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := controller.Run(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Fatal().Err(err).Msg("error running controller")
			}
		}
	}()

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	cancel()
	err = controller.Stop()
	if err != nil {
		log.Fatal().Err(err).Msg("controller failed to stop properly")
	}

	log.Info().Msg("ingredient service interrupted via system signal")
}
