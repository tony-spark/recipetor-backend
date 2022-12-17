package service

import (
	"reflect"
	"testing"

	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
)

func Test_service_CalcRecipeNutritions(t *testing.T) {
	tests := []struct {
		name       string
		recipeDTO  nutrition.RecipeDTO
		wantResult nutrition.RecipeNutritionsDTO
		wantErr    bool
	}{
		{
			name: "unsifficient data",
			recipeDTO: nutrition.RecipeDTO{
				RecipeID: "1",
				Ingredients: []nutrition.RecipeIngredient{
					{
						Ingredient: nutrition.Ingredient{
							ID:             "1",
							BaseUnit:       "г",
							NutritionFacts: nil,
						},
						Unit:   "г",
						Amount: 10,
					},
					{
						Ingredient: nutrition.Ingredient{
							ID:             "2",
							BaseUnit:       "мл",
							NutritionFacts: nil,
						},
						Unit:   "мл",
						Amount: 1000,
					},
				},
			},
			wantResult: nutrition.RecipeNutritionsDTO{},
			wantErr:    true,
		},
		{
			name: "inaccurate result",
			recipeDTO: nutrition.RecipeDTO{
				RecipeID: "1",
				Ingredients: []nutrition.RecipeIngredient{
					{
						Ingredient: nutrition.Ingredient{
							ID:             "1",
							BaseUnit:       "шт",
							NutritionFacts: nil,
						},
						Unit:   "шт",
						Amount: 1,
					}, {
						Ingredient: nutrition.Ingredient{
							ID:       "2",
							BaseUnit: "г",
							NutritionFacts: &nutrition.NutritionFacts{
								Calories:      1,
								Proteins:      0.5,
								Fats:          0.3,
								Carbohydrates: 0.1,
							},
						},
						Unit:   "г",
						Amount: 15,
					}, {
						Ingredient: nutrition.Ingredient{
							ID:       "3",
							BaseUnit: "мл",
							NutritionFacts: &nutrition.NutritionFacts{
								Calories:      0.5,
								Proteins:      0,
								Fats:          0.8,
								Carbohydrates: 0.4,
							},
						},
						Unit:   "мл",
						Amount: 65,
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
			recipeDTO: nutrition.RecipeDTO{
				RecipeID: "1",
				Ingredients: []nutrition.RecipeIngredient{
					{
						Ingredient: nutrition.Ingredient{
							ID:       "1",
							BaseUnit: "шт",
							NutritionFacts: &nutrition.NutritionFacts{
								Calories:      50,
								Proteins:      5,
								Fats:          10,
								Carbohydrates: 15,
							},
						},
						Unit:   "шт",
						Amount: 1,
					}, {
						Ingredient: nutrition.Ingredient{
							ID:       "2",
							BaseUnit: "г",
							NutritionFacts: &nutrition.NutritionFacts{
								Calories:      1,
								Proteins:      0.5,
								Fats:          0.3,
								Carbohydrates: 0.1,
							},
						},
						Unit:   "г",
						Amount: 15,
					}, {
						Ingredient: nutrition.Ingredient{
							ID:       "3",
							BaseUnit: "мл",
							NutritionFacts: &nutrition.NutritionFacts{
								Calories:      0.5,
								Proteins:      0,
								Fats:          0.8,
								Carbohydrates: 0.4,
							},
						},
						Unit:   "мл",
						Amount: 65,
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
			gotResult, err := s.CalcRecipeNutritions(tt.recipeDTO)
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
