package ingredient

import (
	"context"
)

type Service interface {
	Create(ctx context.Context, dto CreateIngredientDTO) (string, error)
	GetById(ctx context.Context, id string) (Ingredient, error)
	SearchByName(ctx context.Context, nameQuery string) ([]Ingredient, error)
}
