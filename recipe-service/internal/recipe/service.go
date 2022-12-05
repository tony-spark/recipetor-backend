package recipe

import "context"

type Service interface {
	Create(ctx context.Context, dto CreateRecipeDTO) (string, error)
	Update(ctx context.Context, dto UpdateRecipeDTO) error
	GetById(ctx context.Context, id string) (Recipe, error)
	GetAllByUser(ctx context.Context, userId string) ([]Recipe, error)
	GetAll(ctx context.Context, start int, limit int) ([]Recipe, error)
	FindByIngredients(ctx context.Context, ingredientIds []string) ([]Recipe, error)
}
