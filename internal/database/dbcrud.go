package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// reads one record from database and puts it in buffer
func ReadOne[T any](filter bson.D, coll CollectionName, buffer *T, db *MongoDb) error {
	return db.GetCollection(coll).FindOne(context.TODO(), filter).Decode(buffer)

}

// this function takes buffer of type T and filles it with its type istaintces
// pass empty filter if you want to get all data from document
func ReadAll[T any](coll CollectionName, buffer *[]T, db *MongoDb, filter bson.D) error {
	collection := db.GetCollection(coll)
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	return cursor.All(context.TODO(), buffer)
}

// wirtes to a collection and returns the mogno type InserOneResult
// TODO make the model of certian type such as entity
func WriteOne(coll CollectionName, model any, db *MongoDb) (*mongo.InsertOneResult, error) {
	return db.GetCollection(coll).InsertOne(context.TODO(), model)
}

func UpdateOne(filter bson.D, coll CollectionName, model any, db *MongoDb) (*mongo.UpdateResult, error) {
	collection := db.GetCollection(coll)
	update := bson.D{
		{Key: "$set", Value: model},
	}
	return collection.UpdateOne(context.TODO(), filter, update)
}

func DeleteOne(filter bson.D, coll CollectionName, db *MongoDb) (*mongo.DeleteResult, error) {
	return db.GetCollection(coll).DeleteOne(context.TODO(), filter)
}
