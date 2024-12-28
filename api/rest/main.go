package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/seebasoft/prompter/goback/database"
	"github.com/seebasoft/prompter/goback/models"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

var ginEngine *gin.Engine
var dbClient *mongo.Client

func main() {
	initialize()
	run()
}

func initialize() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ginEngine = initGin()
	dbClient = initDb()
}

func run() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda
		lambda.Start(lambdaHandler)
	} else {
		// Running locally
		ginEngine.Run(":8080")
	}
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
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	httpReq, _ := http.NewRequest(req.RequestContext.HTTP.Method, req.RawPath, bytes.NewBufferString(req.Body))
	ctx.Request = httpReq
	ginEngine.ServeHTTP(w, ctx.Request)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: w.Code,
		Headers:    convertHeader(w.Header()),
		Body:       w.Body.String(),
	}, nil
}

func getRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title": "Hello API Handler",
		"body":  "Endpoints: users",
	})
}

func setDefaultRoutes(engine *gin.RouterGroup, resourceName string, dalEntity database.DalEntity) {
	engine.POST(resourceName, func(c *gin.Context) { Create(c, dalEntity) })
	engine.GET(fmt.Sprintf("%s/:id", resourceName), func(c *gin.Context) { ReadByID(c, dalEntity) })
	engine.GET(resourceName, func(c *gin.Context) { ReadByFilter(c, dalEntity) })
	engine.PUT(fmt.Sprintf("%s/:id", resourceName), func(c *gin.Context) { UpdateByID(c, dalEntity) })
	engine.DELETE(fmt.Sprintf("%s/:id", resourceName), func(c *gin.Context) { DeleteByID(c, dalEntity) })
}

func initGin() *gin.Engine {
	engine := gin.Default()

    // CORS configuration
    config := cors.DefaultConfig()

    // Use environment variables for allowed origins (BEST PRACTICE)
    allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
    if allowedOrigins == "" {
        // Default for development (VERY IMPORTANT: DO NOT USE "*" IN PRODUCTION)
        config.AllowOrigins = []string{"http://localhost:8000"} // Correct origin for your frontend
		allowNullOrigin := os.Getenv("ALLOW_NULL_ORIGIN") == "true"
		if allowNullOrigin {
			config.AllowOrigins = append(config.AllowOrigins, "*") // Add "null" to allowed origins
			log.Println("WARNING: Allowing requests from 'null' origin. ONLY FOR DEVELOPMENT/TESTING.")
		}
	} else {
        config.AllowOrigins = strings.Split(allowedOrigins, ",")
    }

    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
    config.AllowCredentials = true // Only if you are using cookies or authorization headers

    engine.Use(cors.New(config))

	engine.GET("/", getRoot)
	v1 := engine.Group("/rest/v1")
	setDefaultRoutes(v1, "users", &models.User{})
	return engine
}

func initDb() *mongo.Client {
	var err error
	client, err := database.MongoConnect()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	return client
}
