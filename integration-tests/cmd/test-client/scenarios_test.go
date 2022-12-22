package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/random"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/user"
	"time"
)

type ScenariosTestSuite struct {
	suite.Suite

	clientID string

	// User service
	registrationsWriter *kafka.Writer
	registrationsReader *kafka.Reader

	loginsWriter *kafka.Writer
	loginsReader *kafka.Reader

	rand random.Generator
}

func (suite *ScenariosTestSuite) TestScenarios() {
	suite.Run("scenario 1", func() {
		corID := generateCorrelationID()

		var userID string

		// 1. User registration
		registerDTO := suite.randomCreateUser()
		write(suite.T(), suite.registrationsWriter, registerDTO.Email, registerDTO, corID)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.registrationsReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				if !checkCorrelationID(message, corID) {
					continue
				}
				var registrationDTO user.UserRegistrationDTO
				err = json.Unmarshal(message.Value, &registrationDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), registrationDTO.Error)
				assert.NotEmpty(suite.T(), registrationDTO.ID)

				userID = registrationDTO.ID
				break
			}
		}

		// 2. User login
		var loginDTO = registerDTO
		write(suite.T(), suite.loginsWriter, loginDTO.Email, loginDTO, corID)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.loginsReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				var userLoginDTO user.UserLoginDTO
				err = json.Unmarshal(message.Value, &userLoginDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				if !checkCorrelationID(message, corID) {
					continue
				}
				assert.Empty(suite.T(), userLoginDTO.Error)
				assert.NotEmpty(suite.T(), userLoginDTO.User)
				assert.Equal(suite.T(), userID, userLoginDTO.User.ID)
				break
			}
		}

	})
}

func (suite *ScenariosTestSuite) SetupSuite() {
	suite.Require().NotEmpty(flagKafkaBroker, "--kafka-broker flag required")

	suite.rand = random.NewGenerator()

	suite.clientID = uuid.NewString()

	var err error

	err = createTopics(flagKafkaBroker, TopicRegistrationReq, TopicRegistrations, TopicLoginReq, TopicLogins)
	suite.Require().NoError(err)

	suite.registrationsReader, err = newReader([]string{flagKafkaBroker}, suite.clientID, TopicRegistrations)
	suite.Require().NoError(err)

	suite.loginsReader, err = newReader([]string{flagKafkaBroker}, suite.clientID, TopicLogins)
	suite.Require().NoError(err)

	suite.registrationsWriter = newWriter([]string{flagKafkaBroker}, TopicRegistrationReq)
	suite.loginsWriter = newWriter([]string{flagKafkaBroker}, TopicLoginReq)
}

func (suite *ScenariosTestSuite) TearDownSuite() {
	err := closeAll(suite.registrationsReader, suite.registrationsWriter,
		suite.loginsReader, suite.loginsWriter)
	suite.Assert().NoError(err)
}

func (suite *ScenariosTestSuite) randomCreateUser() user.CreateUserDTO {
	return user.CreateUserDTO{
		Email:    suite.clientID + "." + suite.rand.RandomEmail(),
		Password: suite.rand.RandomString(8),
	}
}
