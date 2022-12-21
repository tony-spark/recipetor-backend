package kafka

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/user-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/user-service/internal/random"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage/mongodb"
	"io"
	"os"
	"testing"
	"time"
)

type ControllerTestSuite struct {
	suite.Suite

	controller controller.Controller

	registrationsWriter *kafka.Writer
	registrationsReader *kafka.Reader

	loginsWriter *kafka.Writer
	loginsReader *kafka.Reader

	rand random.Generator

	cleanupFunc func(ctx context.Context) error
}

func (suite *ControllerTestSuite) TestController() {
	suite.Run("user registration and login", func() {
		registerDTO := suite.randomCreateUser()
		write(suite.registrationsWriter, registerDTO.Email, registerDTO)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.registrationsReader.ReadMessage(ctx)
				if string(message.Key) != registerDTO.Email {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				var registrationDTO user.UserRegistrationDTO
				err = json.Unmarshal(message.Value, &registrationDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), registrationDTO.Error)
				assert.NotEmpty(suite.T(), registrationDTO.ID)
				break
			}
		}

		var loginDTO = registerDTO
		write(suite.loginsWriter, loginDTO.Email, loginDTO)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.loginsReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				var userLoginDTO user.UserLoginDTO
				err = json.Unmarshal(message.Value, &userLoginDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				if loginDTO.Email != registerDTO.Email {
					continue
				}
				assert.Empty(suite.T(), userLoginDTO.Error)
				assert.NotEmpty(suite.T(), userLoginDTO.User)
				break
			}
		}

	})
}

func (suite *ControllerTestSuite) SetupSuite() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).Level(zerolog.InfoLevel)
	kafkaBroker := os.Getenv("TEST_KAFKA_BROKERS")
	if len(kafkaBroker) == 0 {
		kafkaBroker = "localhost:29092"
	}
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}

	var err error

	err = createTopics(kafkaBroker, TopicRegistrationReq, TopicRegistrations, TopicLoginReq, TopicLogins)
	suite.Require().NoError(err)

	{
		var stor storage.Storage
		stor, suite.cleanupFunc, err = mongodb.NewTestStorage(dsn, "test")
		suite.Require().NoError(err)

		suite.controller, err = NewController(service.NewService(stor), kafkaBroker)
		suite.Require().NoError(err)
	}

	suite.registrationsReader, err = newReader([]string{kafkaBroker}, "user-service-test-registrations", TopicRegistrations)
	suite.Require().NoError(err)

	suite.loginsReader, err = newReader([]string{kafkaBroker}, "user-service-test-logins", TopicLogins)
	suite.Require().NoError(err)

	suite.registrationsWriter = newWriter([]string{kafkaBroker}, TopicRegistrationReq)
	suite.loginsWriter = newWriter([]string{kafkaBroker}, TopicLoginReq)

	suite.rand = random.NewRandomGenerator()

	go func() {
		err := suite.controller.Run(context.Background())
		if err != nil {
			suite.Require().ErrorIs(err, io.EOF)
		}
	}()
}

func (suite *ControllerTestSuite) TearDownSuite() {
	err := closeAll(suite.registrationsReader, suite.registrationsWriter,
		suite.loginsReader, suite.loginsWriter)
	suite.Assert().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = suite.cleanupFunc(ctx)
	suite.Assert().NoError(err)

	err = suite.controller.Stop()
	suite.Assert().NoError(err)
}

func (suite *ControllerTestSuite) randomCreateUser() user.CreateUserDTO {
	return user.CreateUserDTO{
		Email:    suite.rand.RandomEmail(),
		Password: suite.rand.RandomString(8),
	}
}

func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
