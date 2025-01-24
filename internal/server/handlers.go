package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/c0rlyy/hermis/internal/client"
	"github.com/c0rlyy/hermis/internal/database"
	"github.com/c0rlyy/hermis/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) createUser(c echo.Context) error {

	email := c.FormValue("email")
	password := c.FormValue("password")
	username := c.FormValue("username")

	validate := validator.New()
	if err := validate.Struct(CreateUserRequest{Username: username, Email: email, Password: password}); err != nil {
		//TODO better error
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, len(validationErrors))
		for i, e := range validationErrors {
			field := e.Field()
			errorMessages[i] = field + " is incorrect try again. "
		}
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": errorMessages,
		})
	}

	if exists, err := s.userExists(username, email); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	} else if exists {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "username or email already assigned to user"})
	}

	hashPass, err := utils.GenerateFromPassword(password, &s.hashParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	user := database.NewUser(email, hashPass, username)
	result, err := database.WriteOne(database.UsersCollection, user, s.db)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	userID, _ := result.InsertedID.(primitive.ObjectID)

	token, err := utils.CreateJwtString(user.Username, userID, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	rtToken, err := utils.CreateRefreshTokenString(userID, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user"})
	}

	utils.SetAuthCookies(c, token, rtToken, "/api")
	userResponse := CreatedUserResponse{
		Id:       userID,
		Username: username,
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

	rtToken, err := utils.CreateRefreshTokenString(user.ID, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
	}
	// setting http only cookies for refresh token and auth
	utils.SetAuthCookies(c, token, rtToken, "/api")

	return c.JSON(http.StatusOK, echo.Map{
		"message": "login succesfull",
	})
}

func (s *Server) getCurrentUser(c echo.Context) error {
	authCookie, err := c.Cookie("authToken")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing auth token cookie in request"})
	}
	claims, err := utils.DecodeJwt(s.cfg.SecretKey, authCookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, claims)
}

// MAKE THIS SENSABLE
func (s *Server) getSchoolId(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	// claims, _ := utils.DecodeJwt("asd", "asd")

	if len(username) == 0 || len(password) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Incorect or malformed payload"})
	}

	// topic := claims.Sub.Hex()
	// s.mb.AddTopic(topic)
	// s.mb.PushEvent(topic, broker.Event{Message: "test", Type: "adding"})

	// go client.Execute(username, password, s.mb,topic)

	return c.JSON(http.StatusOK, "yeah yeah")
}

func (s *Server) refreshToken(c echo.Context) error {
	// Retrieve the refreshToken cookie
	rtCookie, err := c.Cookie("refreshToken")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing or invalid refresh token"})
	}

	rt, err := utils.DecodeRefreshToken(s.cfg.SecretKey, rtCookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	user, err := database.FindById[database.UserModel](database.UsersCollection, rt.Sub, s.db)
	if err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Incorect user authorization"})
	}

	newRt, err := utils.CreateRefreshTokenString(rt.Sub, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server Error"})
	}

	newToken, err := utils.CreateJwtString(user.Username, user.ID, s.cfg.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server Error"})
	}

	utils.SetAuthCookies(c, newToken, newRt, "/api")
	return c.JSON(http.StatusOK, echo.Map{
		"result": "Token refreshed",
	})
}

func (s *Server) logout(c echo.Context) error {
	// deleting refresh token and authtoken
	utils.UnsetAuthCookies(c, "")
	return c.JSON(http.StatusOK, echo.Map{
		"resutl": "User logged out",
	})
}

func (s *Server) getDummyData(c echo.Context) error {
	// TODO auth token user.id search userSchoolId if missing return error pls do stuff
	dateFrom := c.QueryParam("dateFrom")
	dateTo := c.QueryParam("dateTo")
	limit := c.QueryParam("limit")
	start := c.QueryParam("start")

	authCookie, err := c.Cookie("authToken")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing auth token cookie in request"})
	}
	claims, err := utils.DecodeJwt(s.cfg.SecretKey, authCookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	log.Println(claims)
	timeNow := time.Now()
	if dateFrom == "" {
		dateFrom = timeNow.Format("2006-01-02")
	}
	if dateTo == "" {
		dateTo = timeNow.Add(time.Hour * 24 * 30).Format("2006-01-02")
	}
	if limit == "" {
		limit = "45"
	}
	if start == "" {
		start = "0"
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Panicln(err)
	}
	client := client.CreateClient(jar)

	url := "https://wu.varsovia.study/wsrest/rest/phz/harmonogram/spersonalizowany?_dc=1735239928643&authUzytkownikId=118811" +
		"&dataOd=" + url.QueryEscape(dateFrom) +
		"&dataDo=" + url.QueryEscape(dateTo) +
		"&widok=STUDENT_SPERSONALIZOWANY&page=1&start=" + url.QueryEscape(start) +
		"&limit=" + url.QueryEscape(limit)
	log.Println(url)
	res, err := client.Get(url)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"resutasdasdasl": "User logged out",
		})
	}

	jsonDecoder := json.NewDecoder(res.Body)
	var jsonBuff utils.TimetableApiResponse
	jsonDecoder.Decode(&jsonBuff)
	parsed, err := utils.ParseTimeTableData(&jsonBuff)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"kasdwa": "User logged out",
		})
	}
	if len(parsed.TimeTableEntries) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "no dates lol",
		})
	}
	return c.JSON(http.StatusOK, parsed)
}
