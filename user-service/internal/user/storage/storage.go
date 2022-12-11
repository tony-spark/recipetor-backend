package storage

import (
	"context"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
)

type Storage interface {
	Create(ctx context.Context, user user.User) (string, error)
	FindByID(ctx context.Context, id string) (user.User, error)
	FindByEmail(ctx context.Context, email string) (user.User, error)
}
