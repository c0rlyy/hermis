package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindById[T any](coll CollectionName, id primitive.ObjectID, db *MongoDb) (T, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	var buff T
	err := ReadOne(filter, coll, &buff, db)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func FindByField[T any](field string, value any, coll CollectionName, db *MongoDb) (T, error) {
	filter := bson.D{{Key: field, Value: value}}
	var buff T
	err := ReadOne(filter, coll, &buff, db)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func GetAll[T any](coll CollectionName, db *MongoDb) ([]T, error) {
	var buff []T
	err := ReadAll(coll, &buff, db, bson.D{})
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func FindByUsername(username string, db *MongoDb) (UserModel, error) {
	filter := bson.D{{Key: "username", Value: username}}
	var buff UserModel
	err := ReadOne(filter, UsersCollection, &buff, db)
	if err != nil {
		return buff, err
	}
	return buff, nil
}

func FindByEmail(email string, db *MongoDb) (UserModel, error) {
	filter := bson.D{{Key: "email", Value: email}}
	var buff UserModel
	err := ReadOne(filter, UsersCollection, &buff, db)
	if err != nil {
		return buff, err
	}
	return buff, nil
}
