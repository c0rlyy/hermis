package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JwtCustomClaims struct {
	Username string             `json:"username"`
	Sub      primitive.ObjectID `json:"sub"`
	Admin    bool               `json:"admin"`
	jwt.RegisteredClaims
}

type JwtRefreshToken struct {
	Sub primitive.ObjectID `json:"sub"`
	jwt.RegisteredClaims
}

func NewJwtCustomClaims(username string, sub primitive.ObjectID) JwtCustomClaims {
	return JwtCustomClaims{
		Username: username,
		Sub:      sub,
		Admin:    false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func NewJwtRefreshToken(sub primitive.ObjectID) JwtRefreshToken {
	return JwtRefreshToken{
		Sub: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			// 7 days
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, 7)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func EncodeJwt(claims *JwtCustomClaims, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func CreateJwtString(username string, id primitive.ObjectID, secretKey string) (string, error) {
	claims := NewJwtCustomClaims(username, id)
	return EncodeJwt(&claims, secretKey)
}

// decodes token from auth header. Automaticly returns correct error when jwt is malformed or missing
func DecodeJwt(secretKey, authToken string) (*JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(authToken, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired Refresh token")
	}

	authT, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		return nil, errors.New("invalid token structure")
	}
	return authT, nil
}

// make this generic in future TODO
func EncodeRefreshToken(claims *JwtRefreshToken, secretKey string) (string, error) {
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rt.SignedString([]byte(secretKey))
}

func CreateRefreshTokenString(sub primitive.ObjectID, secretKey string) (string, error) {
	rtClaims := NewJwtRefreshToken(sub)
	return EncodeRefreshToken(&rtClaims, secretKey)
}

func DecodeRefreshToken(secretKey, rt string) (*JwtRefreshToken, error) {
	token, err := jwt.ParseWithClaims(rt, &JwtRefreshToken{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired Refresh token")
	}

	refreshToken, ok := token.Claims.(*JwtRefreshToken)
	if !ok {
		return nil, errors.New("invalid token structure")
	}
	return refreshToken, nil
}
