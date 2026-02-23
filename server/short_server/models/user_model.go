package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID              bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID          string        `json:"user_id" bson:"user_id" validate:"required"`
	FirstName       string        `json:"first_name" bson:"first_name" validate:"required,min=2,max=100"`
	LastName        string        `json:"last_name" bson:"last_name" validate:"required,min=2,max=100"`
	Email           string        `json:"email" bson:"email" validate:"required,email"`
	Password        string        `json:"password" bson:"password" validate:"required,min=6"`
	Role            string        `json:"role" bson:"role" validate:"required,oneof=ADMIN USER"`
	CreatedAt       time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" bson:"updated_at"`
	Token           string        `json:"token" bson:"token"`
	RefreshToken    string        `json:"refresh_token" bson:"refresh_token"`
	FavouriteGenres []Genre       `json:"favourite_genres" bson:"favourite_genres" validate:"required,dive"`
}
