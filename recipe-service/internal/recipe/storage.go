package recipe

import "context"

type Storage interface {
	Create(ctx context.Context, recipe Recipe) (string, error)
	GetById(ctx context.Context, id string) (Recipe, error)
	GetAll(ctx context.Context, start int, limit int) ([]Recipe, error)
	GetAllByUser(ctx context.Context, userID string) ([]Recipe, error)
	Update(ctx context.Context, recipe Recipe) error
	Delete(ctx context.Context, id string) error
}
