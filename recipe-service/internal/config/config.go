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
}

func Parse() error {
	flag.StringVar(&Config.Mongo.DSN, "d", "", "mongodb connection string")
	flag.StringVar(&Config.Mongo.DB, "n", "", "mongodb database name")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	log.Info().Msgf("config loaded: %+v", Config)
	return nil
}
