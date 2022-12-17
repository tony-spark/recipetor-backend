package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/random"
	"log"
	"time"
)

const (
	TOPIC_REGISTRATION_REQ = "user.registration.req"
	TOPIC_REGISTRATIONS    = "user.registrations"
)

type CreateUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegistrationDTO struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Error string `json:"error,omitempty"`
}

type UserServiceSuite struct {
	suite.Suite

	registrationsWriter *kafka.Writer
	registrationsReader *kafka.Reader

	generator random.Generator
}

func (suite *UserServiceSuite) SetupSuite() {
	suite.registrationsWriter = newWriter([]string{"localhost:29092"}, TOPIC_REGISTRATION_REQ)
	suite.registrationsReader = newReader([]string{"localhost:29092"}, TOPIC_REGISTRATIONS)
	suite.generator = random.NewRandomGenerator()
}

func (suite *UserServiceSuite) TearDownSuite() {
	err := suite.registrationsWriter.Close()
	if err != nil {
		log.Println(err)
		return
	}
}

func (suite *UserServiceSuite) TestUserService() {
	suite.Run("user registration", func() {
		dto := suite.randomCreateUser()
		suite.write(suite.registrationsWriter, dto.Email, dto)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		for {
			message, err := suite.registrationsReader.ReadMessage(ctx)
			require.NoError(suite.T(), err, "ошибка при чтении сообщения")
			var registrationDTO UserRegistrationDTO
			err = json.Unmarshal(message.Value, &registrationDTO)
			require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
			if registrationDTO.Email != dto.Email {
				continue
			}
			assert.Empty(suite.T(), registrationDTO.Error)
			assert.NotEmpty(suite.T(), registrationDTO.ID)
			return
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
		GroupID:          "test-client",
		MaxWait:          2 * time.Second,
		ReadBatchTimeout: 2 * time.Second,
		MinBytes:         10,
		MaxBytes:         1024 * 1024,
	}
	err := config.Validate()
	if err != nil {
		log.Println(err)
	}
	return kafka.NewReader(config)
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
