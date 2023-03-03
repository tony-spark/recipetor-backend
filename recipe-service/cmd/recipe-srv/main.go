package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tony-spark/recipetor-backend/recipe-service/internal/config"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/controller/kafka"
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
	logLevel, err := zerolog.ParseLevel(config.Config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("unknown log level")
	}
	log.Logger = log.Logger.Level(logLevel)

	stor, err := mongodb.NewStorage(config.Config.Mongo.DSN, config.Config.Mongo.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize storage")
	}
	log.Info().Msg("connected to MongoDB")

	recipeService := service.NewService(stor)

	controller, err := kafka.NewController(recipeService, config.Config.Kafka.Brokers)
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

	log.Info().Msg("recipe service interrupted via system signal")
}
