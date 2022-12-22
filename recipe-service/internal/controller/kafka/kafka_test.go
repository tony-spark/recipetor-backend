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
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/random"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/service"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage/mongodb"
	"io"
	"os"
	"testing"
	"time"
)

type ControllerTestSuite struct {
	suite.Suite

	controller controller.Controller

	recipesReader        *kafka.Reader
	newRecipeWriter      *kafka.Writer
	reqRecipeWriter      *kafka.Writer
	nutritionFactsWriter *kafka.Writer

	rand random.Generator

	cleanupFunc func(ctx context.Context) error
}

func (suite *ControllerTestSuite) TestController() {
	suite.Run("create recipe send nutrition facts and get by id", func() {
		newRecipeDTO := suite.randomCreateRecipe()
		write(suite.newRecipeWriter, newRecipeDTO.Name, newRecipeDTO)

		var createdDTO recipe.RecipeDTO
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.recipesReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				if string(message.Key) != newRecipeDTO.Name {
					continue
				}

				err = json.Unmarshal(message.Value, &createdDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), createdDTO.Error)
				assert.NotEmpty(suite.T(), createdDTO.ID)
				assert.NotEmpty(suite.T(), createdDTO.Recipe)
				assert.Equal(suite.T(), newRecipeDTO.Name, createdDTO.Recipe.Name)
				assert.Equal(suite.T(), newRecipeDTO.Ingredients, createdDTO.Recipe.Ingredients)
				assert.Equal(suite.T(), newRecipeDTO.Steps, createdDTO.Recipe.Steps)
				break
			}
		}

		recipeNutritionsDTO := recipe.RecipeNutritionsDTO{
			RecipeID:       createdDTO.ID,
			NutritionFacts: suite.randomNutritionFacts(),
			Inaccurate:     false,
		}
		write(suite.nutritionFactsWriter, recipeNutritionsDTO.RecipeID, recipeNutritionsDTO)

		findRecipeDTO := recipe.FindRecipeDTO{
			ID: createdDTO.ID,
		}
		write(suite.reqRecipeWriter, findRecipeDTO.ID, findRecipeDTO)

		var gotDTO recipe.RecipeDTO
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.recipesReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				if string(message.Key) != findRecipeDTO.ID {
					continue
				}

				err = json.Unmarshal(message.Value, &gotDTO)
				require.NoError(suite.T(), err)
				assert.Empty(suite.T(), gotDTO.Error)
				assert.Equal(suite.T(), findRecipeDTO.ID, gotDTO.ID)
				assert.NotEmpty(suite.T(), gotDTO.Recipe)
				assert.Equal(suite.T(), createdDTO.Recipe.Steps, gotDTO.Recipe.Steps)
				assert.Equal(suite.T(), createdDTO.Recipe.Ingredients, gotDTO.Recipe.Ingredients)
				assert.Equal(suite.T(), recipeNutritionsDTO.NutritionFacts, *(gotDTO.Recipe.NutritionFacts))
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

	err = createTopics(kafkaBroker, TopicRecipesNew, TopicRecipesReq, TopicRecipes, TopicNutritionFacts)
	suite.Require().NoError(err)

	{
		var stor storage.Storage
		stor, suite.cleanupFunc, err = mongodb.NewTestStorage(dsn, "test")
		suite.Require().NoError(err)

		suite.controller, err = NewController(service.NewService(stor), kafkaBroker)
		suite.Require().NoError(err)
	}

	suite.recipesReader, err = newReader([]string{kafkaBroker}, "recipe-service-test-recipes", TopicRecipes)
	suite.Require().NoError(err)

	suite.newRecipeWriter = newWriter([]string{kafkaBroker}, TopicRecipesNew)
	suite.reqRecipeWriter = newWriter([]string{kafkaBroker}, TopicRecipesReq)
	suite.nutritionFactsWriter = newWriter([]string{kafkaBroker}, TopicNutritionFacts)

	suite.rand = random.NewRandomGenerator()

	go func() {
		err := suite.controller.Run(context.Background())
		if err != nil {
			suite.Require().ErrorIs(err, io.EOF)
		}
	}()
}

func (suite *ControllerTestSuite) TearDownSuite() {
	err := closeAll(suite.recipesReader, suite.reqRecipeWriter, suite.newRecipeWriter, suite.nutritionFactsWriter)
	suite.Assert().NoError(err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = suite.cleanupFunc(ctx)
	suite.Assert().NoError(err)

	err = suite.controller.Stop()
	suite.Assert().NoError(err)
}

func (suite *ControllerTestSuite) randomCreateRecipe() recipe.CreateRecipeDTO {
	return recipe.CreateRecipeDTO{
		Name:      suite.rand.RandomString(8),
		CreatedBy: suite.rand.RandomObjectID(),
		Ingredients: []recipe.RecipeIngredient{
			{
				IngredientID: suite.rand.RandomObjectID(),
				Unit:         suite.rand.RandomString(3),
				Amount:       suite.rand.RandomFloat(200),
			},
			{
				IngredientID: suite.rand.RandomObjectID(),
				Unit:         suite.rand.RandomString(3),
				Amount:       suite.rand.RandomFloat(200),
			},
			{
				IngredientID: suite.rand.RandomObjectID(),
				Unit:         suite.rand.RandomString(3),
				Amount:       suite.rand.RandomFloat(200),
			},
		},
		Steps: []recipe.Step{
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
		},
	}
}

func (suite *ControllerTestSuite) randomNutritionFacts() recipe.NutritionFacts {
	return recipe.NutritionFacts{
		Calories:      suite.rand.RandomFloat(1000),
		Proteins:      suite.rand.RandomFloat(100),
		Fats:          suite.rand.RandomFloat(100),
		Carbohydrates: suite.rand.RandomFloat(100),
	}
}

func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
