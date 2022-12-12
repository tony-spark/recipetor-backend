package storage

import (
	"context"

	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
)

type Storage interface {
	Create(ctx context.Context, recipe recipe.Recipe) (string, error)
	GetById(ctx context.Context, id string) (recipe.Recipe, error)
	GetAll(ctx context.Context, start int64, limit int64) ([]recipe.Recipe, error)
	GetAllByUser(ctx context.Context, userID string) ([]recipe.Recipe, error)
	Update(ctx context.Context, recipe recipe.Recipe) error
	Delete(ctx context.Context, id string) error
}
