package database

// users := []models.User{
//         {ID: 1, Username: "John Doe", Email: "john_doe@gmail.com"},
//         {ID: 2, Username: "Jane Smith", Email: "jane_smith@gmail.com"},
// }

import (
	"context"
	"fmt"

	"github.com/seebasoft/prompter/goback/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//         // ... MongoDB connection logic
// }

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

/// Get the entire user collection, limited to 20 entities
func GetUsers(client *mongo.Client) ([]models.User, error) {
	collection := client.Database("core").Collection("users")
	cursor, err := collection.Find(context.TODO(), bson.D{}, options.Find().SetLimit(20))
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(context.TODO())

	var bsonUsers []bson.M
	if err = cursor.All(context.TODO(), &bsonUsers); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	var users []models.User
	for _, bsonUser := range bsonUsers {
		user := models.User{}

		if id, ok := bsonUser["_id"].(primitive.ObjectID); ok {
			user.ID = id
		} else {
			// Handle the error or set a default value
			return nil, fmt.Errorf("failed to assert _id as ObjectID")
		}

		if username, ok := bsonUser["username"].(string); ok {
			user.Username = username
		} else {
			// Handle the error or set a default value
			return nil, fmt.Errorf("failed to assert username as string")
		}

		if email, ok := bsonUser["email"].(string); ok {
			user.Email = email
		} else {
			// Handle the error or set a default value
			return nil, fmt.Errorf("failed to assert email as string")
		}
		users = append(users, user)
	}

	return users, nil
}

func GetUserByID(client *mongo.Client, id primitive.ObjectID) (models.User, error) {
	collection := client.Database("core").Collection("users")
	var bsonUser bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&bsonUser)

	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:       bsonUser["_id"].(primitive.ObjectID),
		Username: bsonUser["username"].(string),
		Email:    bsonUser["email"].(string),
	}

	return user, err
}


// func CreateUser(dbClient *mongo.Client, user models.User) (*mongo.InsertOneResult, error) {
// 	collection := dbClient.Database("core").Collection("users")
// 	result, err := collection.InsertOne(context.Background(), bson.M{
// 		"username": user.Username,
// 		"email":    user.Email,
// 	})
// 	return result, err
// }

func CreateDbEntity(dbClient *mongo.Client, entity models.DbEntity) (*mongo.InsertOneResult, error) {
	collection := dbClient.Database(entity.DBName()).Collection(entity.CollectionName())
	result, err := collection.InsertOne(context.Background(), entity.ToBSON(true))
	return result, err
}

/// Update a user by ID
func UpdateUser(dbClient *mongo.Client, id primitive.ObjectID, user models.User) (*mongo.UpdateResult, error) {
	collection := dbClient.Database("core").Collection("users")
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": bson.M{
			"username": user.Username,
			"email":    user.Email,
		}})
	return result, err
}

/// Delete a user by ID
func DeleteUser(dbClient *mongo.Client, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	collection := dbClient.Database("core").Collection("users")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return result, err
}

// // Ping Function
// func Ping(client *mongo.Client) error {
//         // Send a ping to confirm a successful connection
//         if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
//                 return fmt.Errorf("failed to ping MongoDB: %w", err)
//         }
//         return nil
// }

func DisconnectFromMongoDB(client *mongo.Client) error {
	// Disconnect from the MongoDB server
	if err := client.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}