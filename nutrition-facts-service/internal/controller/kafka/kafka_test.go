package kafka

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/random"
	"io"
	"os"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition/service"
)

type ControllerTestSuite struct {
	suite.Suite

	controller controller.Controller

	recipesWriter     *kafka.Writer
	ingredientsWriter *kafka.Writer

	ingredientsReqReader *kafka.Reader
	nutritionFactsReader *kafka.Reader

	rand random.Generator
}

func (suite *ControllerTestSuite) TestController() {
	suite.Run("full result", func() {
		recipeDTO := nutrition.RecipeDTO{
			Recipe: nutrition.Recipe{
				ID: "1",
				Ingredients: []nutrition.RecipeIngredient{
					{
						IngredientID: "1",
						Unit:         "шт",
						Amount:       1,
					}, {
						IngredientID: "2",
						Unit:         "г",
						Amount:       15,
					}, {
						IngredientID: "3",
						Unit:         "мл",
						Amount:       65,
					},
				},
			},
			ID:     suite.rand.RandomObjectID(),
			UserID: suite.rand.RandomObjectID(),
		}
		write(suite.recipesWriter, recipeDTO.ID, recipeDTO)

		ingredients := map[string]nutrition.Ingredient{
			"1": {
				ID:       "1",
				BaseUnit: "шт",
				NutritionFacts: &nutrition.NutritionFacts{
					Calories:      50,
					Proteins:      5,
					Fats:          10,
					Carbohydrates: 15,
				},
			},
			"2": {
				ID:       "2",
				BaseUnit: "г",
				NutritionFacts: &nutrition.NutritionFacts{
					Calories:      1,
					Proteins:      0.5,
					Fats:          0.3,
					Carbohydrates: 0.1,
				},
			},
			"3": {
				ID:       "3",
				BaseUnit: "мл",
				NutritionFacts: &nutrition.NutritionFacts{
					Calories:      0.5,
					Proteins:      0,
					Fats:          0.8,
					Carbohydrates: 0.4,
				},
			},
		}

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ingredientsSent := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				var dto nutrition.FindIngredientsDTO
				err := readDTO(ctx, suite.ingredientsReqReader, &dto)
				suite.Require().NoError(err)

				ingr, ok := ingredients[dto.ID]
				suite.Assert().True(ok)
				if ok {
					write(suite.ingredientsWriter, dto.ID, nutrition.IngredientDTO{
						Ingredient: ingr,
						ID:         ingr.ID,
					})
					ingredientsSent++
					if ingredientsSent == len(recipeDTO.Recipe.Ingredients) {
						break
					}
				}

			}
		}

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var dto nutrition.RecipeNutritionsDTO
			err := readDTO(ctx, suite.nutritionFactsReader, &dto)
			suite.Assert().NoError(err)

			suite.Assert().Equal(recipeDTO.Recipe.ID, dto.RecipeID)
			suite.Assert().Equal(nutrition.RecipeNutritionsDTO{
				RecipeID: "1",
				NutritionFacts: nutrition.NutritionFacts{
					Calories:      15*1 + 65*0.5 + 50,
					Proteins:      15*0.5 + 5,
					Fats:          15*0.3 + 65*0.8 + 10,
					Carbohydrates: 15*0.1 + 65*0.4 + 15,
				},
				Inaccurate: false,
			}, dto)

		}
	})
}

func (suite *ControllerTestSuite) SetupSuite() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).Level(zerolog.InfoLevel)
	kafkaBroker := os.Getenv("TEST_KAFKA_BROKERS")
	if len(kafkaBroker) == 0 {
		kafkaBroker = "localhost:29092"
	}

	var err error

	err = createTopics(kafkaBroker, TopicIngredients, TopicRecipes, TopicIngredientsReq, TopicNutritionFacts)
	suite.Require().NoError(err)

	suite.controller, err = NewController(service.NewService(), kafkaBroker)
	suite.Require().NoError(err)

	suite.ingredientsReqReader, err = newReader([]string{kafkaBroker}, "nutrition-facts-test-ingredients-req", TopicIngredientsReq)
	suite.Require().NoError(err)

	suite.nutritionFactsReader, err = newReader([]string{kafkaBroker}, "nutrition-facts-test-facts", TopicNutritionFacts)
	suite.Require().NoError(err)

	suite.recipesWriter = newWriter([]string{kafkaBroker}, TopicRecipes)
	suite.ingredientsWriter = newWriter([]string{kafkaBroker}, TopicIngredients)

	suite.rand = random.NewRandomGenerator()

	go func() {
		err := suite.controller.Run(context.Background())
		if err != nil {
			suite.Require().ErrorIs(err, io.EOF)
		}
	}()
}

func (suite *ControllerTestSuite) TearDownSuite() {
	err := closeAll(suite.ingredientsReqReader, suite.nutritionFactsReader, suite.recipesWriter, suite.ingredientsWriter)
	suite.Assert().NoError(err)

	err = suite.controller.Stop()
	suite.Assert().NoError(err)
}

func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
