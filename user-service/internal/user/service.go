package user

import "context"

type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (string, error)
	GetByEmailAndPassword(ctx context.Context, email string, password string) (User, error)
	GetById(ctx context.Context, id string) (User, error)
}

type CreateUserDTO struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}
