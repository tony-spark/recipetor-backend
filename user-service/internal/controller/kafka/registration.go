package kafka

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
)

type RegistrationWorker struct {
	userService         service.Service
	regReqReader        *kafka.Reader
	registrationsWriter *kafka.Writer
}

func NewRegistrationWorker(userService service.Service, brokers []string) (Worker, error) {
	regReqReader, err := newReader(brokers, "user-service-registrations", TopicRegistrationReq)
	if err != nil {
		return nil, err
	}
	registrationsWriter := newWriter(brokers, TopicRegistrations)
	return RegistrationWorker{
		userService:         userService,
		regReqReader:        regReqReader,
		registrationsWriter: registrationsWriter,
	}, nil
}

func (w RegistrationWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto user.CreateUserDTO
		corID, err := readDTO(ctx, w.regReqReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got CreateUserDTO: %+v", dto)

		cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
		id, err := w.userService.Create(cntx, dto)
		registrationDTO := user.UserRegistrationDTO{
			ID:    id,
			Email: dto.Email,
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to create user")
			registrationDTO.Error = err.Error()
		}
		cancel()

		write(w.registrationsWriter, dto.Email, registrationDTO, corID)
		log.Info().Msgf("sent UserRegistrationDTO: %+v", registrationDTO)
	}
}

func (w RegistrationWorker) Stop() error {
	return closeAll(w.regReqReader, w.registrationsWriter)
}
