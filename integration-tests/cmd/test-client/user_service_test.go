package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/random"
)

type UserServiceSuite struct {
	suite.Suite

	registrationsWriter *kafka.Writer
	registrationsReader *kafka.Reader

	loginsWriter *kafka.Writer
	loginsReader *kafka.Reader

	generator random.Generator
}

func (suite *UserServiceSuite) SetupSuite() {
	suite.Require().NotEmpty(flagKafkaBroker, "--kafka-broker flag required")

	createTopics(flagKafkaBroker, TopicRegistrationReq, TopicRegistrations, TopicLoginReq, TopicLogins)

	suite.registrationsWriter = newWriter([]string{flagKafkaBroker}, TopicRegistrationReq)
	suite.registrationsReader = newReader([]string{flagKafkaBroker}, TopicRegistrations)
	suite.loginsWriter = newWriter([]string{flagKafkaBroker}, TopicLoginReq)
	suite.loginsReader = newReader([]string{flagKafkaBroker}, TopicLogins)
	suite.generator = random.NewRandomGenerator()
}

func (suite *UserServiceSuite) TearDownSuite() {
	err := closeAll(suite.registrationsReader, suite.registrationsWriter,
		suite.loginsReader, suite.loginsWriter)
	assert.NoError(suite.T(), err)
}

func (suite *UserServiceSuite) TestUserService() {
	suite.Run("user registration and login", func() {
		registerDTO := suite.randomCreateUser()
		write(suite.T(), suite.registrationsWriter, registerDTO.Email, registerDTO)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.registrationsReader.ReadMessage(ctx)
				if string(message.Key) != registerDTO.Email {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				var registrationDTO UserRegistrationDTO
				err = json.Unmarshal(message.Value, &registrationDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), registrationDTO.Error)
				assert.NotEmpty(suite.T(), registrationDTO.ID)
				break
			}
		}

		var loginDTO = registerDTO
		write(suite.T(), suite.loginsWriter, loginDTO.Email, loginDTO)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.loginsReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				var userLoginDTO UserLoginDTO
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

func (suite *UserServiceSuite) randomCreateUser() CreateUserDTO {
	return CreateUserDTO{
		Email:    suite.generator.RandomEmail(),
		Password: suite.generator.RandomString(8),
	}
}

type CreateUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegistrationDTO struct {
	ID    string `json:"user_id,omitempty"`
	Email string `json:"email,omitempty"`
	Error string `json:"error,omitempty"`
}

type User struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	Email        string    `json:"email" bson:"email,omitempty"`
	Password     string    `json:"-" bson:"password,omitempty"`
	RegisteredAt time.Time `json:"registered_at" bson:"registered_at,omitempty"`
}

type UserLoginDTO struct {
	User  User   `json:"user,omitempty"`
	Email string `json:"email"`
	Error string `json:"error,omitempty"`
}
