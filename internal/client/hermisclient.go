package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/c0rlyy/hermis/internal/broker"
	"github.com/c0rlyy/hermis/internal/utils"
)

func Execute(username, passowrd string, mb *broker.MessageBroker, topic string) {
	jar, err := CreateJar()
	if err != nil {
		log.Panicln(err)
	}
	client := CreateClient(jar)

	// Get login form HTML
	loginFormURL := "https://wu.humanum.pl/cas/login?service=https%3A%2F%2Fwu.humanum.pl%2Fwu%2Fj_spring_cas_security_check"
	htmlStream, err := utils.GetHtml(loginFormURL, client)
	if err != nil {
		log.Panic(err)
	}

	ltTicket, err := utils.ParseHtmlStreamWithCallback(htmlStream, utils.FindLtTicket)
	if err != nil {
		log.Panic(err)
	}
	execution, err := utils.ParseHtmlStreamWithCallback(htmlStream, utils.FindExecution)
	if err != nil {
		log.Panic(err)
	}

	encodedData := utils.NewFormData(username, passowrd, ltTicket, execution).Encode()
	req, err := http.NewRequest("POST", "https://wu.humanum.pl/cas/login?service=https://wu.humanum.pl/wu/j_spring_cas_security_check", bytes.NewBufferString(encodedData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute login request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error during login request: %v", err)
	}
	defer resp.Body.Close()

	// !!IMPORTANT!!
	// wierd stuff, in my python script i dont need to make a reuqest at this endpoint
	// for whatever reason in the go client, the cookie is dropped, and even if i try to reuse it
	// its no longer valid, so in python either the session id gets revalidated for whatever reason
	// or i get from server a new jsonid cookie that replaces it but my go client does not register it
	// btw love python, without it would have never figured it out
	//SO THIS SETS THE AUTHENTICATED JSONID COOKIE IN THE COOKIE JAR
	client.Get("https://wu.humanum.pl/wsrest/rest/authenticate")

	authReq, err := http.NewRequest("POST", "https://wu.humanum.pl/wsrest/rest/auth_info", nil)
	if err != nil {
		log.Fatalf("Error creating auth request: %v", err)
	}

	authResp, err := client.Do(authReq)
	if err != nil {
		log.Fatalf("Error during auth info request: %v", err)
	}
	defer authResp.Body.Close()

	// reading it into map instead of struct coz the payload has so many fields and type
	// that it would take hour to map and it can all be subject to change
	jsonDecoder := json.NewDecoder(authResp.Body)
	var jsonBuff map[string]any
	jsonDecoder.Decode(&jsonBuff)

	userId, err := GetUserId(jsonBuff)

	if err != nil {
		mb.PushEvent(topic, broker.Event{Message: "adding user failed", Type: "failiure"})
		return
	}
	myString := strconv.FormatFloat(userId, 'f', 2, 64)
	mb.PushEvent(topic, broker.Event{Message: myString, Type: "succes"})
	log.Println(userId)
}

// Creates a client that redirects request, with cookie jar
func CreateClient(jar *cookiejar.Jar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Printf("Redirecting to: %s", req.URL)
			return nil
		},
	}
}

// TODO implementing options
func CreateJar() (*cookiejar.Jar, error) {
	return cookiejar.New(nil)
}

// reads from reponse buffer
func GetUserId(jsonBuff map[string]any) (float64, error) {
	result, ok := jsonBuff["result"].(map[string]any)
	if !ok {
		return 0, errors.New("error while reading from response map reuslt field not found")
	}

	userID, ok := result["userId"]
	if !ok {
		return 0, errors.New("error while reading from response map, userId not found")
	}
	if userID == nil {
		return 0, errors.New("user id was empty, make sure the credentials used to log in were correct")
	}
	return userID.(float64), nil
}
