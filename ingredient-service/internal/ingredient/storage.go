package ingredient

import "context"

type Storage interface {
	Create(ctx context.Context, ingredient Ingredient) (string, error)
	FindById(ctx context.Context, id string) (Ingredient, error)
	SearchByName(ctx context.Context, query string, fillNutritionInfo bool) ([]Ingredient, error)
}
