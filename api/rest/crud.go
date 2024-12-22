package main

import (
	"fmt"
	"net/http"

	"github.com/seebasoft/prompter/goback/database"
	"github.com/seebasoft/prompter/goback/models"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

func Create(c *gin.Context, dbEntity models.DbEntity) {
	if err := c.ShouldBindJSON(dbEntity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dbEntity.SetID(primitive.NewObjectID())
	_, err := database.CreateDbEntity(dbClient, dbEntity)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dbEntity)
}

// do a readAll using DbEntity rather than getUsers
func ReadAll(c *gin.Context, dbEntity models.DbEntity) {
	entities, err := database.GetAll(dbClient, dbEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities)
}

func Read(c *gin.Context, dbEntity models.DbEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	entity, err := database.GetDbEntityByID(dbClient, dbEntity, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func Update(c *gin.Context, dbEntity models.DbEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.ShouldBindJSON(dbEntity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbEntity.SetID(objectID)
	_, err = database.UpdateDbEntity(dbClient, dbEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dbEntity)
}

func Delete(c *gin.Context, dbEntity models.DbEntity) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Save the result to get the DeletedCount
	result, err := database.DeleteDbEntity(dbClient, dbEntity, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No entity found to delete"})
		return
	}

	label := "entries"
	if result.DeletedCount == 1 {
		label = "entry"
	}
	msg := fmt.Sprintf("Deleted %d %s", result.DeletedCount, label)
	c.JSON(http.StatusOK, gin.H{"message": msg})
}