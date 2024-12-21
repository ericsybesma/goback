package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/seebasoft/prompter/goback/database"
	"github.com/seebasoft/prompter/goback/models"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
)

var ginEngine *gin.Engine
var dbClient *mongo.Client

func getRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Hello API Handler",
		"body":  "Endpoints: /rest/users",
	})
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

func registerRoutes(engine *gin.Engine, resourceName string, dbEntity models.DbEntity) {
	engine.POST(fmt.Sprintf("/rest/%s", resourceName), func(c *gin.Context) { Create(c, dbEntity) })
	engine.GET(fmt.Sprintf("/rest/%s", resourceName), func(c *gin.Context) { ReadAll(c, dbEntity) })
	engine.GET(fmt.Sprintf("/rest/%s/:id", resourceName), func(c *gin.Context) { Read(c, dbEntity) })
	engine.PUT(fmt.Sprintf("/rest/%s/:id", resourceName), func(c *gin.Context) { Update(c, dbEntity) })
	//engine.DELETE(fmt.Sprintf("/rest/%s/:id", resourceName), func(c *gin.Context) { DoDelete(c, dbEntity) })
}

func registerRoot(engine *gin.Engine) {
	engine.GET("/", getRoot)
}

func initGin() *gin.Engine {
	engine := gin.Default()
	registerRoot(engine)
	registerRoutes(engine, "users", &models.User{})
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
