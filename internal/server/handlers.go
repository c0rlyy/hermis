package server

import (
	"net/http"

	"github.com/c0rlyy/hermis/internal/broker"
	"github.com/c0rlyy/hermis/internal/database"
	"github.com/c0rlyy/hermis/internal/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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
	user, err := database.FindByUsername(username, s.db)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error retrieving user"})
	}
	return c.JSON(http.StatusOK, user)
}

func (s *Server) readUsers(c echo.Context) error {
	users, err := database.GetAll[database.UserModel](database.UsersCollection, s.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error retrieving user"})
	}
	return c.JSON(http.StatusOK, users)
}

// TODO REFACTOR THIS
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
	if len(email) == 0 || len(password) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Incorect or malformed payload"})
	}

	user, err := database.FindByEmail(email, s.db)
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
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
	})
}

func (s *Server) getCurrentUser(c echo.Context) error {
	claims := utils.DecodeJwt(c)
	return c.JSON(http.StatusOK, claims)
}

// MAKE THIS SENSABLE
func (s *Server) getSchoolId(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	claims := utils.DecodeJwt(c)

	if len(username) == 0 || len(password) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Incorect or malformed payload"})
	}

	topic := claims.Sub.Hex()
	s.mb.AddTopic(topic)
	s.mb.PushEvent(topic, broker.Event{Message: "test", Type: "adding"})

	// go client.Execute(username, password, s.mb,topic)

	return c.JSON(http.StatusOK, "yeah yeah")
}

// func (s *Server) refreshToken(c echo.Context) error {
// 	claims := utils.DecodeJwt(c)
// 	userId := claims.Id
// 	user, err := database.FindById[database.UserModel](database.UsersCollection, userId, s.db)

// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"errro": "internal server errro"})
// 	}

// 	rt, err := database.FindById[database.RefreshTokenModel]()
// 	refreshToken := database.NewRefreshToken(user.ID, "asdjkasldkjaslkdjaslkdjsalkdjlkadjs")
// 	_, err := database.WriteOne(database.RefreshTokenCollection, refreshToken, s.db)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Server internal error"})
// 	}
// 	var rtbuf []database.RefreshTokenModel
// 	_ = database.ReadAll(database.RefreshTokenCollection, &rtbuf, s.db, bson.D{})

// 	return c.JSON(http.StatusOK, echo.Map{"result": rtbuf})
// }

// func (s *Server)
