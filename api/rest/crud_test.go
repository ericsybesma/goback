package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/seebasoft/prompter/goback/dal"
)


func TestReadByKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/read/:id", func(c *gin.Context) {
		item := &dal.MockItem{}
		ReadByKey(c, item)
	})

	objectID := primitive.NewObjectID()
	req, _ := http.NewRequest(http.MethodGet, "/read/"+objectID.Hex(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
