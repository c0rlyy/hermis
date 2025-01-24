package database

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"` // ObjectID for the unique identifier better compabilty
	Email    string             `bson:"email"`
	Password string             `bson:"password"`
	Username string             `bson:"username"`
	SchoolID int                `bson:"school_id"`
}

type RefreshTokenModel struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserId       primitive.ObjectID `bson:"user_id,omitempty"`
	RefreshToken string             `bson:"refresh_token"`
}

func NewRefreshToken(userId primitive.ObjectID, refreshToken string) *RefreshTokenModel {
	return &RefreshTokenModel{
		UserId:       userId,
		RefreshToken: refreshToken,
	}
}

func NewUser(email, password, username string) *UserModel {
	return &UserModel{
		Email:    email,
		Password: password,
		Username: username,
	}

}

// type Model interface {
// 	GetId() primitive.ObjectID
// }

// func (u *UserModel) GetId() primitive.ObjectID {
// 	return u.ID
// }
