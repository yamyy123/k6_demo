package main

import (
	"context"
	"errors"
	//"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
    client     *mongo.Client
    collection *mongo.Collection
    mu         sync.Mutex // Mutex for protecting shared resources
)

func init() {
    // Connect to MongoDB
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, _ = mongo.Connect(context.TODO(), clientOptions)

    // Check the connection
    err := client.Ping(context.Background(), nil)
    if err != nil {
        panic(err)
    }

    // Access the database and collection
    database := client.Database("employee_database")        // Replace "mydb" with your database name
    collection = database.Collection("tokens") // Replace "tokens" with your collection name
}

func main() {
    router := gin.Default()

    // POST endpoint for storing tokens concurrently
    router.POST("/tokens", func(c *gin.Context) {
        // Check if the request has a token in the "Authorization" header
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Token not found in the header"})
            return
        }

        // Start a goroutine to store the token in MongoDB
        go func() {
            if err := storeToken(token); err != nil {
                // Handle the error, log it, etc.
            }
        }()

        c.JSON(http.StatusOK, gin.H{"message": "Token storage request received"})
    })

    // GET endpoint for retrieving stored tokens concurrently
    router.GET("/gettokens", func(c *gin.Context) {
    
            tokens, err := retrieveTokens()
            if err != nil {
                // Handle the error, log it, etc.
                return
            }
        

            c.JSON(http.StatusOK, gin.H{"tokens": tokens})
     
    
    })

    router.Run(":6000")
   
}

// storeToken stores a token in MongoDB
func storeToken(token string) error {
    mu.Lock()
    defer mu.Unlock()

    _, err := collection.InsertOne(context.TODO(), map[string]interface{}{"token": token})
    return err
}

// retrieveTokens retrieves stored tokens from MongoDB
func retrieveTokens() ([]string, error) {
    // Create a context with a timeout or deadline if needed.
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel() // Cancel the context to clean up resources.

    mu.Lock()
    defer mu.Unlock()
    filter := bson.M{}
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var tokens []string

    for cursor.Next(ctx) {
        var result map[string]interface{}
        if err := cursor.Decode(&result); err != nil {
            return nil, err
        }

        tokenValue, ok := result["token"].(string)
        if !ok {
            return nil, errors.New("Token field is not a string")
        }

        tokens = append(tokens, tokenValue)
    }
   // fmt.Println(tokens)

    return tokens, nil
}
