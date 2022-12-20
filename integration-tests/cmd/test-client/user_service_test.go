package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/random"
)

const (
	TopicRegistrationReq = "user.registration.req"
	TopicLoginReq        = "user.login.req"
	TopicRegistrations   = "user.registrations"
	TopicLogins          = "user.logins"
)

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
		suite.write(suite.registrationsWriter, registerDTO.Email, registerDTO)

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
		suite.write(suite.loginsWriter, loginDTO.Email, loginDTO)

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

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func newReader(brokers []string, topic string) *kafka.Reader {
	config := kafka.ReaderConfig{
		Brokers:          brokers,
		Topic:            topic,
		GroupID:          "test-client" + topic,
		MaxWait:          1 * time.Second,
		ReadBatchTimeout: 2 * time.Second,
	}
	err := config.Validate()
	if err != nil {
		log.Println(err)
	}
	return kafka.NewReader(config)
}

func closeAll(closers ...io.Closer) error {
	var result error
	for _, closer := range closers {
		err := closer.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (suite *UserServiceSuite) write(writer *kafka.Writer, key string, msg interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bs, err := json.Marshal(msg)
	require.NoError(suite.T(), err)

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bs,
	})

	require.NoError(suite.T(), err, "не удалось записать сообщение")
}

func (suite *UserServiceSuite) randomCreateUser() CreateUserDTO {
	return CreateUserDTO{
		Email:    suite.generator.RandomEmail(),
		Password: suite.generator.RandomString(8),
	}
}
