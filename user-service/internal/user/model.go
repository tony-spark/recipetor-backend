package user

import "time"

type User struct {
	Id           string    `json:"id" bson:"_id,omitempty"`
	Email        string    `json:"email" bson:"email,omitempty"`
	Password     string    `json:"-" bson:"password,omitempty"`
	RegisteredAt time.Time `json:"registered_at" bson:"registered_at,omitempty"`
}

type CreateUserDTO struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}
