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
	MongoDSN string `env:"MONGO_DSN"`
}

func Parse() error {
	flag.StringVar(&Config.MongoDSN, "d", "", "mongodb connection string")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	log.Info().Msgf("config loaded: %+v", Config)
	return nil
}
