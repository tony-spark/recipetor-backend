package user

import "context"

type Storage interface {
	Create(ctx context.Context, user User) (string, error)
	FindById(ctx context.Context, id string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
}
