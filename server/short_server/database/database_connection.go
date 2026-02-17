package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var (
	clientOnce   sync.Once
	client       *mongo.Client
	clientErr    error
	connectMongo = func(uri string) (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(uri))
	}
	pingMongo = func(c *mongo.Client, ctx context.Context) error {
		return c.Ping(ctx, readpref.Primary())
	}
)

// LoadEnv loads environment variables from a ".env" file.
// It logs a warning if the file cannot be read but does not stop execution.
func LoadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: .env file not loaded: %v", err)
	}
}

// InstanceDB initializes and returns the package-level singleton MongoDB client.
//
// On the first call, it loads environment variables, reads MONGODB_URI, creates a client,
// and verifies connectivity with a 10-second ping; any initialization error is recorded
// and returned on this and subsequent calls. Subsequent calls return the same client
// instance and the stored initialization error, if any.
func InstanceDB() (*mongo.Client, error) {
	clientOnce.Do(func() {
		LoadEnv()

		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			clientErr = errors.New("MONGODB_URI is not set")
			return
		}

		client, clientErr = connectMongo(uri)
		if clientErr != nil {
			clientErr = fmt.Errorf("connect mongo: %w", clientErr)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := pingMongo(client, ctx); err != nil {
			clientErr = fmt.Errorf("ping mongo: %w", err)
		}
	})

	return client, clientErr
}

// OpenCollection obtains a MongoDB collection handle for the named collection from the
// database specified by the DATABASE_NAME environment variable.
// It returns an error if collectionName is empty, if DATABASE_NAME is not set, or if
// obtaining the MongoDB client via InstanceDB fails.
func OpenCollection(collectionName string) (*mongo.Collection, error) {
	if collectionName == "" {
		return nil, errors.New("collectionName is required")
	}

	dbClient, err := InstanceDB()
	if err != nil {
		return nil, err
	}

	LoadEnv()
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		return nil, errors.New("DATABASE_NAME is not set")
	}

	return dbClient.Database(databaseName).Collection(collectionName), nil
}

// Disconnect closes the global MongoDB client if one has been initialized.
// It returns any error produced by the client's Disconnect call.
func Disconnect(ctx context.Context) error {
	if client == nil {
		return nil
	}
	return client.Disconnect(ctx)
}