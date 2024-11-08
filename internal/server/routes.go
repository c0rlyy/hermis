package server

import (
	"net/http"

	"github.com/c0rlyy/hermis/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(utils.JwtCustomClaims)
		},
		SigningKey: []byte(s.cfg.SecretKey),
	}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `time=${time_rfc3339} method=${method}, uri=${uri}, status=${status}, ip = ${remote_ip}` + "\n",
	}))
	e.Use(middleware.Recover())

	e.POST("/api/login", s.login)
	e.POST("/api/register", s.createUser)
	e.GET("/api/test", s.getSchoolId)
	// e.GET("/api/rt", s.refreshToken)

	// User routes no JWT required for creating and reading users
	userGroup := e.Group("/api/users")
	userGroup.GET("", s.readUsers)
	userGroup.GET("/:username", s.readUser)

	// Restricted routes (Require JWT Auth)
	restrictedGroup := e.Group("/api")
	restrictedGroup.Use(echojwt.WithConfig(jwtConfig))
	restrictedGroup.GET("/auth/me", s.getCurrentUser)

	return e
}
