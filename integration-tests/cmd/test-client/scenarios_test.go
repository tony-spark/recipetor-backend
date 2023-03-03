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
	"github.com/tony-spark/recipetor-backend/integration-tests/internal/recipe"
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
	newIngredientWriter *kafka.Writer
	ingredientsReader   *kafka.Reader

	// Recipe service
	recipesReader   *kafka.Reader
	newRecipeWriter *kafka.Writer
	reqRecipeWriter *kafka.Writer

	rand random.Generator
}

func (suite *ScenariosTestSuite) TestScenarios() {
	suite.Run("scenario 1", func() {
		corID := generateCorrelationID()

		var userID string
		var ingredientIDs []string
		var recipeID string

		// 1. User registration
		suite.T().Log("1. User registration")
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
				suite.T().Logf("User registered, id = %s", userID)
				break
			}
		}

		// 2. User login
		suite.T().Log("2. User login")
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
				suite.T().Log("user logged in")
				break
			}
		}

		// 3. Add ingredients
		suite.T().Log("3. Add ingredients")
		createIngredientDTOs := []ingredient.CreateIngredientDTO{
			suite.createIngredient("пшеничная мука", "г", 3.64, 0.1033, 0.0098, 0.7631),
			suite.createIngredient("кефир", "мл", 0.41, 0.0379, 0.0093, 0.0448),
			suite.createIngredient("яйцо", "шт", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("сахар", "г", 4, 0.0, 0.0, 0.9998),
			suite.createIngredient("соль", "г", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("сода", "г", 0.0, 0.0, 0.0, 0.0),
			suite.createIngredient("растительное масло", "мл", 0.0, 0.0, 0.0, 0.0),
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
				suite.T().Logf("ingredient added id = %s", createdDTO.ID)
				break
			}
		}
		suite.Require().Equal(len(createIngredientDTOs), len(ingredientIDs))

		// 4. Add recipe
		suite.T().Log("4. Add recipe")
		createRecipeDTO := recipe.CreateRecipeDTO{
			Name:      "Пышные оладушки на кефире." + suite.clientID,
			CreatedBy: userID,
			Ingredients: []recipe.RecipeIngredient{
				{IngredientID: ingredientIDs[0], Unit: "г", Amount: 240},
				{IngredientID: ingredientIDs[1], Unit: "мл", Amount: 250},
				{IngredientID: ingredientIDs[2], Unit: "шт", Amount: 1},
				{IngredientID: ingredientIDs[3], Unit: "г", Amount: 10},
				{IngredientID: ingredientIDs[4], Unit: "г", Amount: 1},
				{IngredientID: ingredientIDs[5], Unit: "г", Amount: 3},
				{IngredientID: ingredientIDs[6], Unit: "мл", Amount: 1},
			},
			Steps: []recipe.Step{
				{Description: "Смешиваем яйцо с сахаром и солью"},
				{Description: "В микроволновке или на плите слегка подогреваем кефир. Должна быть видна сыворотка, но кефир не должен пойти хлопьями"},
				{Description: "В яичную смесь, постоянно помешивая, добавляем тёплый кефир"},
				{Description: "В жидкую смесь небольшими порциями (по 2 столовые ложки с горкой), добавляем муку, замешиваем густое тесто.\n(Тесто не должно свободно скользить с ложки, должно лениво сползать.\nЕсли жидковато - добавить муки, если сильно густое - кефира, но очень по-малу, чтобы уловить нужную консистенцию)."},
				{Description: "Добавляем соду. Тщательно, но быстро вмешиваем в тесто.\n(Сода сделает тесто чуть жиже, не пугайтесь)"},
				{Description: "Даём тесту постоять 10-15 минут\nС этого момента тесто не перемешиваем!"},
				{Description: "На разогретую сковородку добавляем масло, ждём пока оно разогреется"},
				{Description: "Аккуратно набираем тесто столовой ложкой !по краю миски! (в идеале окунуть ложку в масло, чтобы тесто лучше сползало)"},
				{Description: "Обжариваем с двух сторон на среднем огне (примерно по 3 минуты)"},
			},
		}
		write(suite.T(), suite.newRecipeWriter, createRecipeDTO.Name, createRecipeDTO, corID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.recipesReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				if !checkCorrelationID(message, corID) {
					continue
				}

				var createdRecipeDTO recipe.RecipeDTO
				err = json.Unmarshal(message.Value, &createdRecipeDTO)
				require.NoError(suite.T(), err, "ошибка при раскодировании сообщения")
				assert.Empty(suite.T(), createdRecipeDTO.Error)
				assert.NotEmpty(suite.T(), createdRecipeDTO.ID)
				assert.NotEmpty(suite.T(), createdRecipeDTO.Recipe)
				assert.Equal(suite.T(), createRecipeDTO.Name, createdRecipeDTO.Recipe.Name)
				assert.Equal(suite.T(), createRecipeDTO.Ingredients, createdRecipeDTO.Recipe.Ingredients)
				assert.Equal(suite.T(), createRecipeDTO.Steps, createdRecipeDTO.Recipe.Steps)

				recipeID = createdRecipeDTO.ID
				suite.T().Logf("recipe added id = %s", recipeID)
				break
			}
		}

		// 5. Wait and check recipe's nutrition facts is calculated
		suite.T().Log("5. Wait and check recipe's nutrition facts is calculated")
		time.Sleep(30 * time.Second)

		findRecipeDTO := recipe.FindRecipeDTO{
			ID: recipeID,
		}
		write(suite.T(), suite.reqRecipeWriter, findRecipeDTO.ID, findRecipeDTO, corID)
		{
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for {
				message, err := suite.recipesReader.ReadMessage(ctx)
				require.NoError(suite.T(), err, "ошибка при чтении сообщения")
				if !checkCorrelationID(message, corID) {
					continue
				}

				var gotRecipeDTO recipe.RecipeDTO
				err = json.Unmarshal(message.Value, &gotRecipeDTO)
				require.NoError(suite.T(), err)
				assert.Empty(suite.T(), gotRecipeDTO.Error)
				assert.Equal(suite.T(), findRecipeDTO.ID, gotRecipeDTO.ID)
				assert.NotEmpty(suite.T(), gotRecipeDTO.Recipe)
				assert.NotEmpty(suite.T(), gotRecipeDTO.Recipe.NutritionFacts)
				suite.T().Logf("got recipe's nutrition facts: %+v", gotRecipeDTO.Recipe.NutritionFacts)
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

	suite.ingredientsReader, err = newReader([]string{flagKafkaBroker}, suite.clientID, TopicIngredients)
	suite.Require().NoError(err)

	suite.newIngredientWriter = newWriter([]string{flagKafkaBroker}, TopicIngredientsNew)

	suite.recipesReader, err = newReader([]string{flagKafkaBroker}, suite.clientID, TopicRecipes)
	suite.Require().NoError(err)

	suite.newRecipeWriter = newWriter([]string{flagKafkaBroker}, TopicRecipesNew)
	suite.reqRecipeWriter = newWriter([]string{flagKafkaBroker}, TopicRecipesReq)
}

func (suite *ScenariosTestSuite) TearDownSuite() {
	err := closeAll(suite.registrationsReader, suite.registrationsWriter, suite.loginsReader, suite.loginsWriter,
		suite.ingredientsReader, suite.newIngredientWriter,
		suite.newRecipeWriter, suite.recipesReader, suite.reqRecipeWriter)
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
