package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

var (
	Config config
)

type config struct {
	Mongo struct {
		DSN string `env:"MONGO_DSN"`
		DB  string `env:"MONGO_DB"`
	}
	Kafka struct {
		Brokers string `env:"KAFKA_BROKERS"`
	}
}

func Parse() error {
	flag.StringVar(&Config.Mongo.DSN, "mongo-dsn", "", "mongodb connection string")
	flag.StringVar(&Config.Mongo.DB, "mongo-db", "", "mongodb database name")
	flag.StringVar(&Config.Kafka.Brokers, "kafka-brokers", "", "kafka broker list")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	log.Info().Msgf("config loaded: %+v", Config)
	return nil
}
