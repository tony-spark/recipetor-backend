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
	LogLevel string `env:"LOG_LEVEL"`
	Kafka    struct {
		Brokers string `env:"KAFKA_BROKERS"`
	}
}

func Parse() error {
	flag.StringVar(&Config.LogLevel, "log-level", "debug", "application log level")
	flag.StringVar(&Config.Kafka.Brokers, "kafka-brokers", "", "kafka broker list")
	flag.Parse()

	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	log.Info().Msgf("config loaded: %+v", Config)
	return nil
}
