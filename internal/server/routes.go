package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	// jwtConfig := echojwt.Config{
	// 	NewClaimsFunc: func(c echo.Context) jwt.Claims {
	// 		return new(utils.JwtCustomClaims)
	// 	},
	// 	SigningKey: []byte(s.cfg.SecretKey),
	// 	// TokenLookup:"cookie:" ,
	// }

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `time=${time_rfc3339} method=${method}, uri=${uri}, status=${status}, ip = ${remote_ip}` + "\n",
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowCredentials: true,
	}))

	e.POST("/api/login", s.login)
	e.POST("/api/register", s.createUser)
	e.POST("/api/logout", s.logout)
	e.GET("/api/test", s.getDummyData)
	e.POST("/api/auth/refresh-token", s.refreshToken)

	userGroup := e.Group("/api/users")
	userGroup.GET("", s.readUsers)
	userGroup.GET("/:username", s.readUser)

	restrictedGroup := e.Group("/api")
	// restrictedGroup.Use(echojwt.WithConfig(jwtConfig)) // Apply JWT middleware for all routes under /api

	restrictedGroup.GET("/auth/me", s.getCurrentUser)

	return e
}
