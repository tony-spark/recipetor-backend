package service

import (
	"fmt"

	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
)

type Service interface {
	CalcRecipeNutritions(recipe nutrition.Recipe, ingredients map[string]nutrition.Ingredient) (nutrition.RecipeNutritionsDTO, error)
}

type service struct {
	failThreshold       float64
	inaccurateThreshold float64
}

func NewService() Service {
	return service{
		// TODO: move thresholds to config eventually
		failThreshold:       0.5,
		inaccurateThreshold: 0.2,
	}
}

func (s service) CalcRecipeNutritions(recipe nutrition.Recipe, ingredients map[string]nutrition.Ingredient) (result nutrition.RecipeNutritionsDTO, err error) {
	facts := nutrition.NutritionFacts{}
	inaccurate := false

	unknown := 0.0
	for _, ing := range recipe.Ingredients {
		ingredient, ok := ingredients[ing.IngredientID]
		if ok && ingredient.NutritionFacts != nil && ing.Unit == ingredient.BaseUnit { // TODO: unit conversion
			facts.Calories += ing.Amount * ingredient.NutritionFacts.Calories
			facts.Fats += ing.Amount * ingredient.NutritionFacts.Fats
			facts.Carbohydrates += ing.Amount * ingredient.NutritionFacts.Carbohydrates
			facts.Proteins += ing.Amount * ingredient.NutritionFacts.Proteins
		} else {
			unknown += 1
		}
	}

	rate := unknown / float64(len(recipe.Ingredients))
	if rate > s.failThreshold {
		return result, fmt.Errorf("could not calclulate nutrition facts: unsufficient data")
	}
	if rate > s.inaccurateThreshold {
		inaccurate = true
	}

	result = nutrition.RecipeNutritionsDTO{
		RecipeID:       recipe.ID,
		NutritionFacts: facts,
		Inaccurate:     inaccurate,
	}

	return
}
