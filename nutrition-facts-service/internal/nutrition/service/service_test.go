package service

import (
	"reflect"
	"testing"

	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
)

func Test_service_CalcRecipeNutritions(t *testing.T) {
	tests := []struct {
		name        string
		recipe      nutrition.Recipe
		ingredients map[string]nutrition.Ingredient
		wantResult  nutrition.RecipeNutritionsDTO
		wantErr     bool
	}{
		{
			name: "unsifficient data",
			ingredients: map[string]nutrition.Ingredient{
				"1": {
					ID:             "1",
					BaseUnit:       "г",
					NutritionFacts: nil,
				},
				"2": {
					ID:             "2",
					BaseUnit:       "мл",
					NutritionFacts: nil,
				},
			},
			recipe: nutrition.Recipe{
				ID: "1",
				Ingredients: []nutrition.RecipeIngredient{
					{
						IngredientID: "1",
						Unit:         "г",
						Amount:       10,
					},
					{
						IngredientID: "2",
						Unit:         "мл",
						Amount:       1000,
					},
				},
			},
			wantResult: nutrition.RecipeNutritionsDTO{},
			wantErr:    true,
		},
		{
			name: "inaccurate result",
			ingredients: map[string]nutrition.Ingredient{
				"1": {
					ID:             "1",
					BaseUnit:       "шт",
					NutritionFacts: nil,
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
			},
			recipe: nutrition.Recipe{
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
			wantResult: nutrition.RecipeNutritionsDTO{
				RecipeID: "1",
				NutritionFacts: nutrition.NutritionFacts{
					Calories:      15*1 + 65*0.5,
					Proteins:      15 * 0.5,
					Fats:          15*0.3 + 65*0.8,
					Carbohydrates: 15*0.1 + 65*0.4,
				},
				Inaccurate: true,
			},
			wantErr: false,
		},
		{
			name: "full result",
			ingredients: map[string]nutrition.Ingredient{
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
			},
			recipe: nutrition.Recipe{
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
			wantResult: nutrition.RecipeNutritionsDTO{
				RecipeID: "1",
				NutritionFacts: nutrition.NutritionFacts{
					Calories:      15*1 + 65*0.5 + 50,
					Proteins:      15*0.5 + 5,
					Fats:          15*0.3 + 65*0.8 + 10,
					Carbohydrates: 15*0.1 + 65*0.4 + 15,
				},
				Inaccurate: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService()
			gotResult, err := s.CalcRecipeNutritions(tt.recipe, tt.ingredients)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcRecipeNutritions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("CalcRecipeNutritions() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
