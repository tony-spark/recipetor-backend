package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/user-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
)

const (
	TOPIC_REGISTRATION_REQ = "user.registration.req"
	TOPIC_REGISTRATIONS    = "user.registrations"
)

type kafkaController struct {
	userService         service.Service
	regReqReader        *kafka.Reader
	registrationsWriter *kafka.Writer
}

func NewController(userService service.Service, kafkaBrokerURLs string) (controller.Controller, error) {
	brokers := strings.Split(kafkaBrokerURLs, ",")
	regReqReader, err := newReader(brokers, TOPIC_REGISTRATION_REQ)
	if err != nil {
		return nil, err
	}
	registrationsWriter := newWriter(brokers, TOPIC_REGISTRATIONS)
	return kafkaController{
		userService:         userService,
		regReqReader:        regReqReader,
		registrationsWriter: registrationsWriter,
	}, nil
}

func newReader(brokers []string, topic string) (*kafka.Reader, error) {
	config := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "user-service",
	}
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid kafka config: %w", err)
	}
	return kafka.NewReader(config), nil
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func (k kafkaController) Run() error {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		m, err := k.regReqReader.ReadMessage(context.Background())
		cancel()
		if err != nil {
			log.Error().Err(err).Msg("error receiving message")
			continue
		}

		var dto user.CreateUserDTO
		err = json.Unmarshal(m.Value, &dto)
		if err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			continue
		}
		log.Info().Msgf("got: %+v", dto)

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		id, err := k.userService.Create(ctx, dto)
		registrationDTO := user.UserRegistrationDTO{
			ID:    id,
			Email: dto.Email,
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to create user")
			registrationDTO.Error = err.Error()
			write(k.registrationsWriter, dto.Email, registrationDTO)
			cancel()
			continue
		}

		log.Info().Msgf("send: %+v", registrationDTO)
		write(k.registrationsWriter, dto.Email, registrationDTO)
		cancel()
	}
	return nil
}

func write(writer *kafka.Writer, key string, msg interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bs, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed marshal outcoming message")
		return
	}

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bs,
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to write message")
	}
}

func (k kafkaController) Stop() error {
	var result error
	if err := k.regReqReader.Close(); err != nil {
		result = multierror.Append(result, err)
	}
	if err := k.registrationsWriter.Close(); err != nil {
		result = multierror.Append(result, err)
	}
	return result
}
