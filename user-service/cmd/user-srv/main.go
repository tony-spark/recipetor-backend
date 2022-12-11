package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/user-service/internal/config"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Info().Msg("starting user service")

	err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("could not load config")
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(config.Config.MongoDSN))
	if err != nil {
		log.Fatal().Err(err).Msg("could not create connection to MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to MongoDB")
	}
	log.Info().Msg("connected to MongoDB")

	db := mongoClient.Database("test")
	storage := mongodb.NewStorage(db)
	_ = service.NewService(storage)

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	log.Info().Msg("user service interrupted via system signal")
}
