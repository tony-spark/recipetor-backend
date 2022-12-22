package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/ingredient"
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
	loginsWriter        *kafka.Writer
	loginsReader        *kafka.Reader

	// Ingredients service
	newIngredientWriter  *kafka.Writer
	ingredientsReader    *kafka.Reader
	reqIngredientsWriter *kafka.Writer

	rand random.Generator
}

func (suite *ScenariosTestSuite) TestScenarios() {
	suite.Run("scenario 1", func() {
		corID := generateCorrelationID()

		var userID string
		var ingredientIDs []string

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

		// 3. Add ingredients
		createIngredientDTOs := []ingredient.CreateIngredientDTO{
			suite.createIngredient("пшеничная мука", "гр", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("кефир", "мл", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("яйцо", "шт", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("сахар", "г", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("соль", "г", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("сода", "г", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("растительное масло", "г", 0.0, 0.0, 0.0, 0.0),
		}
		for _, createIngredientDTO := range createIngredientDTOs {
			write(suite.T(), suite.newIngredientWriter, createIngredientDTO.Name, createIngredientDTO, corID)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.ingredientsReader.ReadMessage(ctx)
				if !checkCorrelationID(message, corID) {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")

				var createdDTO ingredient.IngredientDTO
				err = json.Unmarshal(message.Value, &createdDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), createdDTO.Error)
				assert.NotEmpty(suite.T(), createdDTO.ID)
				ingredientIDs = append(ingredientIDs, createdDTO.ID)
				break
			}
		}
		suite.Require().Equal(len(createIngredientDTOs), len(ingredientIDs))

		// 4. Add recipe

		// 5. Make sure recipe's nutrition facts is calculated

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

	suite.ingredientsReader, err = newReader([]string{flagKafkaBroker}, suite.clientID, TopicIngredients)
	suite.Require().NoError(err)

	suite.newIngredientWriter = newWriter([]string{flagKafkaBroker}, TopicIngredientsNew)
	suite.reqIngredientsWriter = newWriter([]string{flagKafkaBroker}, TopicIngredientsReq)
}

func (suite *ScenariosTestSuite) TearDownSuite() {
	err := closeAll(suite.registrationsReader, suite.registrationsWriter, suite.loginsReader, suite.loginsWriter,
		suite.ingredientsReader, suite.newIngredientWriter, suite.reqIngredientsWriter)
	suite.Assert().NoError(err)
}

func (suite *ScenariosTestSuite) randomCreateUser() user.CreateUserDTO {
	return user.CreateUserDTO{
		Email:    suite.clientID + "." + suite.rand.RandomEmail(),
		Password: suite.rand.RandomString(8),
	}
}

func (suite *ScenariosTestSuite) createIngredient(name string, baseUnit string,
	calories float64, proteins float64, fats float64, carbohydrates float64) ingredient.CreateIngredientDTO {
	return ingredient.CreateIngredientDTO{
		Name:     name + "-" + suite.clientID,
		BaseUnit: baseUnit,
		NutritionFacts: &ingredient.NutritionFacts{
			Calories:      calories,
			Proteins:      proteins,
			Fats:          fats,
			Carbohydrates: carbohydrates,
		},
	}
}
