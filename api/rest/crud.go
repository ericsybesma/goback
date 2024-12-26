package main

import (
	"fmt"
	"net/http"

	"github.com/seebasoft/prompter/goback/database"
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
	var filter map[string]interface{}

	// if err := c.ShouldBindJSON(&filter); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	dalRepo := database.NewMongoDalRepo(dbClient, dalEntity, c.Request.Context())
	entities, err := dalRepo.ReadByFilter(filter, 1, 50)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities)
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
