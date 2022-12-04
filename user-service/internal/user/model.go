package user

import "time"

type User struct {
	Id           string    `json:"id" bson:"_id"`
	Email        string    `json:"email" bson:"email"`
	Password     string    `json:"-" bson:"password"`
	RegisteredAt time.Time `json:"registered_at" bson:"registered_at"`
}
