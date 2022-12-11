package recipe

import "context"

type Service interface {
	Create(ctx context.Context, dto CreateRecipeDTO) (string, error)
	Update(ctx context.Context, dto UpdateRecipeDTO) error
	GetByID(ctx context.Context, id string) (Recipe, error)
	GetAllByUser(ctx context.Context, userID string) ([]Recipe, error)
	GetAll(ctx context.Context, start int, limit int) ([]Recipe, error)
	FindByIngredients(ctx context.Context, ingredientIDs []string) ([]Recipe, error)
}
