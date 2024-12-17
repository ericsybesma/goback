package main

import (
	//"encoding/json"
	//"fmt"A
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	//"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

// User represents a user object
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// handleRoot serves the root endpoint
func handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Hello API Handler",
		"body":  "API Response",
	})
}

// handleText serves the text endpoint
func handleText(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
        <!DOCTYPE html>
        <html>
        <head>
        <title>Hello</title>
        </head>
        <body>
        <h1>Hello, World!</h1>
        </body>
        </html>
        `)
}

// handleGetUsers serves the users endpoint
func handleGetUsers(c *gin.Context) {
	users := []User{
		{ID: 1, Username: "John Doe", Email: "john_doe@gmail.com"},
		{ID: 2, Username: "Jane Smith", Email: "jane_smith@gmail.com"},
	}
	c.JSON(http.StatusOK, users)
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "Hello3,  World!",
	// })
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
	log.Println("Received request path:", req.RawPath)
	log.Println("HTTP Method:", req.RequestContext.HTTP.Method)
	log.Println("Request Headers:", req.Headers)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	httpReq, _ := http.NewRequest(req.RequestContext.HTTP.Method, req.RawPath, nil)
	ctx.Request = httpReq
	ginEngine.ServeHTTP(w, ctx.Request)

	log.Println("Response Status Code:", w.Code)
	log.Println("Response Headers:", w.Header())
	log.Println("Response Body: ", w.Body.String())

	return events.APIGatewayV2HTTPResponse{
		StatusCode: w.Code,
		Headers:    convertHeader(w.Header()),
		Body:       w.Body.String(),
	}, nil
}

var ginEngine *gin.Engine

func main() {
	ginEngine = initGin()

	// Determine if running in Lambda or locally
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda
		lambda.Start(lambdaHandler)
	} else {
		// Running locally
		//r := initGin()
		ginEngine.Run(":8080")
	}
}

func initGin() *gin.Engine {
	engine := gin.Default()
	engine.GET("/graphql", handleRoot)
	engine.GET("/graphql/text", handleText)
	engine.GET("/graphql/users", handleGetUsers)
	return engine
}
