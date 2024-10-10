package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JwtCustomClaims struct {
	Username string             `json:"username"`
	Id       primitive.ObjectID `json:"id"`
	Admin    bool               `json:"admin"`
	jwt.RegisteredClaims
}

type JwtRefreshToken struct {
	Id primitive.ObjectID `json:"id"`
	jwt.RegisteredClaims
}

func NewJwtCustomClaims(username string, id primitive.ObjectID) JwtCustomClaims {
	return JwtCustomClaims{
		Username: username,
		Id:       id,
		Admin:    false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func NewJwtRefreshToken(id primitive.ObjectID) JwtRefreshToken {
	return JwtRefreshToken{
		Id: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 35)),
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
func DecodeJwt(c echo.Context) *JwtCustomClaims {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*JwtCustomClaims)
	return claims
}

// make this generic in future TODO
func EncodeRefreshToken(claims *JwtRefreshToken, secretKey string) (string, error) {
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rt.SignedString([]byte(secretKey))
}

func CreateRefreshTokenString(id primitive.ObjectID, secretKey string) (string, error) {
	rtClaims := NewJwtRefreshToken(id)
	return EncodeRefreshToken(&rtClaims, secretKey)
}

func (c JwtRefreshToken) Valid() error {
	vErr := new(ValidationError)
	now := TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if !c.VerifyExpiresAt(now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= ValidationErrorExpired
	}

	if !c.VerifyIssuedAt(now, false) {
		vErr.Inner = fmt.Errorf("Token used before issued")
		vErr.Errors |= ValidationErrorIssuedAt
	}

	if !c.VerifyNotBefore(now, false) {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= ValidationErrorNotValidYet
	}

	if vErr.valid() {
		return nil
	}

	return vErr
}
