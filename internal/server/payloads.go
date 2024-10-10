package server

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreatedUserResponse struct {
	Id       primitive.ObjectID `json:"id"`
	Username string             `json:"username"`
	Token    string             `json:"token"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12"`
}

type GetUserRequest struct {
	Username string `json:"username" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}
