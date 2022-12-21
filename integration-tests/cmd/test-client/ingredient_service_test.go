package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/random"
	"time"
)

type IngredientServiceSuite struct {
	suite.Suite

	newIngredientWriter *kafka.Writer
	ingredientsReader   *kafka.Reader

	reqIngredientsWriter *kafka.Writer

	generator random.Generator
}

func (suite *IngredientServiceSuite) SetupSuite() {
	suite.Require().NotEmpty(flagKafkaBroker, "--kafka-broker flag required")

	createTopics(flagKafkaBroker, TopicIngredientsNew, TopicIngredients, TopicIngredientsReq)

	suite.newIngredientWriter = newWriter([]string{flagKafkaBroker}, TopicIngredientsNew)
	suite.ingredientsReader = newReader([]string{flagKafkaBroker}, TopicIngredients)
	suite.reqIngredientsWriter = newWriter([]string{flagKafkaBroker}, TopicIngredientsReq)
	suite.generator = random.NewRandomGenerator()
}

func (suite *IngredientServiceSuite) TearDownSuite() {
	err := closeAll(suite.reqIngredientsWriter, suite.newIngredientWriter, suite.ingredientsReader)
	assert.NoError(suite.T(), err)
}

func (suite *IngredientServiceSuite) TestIngredientService() {
	suite.Run("create ingredient and find by id", func() {
		newIngredientDTO := suite.randomCreateIngredient()
		write(suite.T(), suite.newIngredientWriter, newIngredientDTO.Name, newIngredientDTO)

		var createdDTO IngredientDTO
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.ingredientsReader.ReadMessage(ctx)
				if string(message.Key) != newIngredientDTO.Name {
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

		findDTO := FindIngredientsDTO{
			ID: createdDTO.ID,
		}
		write(suite.T(), suite.reqIngredientsWriter, findDTO.ID, findDTO)

		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.ingredientsReader.ReadMessage(ctx)
				if string(message.Key) != findDTO.ID {
					continue
				}
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")

				var gotDTO IngredientDTO
				err = json.Unmarshal(message.Value, &gotDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), gotDTO.Error)
				assert.Equal(suite.T(), createdDTO.Ingredient, gotDTO.Ingredient)
				break
			}
		}
	})
}

func (suite *IngredientServiceSuite) randomCreateIngredient() CreateIngredientDTO {
	return CreateIngredientDTO{
		Name:     suite.generator.RandomString(8),
		BaseUnit: suite.generator.RandomString(2),
		NutritionFacts: &NutritionFacts{
			Calories:      suite.generator.RandomFloat(500),
			Proteins:      suite.generator.RandomFloat(100),
			Fats:          suite.generator.RandomFloat(100),
			Carbohydrates: suite.generator.RandomFloat(100),
		},
	}
}

type Ingredient struct {
	ID             string          `json:"id" bson:"_id,omitempty"`
	Name           string          `json:"name" bson:"name,omitempty"`
	BaseUnit       string          `json:"base_unit" bson:"base_unit,omitempty"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type NutritionFacts struct {
	Calories      float64 `json:"calories" bson:"calories,omitempty"`
	Proteins      float64 `json:"proteins" bson:"proteins,omitempty"`
	Fats          float64 `json:"fats" bson:"fats,omitempty"`
	Carbohydrates float64 `json:"carbohydrates" bson:"carbohydrates,omitempty"`
}

type CreateIngredientDTO struct {
	Name           string          `json:"name"`
	BaseUnit       string          `json:"base_unit"`
	NutritionFacts *NutritionFacts `json:"nutrition_facts,omitempty"`
}

type IngredientDTO struct {
	Ingredient Ingredient `json:"ingredient,omitempty"`
	Name       string     `json:"name,omitempty"`
	ID         string     `json:"ingredient_id,omitempty"`
	NameQuery  string     `json:"name_query,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type FindIngredientsDTO struct {
	ID        string `json:"ingredient_id,omitempty"`
	NameQuery string `json:"name_query,omitempty"`
}
