package storage

import (
	"context"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
)

type Storage interface {
	Create(ctx context.Context, ingredient ingredient.Ingredient) (string, error)
	FindByID(ctx context.Context, id string) (ingredient.Ingredient, error)
	SearchByName(ctx context.Context, query string, fillNutritionInfo bool) ([]ingredient.Ingredient, error)
}
