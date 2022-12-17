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

type UserRegistrationDTO struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Error string `json:"error,omitempty"`
}
