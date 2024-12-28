package main

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/seebasoft/prompter/goback/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

// Implement handlers that correspond to all of the DalRepo options

// Provide a CRUD interface for database.DalEntity, enabling a REST API for
// any entity implementing this interface.
func Create(c *gin.Context, dalEntity database.DalEntity) {

	if err := c.ShouldBindJSON(dalEntity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	dalEntity.SetID(primitive.NewObjectID())
	objectID, err := dalRepo.Create(dalEntity)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalEntity.SetID(objectID)
	c.JSON(http.StatusCreated, dalEntity)
}

func ReadByID(c *gin.Context, dalEntity database.DalEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	entity, err := dalRepo.ReadByID(objectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func ReadByFilter(c *gin.Context, dalEntity database.DalEntity) {
	filter, err := ExtractFilter(c, dalEntity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	bsonEntities, err := dalRepo.ReadBSON(filter, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	remapMongoIDs(bsonEntities)
	c.JSON(http.StatusOK, bsonEntities)
}

func remapMongoIDs(bsonEntities []bson.M) {
	// handle re-mapping of _id to id
	for _, result := range bsonEntities {
		if id, ok := result["_id"]; ok { // Check if _id exists
			result["id"] = id     // Assign the value to the new key
			delete(result, "_id") // Delete the old key
		}
	}
}

func ExtractFilter(c *gin.Context, dalEntity database.DalEntity) (bson.M, error) {
	filter := make(bson.M)
	query := c.Request.URL.Query()

	// Get the value of the interface
	entityValue := reflect.ValueOf(dalEntity)

	// If it's a pointer, get the value it points to
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
	}

	// Get the type of the concrete value
	entityType := entityValue.Type()

	if entityType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dalEntity must be a struct or pointer to struct")
	}

	for fieldName := 0; fieldName < entityType.NumField(); fieldName++ {
		field := entityType.Field(fieldName)
		bsonTag := field.Tag.Get("bson")
		if bsonTag == "" {
			continue // Skip fields without bson tags
		}
		for param, values := range query {
			if len(values) == 0 {
				continue
			}
			value := values[0]
			if strings.HasPrefix(param, bsonTag+"_") {
				op := strings.TrimPrefix(param, bsonTag+"_")
				switch op {
				case "eq":
					filter[bsonTag] = value
				case "ne":
					filter[bsonTag] = bson.M{"$ne": value}
				case "gt":
					filter[bsonTag] = bson.M{"$gt": value}
				case "gte":
					filter[bsonTag] = bson.M{"$gte": value}
				case "lt":
					filter[bsonTag] = bson.M{"$lt": value}
				case "lte":
					filter[bsonTag] = bson.M{"$lte": value}
				case "contains":
					filter[bsonTag] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(value), Options: "i"}} // Case-insensitive contains
				case "before":
					// Parse value to time.Time if the field type is time.Time
					if field.Type.Kind() == reflect.Struct && field.Type.String() == "time.Time" {
						t, err := time.Parse(time.RFC3339, value)
						if err != nil {
							return nil, err
						}
						filter[bsonTag] = bson.M{"$lt": t}
					}
				case "after":
					if field.Type.Kind() == reflect.Struct && field.Type.String() == "time.Time" {
						t, err := time.Parse(time.RFC3339, value)
						if err != nil {
							return nil, err
						}
						filter[bsonTag] = bson.M{"$gt": t}
					}
				default:
					err := fmt.Errorf("invalid filter operator: %s", op)
					return nil, err
				}
			}
		}
	}

	return filter, nil
}

func UpdateByID(c *gin.Context, dalEntity database.DalEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.ShouldBindJSON(dalEntity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalEntity.SetID(objectID)
	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	_, err = dalRepo.UpdateByID(objectID, dalEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dalEntity)
}

func DeleteByID(c *gin.Context, dalEntity database.DalEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Save the result to get the DeletedCount
	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	deletedCount, err := dalRepo.DeleteByID(objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if deletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No entity found to delete"})
		return
	}

	label := "entries"
	if deletedCount == 1 {
		label = "entry"
	}
	msg := fmt.Sprintf("Deleted %d %s", deletedCount, label)
	c.JSON(http.StatusOK, gin.H{"message": msg})
}
