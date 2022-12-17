package service

import (
	"fmt"

	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
)

type Service interface {
	CalcRecipeNutritions(recipe nutrition.RecipeDTO) (nutrition.RecipeNutritionsDTO, error)
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

func (s service) CalcRecipeNutritions(recipe nutrition.RecipeDTO) (result nutrition.RecipeNutritionsDTO, err error) {
	facts := nutrition.NutritionFacts{}
	inaccurate := false

	unknown := 0.0
	for _, ing := range recipe.Ingredients {
		if ing.Ingredient.NutritionFacts != nil && ing.Unit == ing.Ingredient.BaseUnit { // TODO: unit conversion
			facts.Calories += ing.Amount * ing.Ingredient.NutritionFacts.Calories
			facts.Fats += ing.Amount * ing.Ingredient.NutritionFacts.Fats
			facts.Carbohydrates += ing.Amount * ing.Ingredient.NutritionFacts.Carbohydrates
			facts.Proteins += ing.Amount * ing.Ingredient.NutritionFacts.Proteins
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
		RecipeID:       recipe.RecipeID,
		NutritionFacts: facts,
		Inaccurate:     inaccurate,
	}

	return
}
