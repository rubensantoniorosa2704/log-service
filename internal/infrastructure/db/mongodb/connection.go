package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		if discErr := client.Disconnect(ctx); discErr != nil {
			log.Printf("Error during client cleanup after ping failure: %v", discErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("Successfully connected and pinged MongoDB!")
	return client, nil
}

func Disconnect(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		fmt.Printf("Error during MongoDB disconnection: %v\n", err)
		return
	}
	fmt.Println("MongoDB connection closed.")
}
