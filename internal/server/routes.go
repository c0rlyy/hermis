package server

import (
	"net/http"

	"github.com/c0rlyy/hermis/internal/database"
	"github.com/c0rlyy/hermis/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	// User routes no JWT required for creating and reading users
	userGroup := e.Group("/api/user")
	userGroup.GET("/:username", s.readUser)
	userGroup.GET("", s.readUsers)

	// Restricted routes (Require JWT Auth)
	restrictedGroup := e.Group("/api")
	restrictedGroup.Use(echojwt.WithConfig(jwtConfig))
	restrictedGroup.GET("/auth/me", s.getCurrentUser)

	return e
}

func (s *Server) createUser(c echo.Context) error {
	var req CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	if exists, err := s.userExists(req.Username, req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	} else if exists {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "username or email already assigned to user"})
	}

	hashPass, err := utils.GenerateFromPassword(req.Password, &s.hashParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	user := database.NewUser(req.Email, hashPass, req.Username)
	result, err := database.WriteOne(database.UsersCollection, user, s.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	userID, _ := result.InsertedID.(primitive.ObjectID)

	token, err := utils.CreateJwtString(user.Username, userID, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	userResponse := CreatedUserResponse{
		Id:       userID,
		Username: req.Username,
		Token:    token,
	}

	return c.JSON(http.StatusOK, userResponse)
}

// TODO make this work better dude
func (s *Server) readUser(c echo.Context) error {
	username := c.Param("username")
	filter := bson.D{{Key: "username", Value: username}}
	var user database.UserModel
	err := database.ReadOne(filter, database.UsersCollection, &user, s.db)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error retrieving user"})
	}

	return c.JSON(http.StatusOK, user)
}

func (s *Server) readUsers(c echo.Context) error {
	var users []database.UserModel
	err := database.ReadAll(database.UsersCollection, &users, s.db, bson.D{})
	if err != nil {

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error retrieving user"})
	}

	return c.JSON(http.StatusOK, users)
}

// Filters in this fcuntions are broken do not understand why
func (s *Server) userExists(username, email string) (bool, error) {
	filter := bson.D{
		{Key: "username", Value: username},
	}
	filter2 := bson.D{
		{Key: "email", Value: email},
	}

	var foundUsername []database.UserModel

	err := database.ReadAll(database.UsersCollection, &foundUsername, s.db, filter)
	if err != nil {
		return false, err
	}
	var foundEmial []database.UserModel
	err = database.ReadAll(database.UsersCollection, &foundEmial, s.db, filter2)
	if err != nil {
		return false, nil
	}

	if len(foundUsername) == 0 && len(foundEmial) == 0 {
		return false, nil
	}

	return true, nil
}

func (s *Server) login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	var user database.UserModel

	err := database.ReadOne(bson.D{{Key: "email", Value: email}}, database.UsersCollection, &user, s.db)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Incorect user credentials"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	isPasswordCorrect, err := utils.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	if !isPasswordCorrect {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Incorect user credentials"})
	}

	token, err := utils.CreateJwtString(user.Username, user.ID, s.cfg.SecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func (s *Server) getCurrentUser(c echo.Context) error {
	claims := utils.DecodeJwt(c)
	return c.JSON(http.StatusOK, claims)
}

func (s *Server) refreshToken(c echo.Context) error {

	return c.JSON(http.StatusOK, "tokenssss")
}
