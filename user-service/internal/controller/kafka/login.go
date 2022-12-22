package kafka

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
	"io"
	"time"
)

type LoginWorker struct {
	userService    service.Service
	loginReqReader *kafka.Reader
	loginsWriter   *kafka.Writer
}

func NewLoginWorker(userService service.Service, brokers []string) (Worker, error) {
	loginReqReader, err := newReader(brokers, "user-service-logins", TopicLoginReq)
	if err != nil {
		return nil, err
	}
	loginWriter := newWriter(brokers, TopicLogins)
	return LoginWorker{
		userService:    userService,
		loginReqReader: loginReqReader,
		loginsWriter:   loginWriter,
	}, nil
}

func (w LoginWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var loginDTO user.LoginDTO
		corID, err := readDTO(ctx, w.loginReqReader, &loginDTO)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got LoginDTO: %+v", loginDTO)

		cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
		usr, err := w.userService.GetByEmailAndPassword(cntx, loginDTO.Email, loginDTO.Password)
		userLoginDTO := user.UserLoginDTO{
			User:  usr,
			Email: loginDTO.Email,
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to login user")
			userLoginDTO.Error = err.Error()
		}
		cancel()

		write(w.loginsWriter, loginDTO.Email, userLoginDTO, corID)
		log.Info().Msgf("sent UserLoginDTO: %+v", userLoginDTO)
	}
}

func (w LoginWorker) Stop() error {
	return closeAll(w.loginReqReader, w.loginsWriter)
}
