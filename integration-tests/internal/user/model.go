package user

import "time"

type User struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	Email        string    `json:"email" bson:"email,omitempty"`
	Password     string    `json:"-" bson:"password,omitempty"`
	RegisteredAt time.Time `json:"registered_at" bson:"registered_at,omitempty"`
}

type CreateUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginDTO = CreateUserDTO

type UserRegistrationDTO struct {
	ID    string `json:"user_id,omitempty"`
	Email string `json:"email"`
	Error string `json:"error,omitempty"`
}

type UserLoginDTO struct {
	User  User   `json:"user,omitempty"`
	Email string `json:"email"`
	Error string `json:"error,omitempty"`
}
