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
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/service"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage/mongodb"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/random"
	"io"
	"os"
	"testing"
	"time"
)

type ControllerTestSuite struct {
	suite.Suite

	controller controller.Controller

	newIngredientWriter  *kafka.Writer
	ingredientsReader    *kafka.Reader
	reqIngredientsWriter *kafka.Writer

	rand random.Generator

	cleanupFunc func(ctx context.Context) error
}

func (suite *ControllerTestSuite) TestController() {
	suite.Run("create ingredient and find by id", func() {
		newIngredientDTO := suite.randomCreateIngredient()
		corID := generateCorrelationID()
		write(suite.newIngredientWriter, newIngredientDTO.Name, newIngredientDTO, corID)

		var createdDTO ingredient.IngredientDTO
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.ingredientsReader.ReadMessage(ctx)
				if !checkCorrelationID(message, corID) {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")

				err = json.Unmarshal(message.Value, &createdDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), createdDTO.Error)
				assert.NotEmpty(suite.T(), createdDTO.ID)
				assert.Equal(suite.T(), newIngredientDTO.Name, createdDTO.Name)
				assert.NotEmpty(suite.T(), createdDTO.Ingredient)
				break
			}
		}

		findDTO := ingredient.FindIngredientsDTO{
			ID: createdDTO.ID,
		}
		write(suite.reqIngredientsWriter, findDTO.ID, findDTO, corID)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.ingredientsReader.ReadMessage(ctx)
				if !checkCorrelationID(message, corID) {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")

				var gotDTO ingredient.IngredientDTO
				err = json.Unmarshal(message.Value, &gotDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), gotDTO.Error)
				assert.Equal(suite.T(), createdDTO.Ingredient, gotDTO.Ingredient)
				break
			}
		}
	})

}

func (suite *ControllerTestSuite) SetupSuite() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).Level(zerolog.ErrorLevel)
	kafkaBroker := os.Getenv("TEST_KAFKA_BROKERS")
	if len(kafkaBroker) == 0 {
		kafkaBroker = "localhost:29092"
	}
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}

	var err error

	err = createTopics(kafkaBroker, TopicIngredients, TopicIngredientsReq, TopicIngredientsNew)
	suite.Require().NoError(err)

	{
		var stor storage.Storage
		stor, suite.cleanupFunc, err = mongodb.NewTestStorage(dsn, "test")
		suite.Require().NoError(err)

		suite.controller, err = NewController(service.NewService(stor), kafkaBroker)
		suite.Require().NoError(err)
	}

	suite.ingredientsReader, err = newReader([]string{kafkaBroker}, "ingredient-service-test-ingredients", TopicIngredients)
	suite.Require().NoError(err)

	suite.newIngredientWriter = newWriter([]string{kafkaBroker}, TopicIngredientsNew)
	suite.reqIngredientsWriter = newWriter([]string{kafkaBroker}, TopicIngredientsReq)

	suite.rand = random.NewRandomGenerator()

	go func() {
		err := suite.controller.Run(context.Background())
		if err != nil {
			suite.Require().ErrorIs(err, io.EOF)
		}
	}()
}

func (suite *ControllerTestSuite) TearDownSuite() {
	err := closeAll(suite.ingredientsReader, suite.newIngredientWriter, suite.reqIngredientsWriter)
	suite.Assert().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = suite.cleanupFunc(ctx)
	suite.Assert().NoError(err)

	err = suite.controller.Stop()
	suite.Assert().NoError(err)
}

func (suite *ControllerTestSuite) randomCreateIngredient() ingredient.CreateIngredientDTO {
	return ingredient.CreateIngredientDTO{
		Name:     suite.rand.RandomString(8),
		BaseUnit: suite.rand.RandomString(2),
		NutritionFacts: &ingredient.NutritionFacts{
			Calories:      suite.rand.RandomFloat(500),
			Proteins:      suite.rand.RandomFloat(100),
			Fats:          suite.rand.RandomFloat(100),
			Carbohydrates: suite.rand.RandomFloat(100),
		},
	}
}

func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
