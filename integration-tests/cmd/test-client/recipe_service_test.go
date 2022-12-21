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

type RecipeServiceSuite struct {
	suite.Suite

	newRecipeWriter      *kafka.Writer
	reqRecipeWriter      *kafka.Writer
	nutritionFactsWriter *kafka.Writer
	recipesReader        *kafka.Reader

	rand random.Generator
}

func (suite *RecipeServiceSuite) SetupSuite() {
	suite.Require().NotEmpty(flagKafkaBroker, "--kafka-broker flag required")

	createTopics(flagKafkaBroker, TopicRecipesNew, TopicRecipesReq, TopicRecipes, TopicNutritionFacts)

	suite.newRecipeWriter = newWriter([]string{flagKafkaBroker}, TopicRecipesNew)
	suite.recipesReader = newReader([]string{flagKafkaBroker}, TopicRecipes)
	suite.reqRecipeWriter = newWriter([]string{flagKafkaBroker}, TopicRecipesReq)
	suite.nutritionFactsWriter = newWriter([]string{flagKafkaBroker}, TopicNutritionFacts)

	suite.rand = random.NewRandomGenerator()
}

func (suite *RecipeServiceSuite) TearDownSuite() {
	err := closeAll(suite.newRecipeWriter, suite.reqRecipeWriter, suite.recipesReader, suite.nutritionFactsWriter)
	assert.NoError(suite.T(), err)
}

func (suite *RecipeServiceSuite) TestRecipeService() {
	suite.Run("create recipe send nutrition facts and get by id", func() {
		newRecipeDTO := suite.randomCreateRecipe()
		write(suite.T(), suite.newRecipeWriter, newRecipeDTO.Name, newRecipeDTO)

		var createdDTO RecipeDTO
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

		recipeNutritionsDTO := RecipeNutritionsDTO{
			RecipeID:       createdDTO.ID,
			NutritionFacts: suite.randomNutritionFacts(),
			Inaccurate:     false,
		}
		write(suite.T(), suite.nutritionFactsWriter, recipeNutritionsDTO.RecipeID, recipeNutritionsDTO)

		findRecipeDTO := FindRecipeDTO{
			ID: createdDTO.ID,
		}
		write(suite.T(), suite.reqRecipeWriter, findRecipeDTO.ID, findRecipeDTO)

		var gotDTO RecipeDTO
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

func (suite *RecipeServiceSuite) randomCreateRecipe() CreateRecipeDTO {
	return CreateRecipeDTO{
		Name:      suite.rand.RandomString(8),
		CreatedBy: suite.rand.RandomObjectID(),
		Ingredients: []RecipeIngredient{
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
		Steps: []Step{
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
			{Description: suite.rand.RandomString(200)},
		},
	}
}

func (suite *RecipeServiceSuite) randomNutritionFacts() NutritionFacts {
	return NutritionFacts{
		Calories:      suite.rand.RandomFloat(1000),
		Proteins:      suite.rand.RandomFloat(100),
		Fats:          suite.rand.RandomFloat(100),
		Carbohydrates: suite.rand.RandomFloat(100),
	}
}

type Step struct {
	Description string `json:"description" bson:"description"`
}

type RecipeIngredient struct {
	IngredientID string  `json:"ingredient_id" bson:"ingredient_id"`
	Unit         string  `json:"unit" bson:"unit"`
	Amount       float64 `json:"amount" bson:"amount"`
}

type Recipe struct {
	ID             string             `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name,omitempty"`
	CreatedBy      string             `json:"created_by" bson:"created_by,omitempty"`
	Ingredients    []RecipeIngredient `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Steps          []Step             `json:"steps,omitempty" bson:"steps,omitempty"`
	NutritionFacts *NutritionFacts    `json:"nutrition_facts,omitempty" bson:"nutrition_facts,omitempty"`
}

type CreateRecipeDTO struct {
	Name        string             `json:"name"`
	CreatedBy   string             `json:"created_by"`
	Ingredients []RecipeIngredient `json:"ingredients,omitempty"`
	Steps       []Step             `json:"steps,omitempty"`
}

type FindRecipeDTO struct {
	ID            string   `json:"recipe_id"`
	UserID        string   `json:"user_id"`
	IngredientIDs []string `json:"ingredient_ids"`
}

type RecipeDTO struct {
	Recipe Recipe `json:"recipe,omitempty"`
	ID     string `json:"recipe_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type RecipeNutritionsDTO struct {
	RecipeID       string         `json:"recipe_id"`
	NutritionFacts NutritionFacts `json:"nutrition_facts"`
	Inaccurate     bool           `json:"is_inaccurate"`
}
