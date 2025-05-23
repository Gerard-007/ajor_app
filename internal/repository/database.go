package repository

import (
    "context"
    "log"
    "os"

    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// InitDatabase initializes the MongoDB connection and returns the users collection.
func InitDatabase() (*mongo.Collection, error) {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, relying on system environment variables")
    }

    // Get MongoDB URI from environment
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        log.Fatal("MONGODB_URI not set in environment variables")
    }

    // Set up MongoDB client options
    clientOptions := options.Client().ApplyURI(mongoURI)

    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Ping the database to verify connection
    err = client.Ping(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Connected to MongoDB!")

    // Select the database and collection
    collection := client.Database("ajor_app_db").Collection("users")

    return collection, nil
}