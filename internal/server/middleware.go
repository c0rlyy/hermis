package server

import (
	"net/http"

	"github.com/c0rlyy/hermis/internal/utils"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func RefreshTokenMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Retrieve refresh token from cookie (or another source, like headers)
			rtCookie, err := c.Cookie("refreshToken")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid refresh token")
			}

			refreshToken := rtCookie.Value

			// Parse and validate the refresh token
			token, err := jwt.ParseWithClaims(refreshToken, &utils.JwtRefreshToken{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token")
			}

			c.Set("refreshToken", token)

			return next(c)
		}
	}
}
