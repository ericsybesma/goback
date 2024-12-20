package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/seebasoft/prompter/goback/database"
	"github.com/seebasoft/prompter/goback/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

var ginEngine *gin.Engine
var dbClient *mongo.Client

// handleRoot serves the /rest root endpoint
func getRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Hello API Handler",
		"body":  "Endpoints: /rest/users",
	})
}

// handleUsers serves the /rest/users endpoint
func getUsers(c *gin.Context) {
	users, err := database.GetUsers(dbClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := database.GetUserByID(dbClient, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func create(c *gin.Context, dbEntity models.DbEntity) {
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

func updateUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		var bodyBytes bytes.Buffer
		_, err = io.Copy(&bodyBytes, c.Request.Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			return
		}
		log.Println("Received request body:", bodyBytes.String())

		// Replace the body so it can be read again later
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.ID = objectID
	_, err = database.UpdateUser(dbClient, objectID, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	_, err = database.DeleteUser(dbClient, objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func convertHeader(header http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range header {
		result[k] = v[0]
	}
	return result
}

// lambdaHandler handles Lambda requests and routes them using Gin
func lambdaHandler(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Ensure log output is set to stdout
	log.SetOutput(os.Stdout)
	log.Println("Received request path:", req.RawPath)
	log.Println("HTTP Method:", req.RequestContext.HTTP.Method)
	log.Println("Request Headers:", req.Headers)
	// Explicitly flush logs
	log.Println("Flushing logs")
	log.Writer().(*os.File).Sync()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	httpReq, _ := http.NewRequest(req.RequestContext.HTTP.Method, req.RawPath, bytes.NewBufferString(req.Body))
	ctx.Request = httpReq
	ginEngine.ServeHTTP(w, ctx.Request)

	log.Println("Response Status Code:", w.Code)
	log.Println("Response Headers:", w.Header())
	log.Println("Response Body: ", w.Body.String())

	// Explicitly flush logs
	log.Println("Flushing logs")
	log.Writer().(*os.File).Sync()

	return events.APIGatewayV2HTTPResponse{
		StatusCode: w.Code,
		Headers:    convertHeader(w.Header()),
		Body:       w.Body.String(),
	}, nil
}

func main() {
	// Set log flags to include date and time
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ginEngine = initGin()

	// Initialize dbClient
	dbClient = initDb()

	// Determine if running in Lambda or locally
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda
		lambda.Start(lambdaHandler)
	} else {
		// Running locally
		ginEngine.Run(":8080")
	}
}

func initGin() *gin.Engine {
	engine := gin.Default()
	engine.GET("/rest/", getRoot)
	engine.GET("/rest/users", getUsers)
	engine.GET("/rest/users/:id", getUser)
	user := models.User{}
	engine.POST("/rest/users", func(c *gin.Context) { create(c, &user) })
	engine.PUT("/rest/users/:id", updateUser)
	engine.DELETE("/rest/users/:id", deleteUser)
	return engine
}

func initDb() *mongo.Client {
	var err error
	client, err := database.ConnectToMongoDB()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	return client
}
