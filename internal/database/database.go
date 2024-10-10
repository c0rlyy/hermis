package database

import (
	"context"
	"log"

	"github.com/c0rlyy/hermis/internal/config"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionName string
type DatabaseName string

const (
	UsersCollection        CollectionName = "users"
	RefreshTokenCollection CollectionName = "refresh_tokens"
	DbName                 DatabaseName   = "hermis"
)

type MongoDb struct {
	MongoDb *mongo.Client
}

func NewDb(cfg config.EnvContents) *MongoDb {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoUrl))

	if err != nil {
		log.Fatal(err)

	}

	return &MongoDb{
		MongoDb: client,
	}
}

func (s *MongoDb) GetCollection(CollectionName) *mongo.Collection {
	return s.MongoDb.Database(string(DbName)).Collection(string(UsersCollection))
}

func (s *MongoDb) GetDb() *mongo.Client {
	return s.MongoDb
}
