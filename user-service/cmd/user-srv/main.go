package main

import (
	"context"
	"errors"
	"github.com/tony-spark/recipetor-backend/user-service/internal/controller/kafka"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/user-service/internal/config"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage/mongodb"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano})
	log.Info().Msg("starting user service")

	err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load config")
	}

	stor, err := mongodb.NewStorage(config.Config.Mongo.DSN, config.Config.Mongo.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize storage")
	}
	log.Info().Msg("connected to MongoDB")

	userService := service.NewService(stor)

	// TODO: move to config
	controller, err := kafka.NewController(userService, "localhost:29092")
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

	log.Info().Msg("user service interrupted via system signal")
}
