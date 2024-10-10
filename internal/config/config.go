package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvContents struct {
	MongoUrl      string
	MongoUsername string
	MongoPassword string
	Port          string
	SecretKey     string
}

func GetEnv() EnvContents {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoUser := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	mongoPass := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	mongoUrl := os.Getenv("MONGO_URL")
	port := os.Getenv("PORT")
	secretKey := os.Getenv("SECRET_KEY")

	env := newEnvContents(mongoUrl, mongoUser, mongoPass, port, secretKey)
	return env
}

func newEnvContents(monogUrl, mongoUsername, mongoPassowrd, port, secretKey string) EnvContents {

	return EnvContents{
		MongoUrl:      monogUrl,
		MongoUsername: mongoUsername,
		MongoPassword: mongoPassowrd,
		Port:          port,
		SecretKey:     secretKey,
	}
}
