package database

import (
	"context"
	"fmt"

	"github.com/seebasoft/prompter/goback/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB() (*mongo.Client, error) {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(
		"mongodb+srv://promptDbUser:PdoNcJ3DzUWu42FP@promptdbaws.f7iaa.mongodb.net/" +
			"?retryWrites=true&w=majority&appName=PromptDbAws").SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	return client, nil
}

func GetAll(client *mongo.Client, entity models.DbEntity) ([]models.DbEntity, error) {
	collection := client.Database(entity.DBName()).Collection(entity.CollectionName())
	cursor, err := collection.Find(context.Background(), bson.D{}, options.Find().SetLimit(20))
	if err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	defer cursor.Close(context.Background())

	var bsonEntities []bson.M
	if err = cursor.All(context.Background(), &bsonEntities); err != nil {
		return nil, fmt.Errorf("failed to decode entities: %w", err)
	}

	var entities []models.DbEntity
	for _, bsonEntity := range bsonEntities {
		entities = append(entities, entity.FromBSON(bsonEntity))
	}

	return entities, nil
}

func GetDbEntityByID(client *mongo.Client, entity models.DbEntity, id primitive.ObjectID) (models.DbEntity, error) {
	collection := client.Database(entity.DBName()).Collection(entity.CollectionName())
	var bsonEntity bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&bsonEntity)

	if err != nil {
		return nil, err
	}

	return entity.FromBSON(bsonEntity), nil
}

func CreateDbEntity(dbClient *mongo.Client, entity models.DbEntity) (*mongo.InsertOneResult, error) {
	collection := dbClient.Database(entity.DBName()).Collection(entity.CollectionName())
	result, err := collection.InsertOne(context.Background(), entity.ToBSON(true))
	return result, err
}

func UpdateDbEntity(dbClient *mongo.Client, entity models.DbEntity) (*mongo.UpdateResult, error) {
	collection := dbClient.Database(entity.DBName()).Collection(entity.CollectionName())
	filter := bson.M{"_id": entity.GetID()}
	update := bson.M{"$set": entity.ToBSON(false)}
	return collection.UpdateOne(context.Background(), filter, update)
}

func DeleteDbEntity(dbClient *mongo.Client, entity models.DbEntity, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := dbClient.Database(entity.DBName()).Collection(entity.CollectionName())
    filter := bson.M{"_id": id}
    return collection.DeleteOne(context.Background(), filter)
}

func DisconnectFromMongoDB(client *mongo.Client) error {
	// Disconnect from the MongoDB server
	if err := client.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}