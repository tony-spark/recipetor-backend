package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func TestIngredientsService(t *testing.T) {
	suite.Run(t, new(IngredientServiceSuite))
}

func TestRecipeService(t *testing.T) {
	suite.Run(t, new(RecipeServiceSuite))
}
