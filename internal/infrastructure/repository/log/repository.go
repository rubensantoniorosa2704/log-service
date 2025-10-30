package log

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
)

const LogsCollection = "logs"

type LogRepository struct {
	collection *mongo.Collection
}

func NewLogRepository(client *mongo.Client, databaseName string) *LogRepository {
	collection := client.Database(databaseName).Collection(LogsCollection)

	// Create indexes for better query performance
	// Run this in a background goroutine to avoid blocking initialization
	go ensureIndexes(collection)

	return &LogRepository{
		collection: collection,
	}
}

// ensureIndexes creates database indexes for optimal query performance
func ensureIndexes(collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Index on application_id for fast filtering by application
	applicationIDIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "application_id", Value: 1}},
		Options: options.Index().
			SetName("idx_application_id").
			SetBackground(true),
	}

	// Index on timestamp for time-based queries and sorting
	timestampIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().
			SetName("idx_timestamp").
			SetBackground(true),
	}

	// Compound index on application_id and timestamp for optimal log retrieval
	compoundIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "application_id", Value: 1},
			{Key: "timestamp", Value: -1},
		},
		Options: options.Index().
			SetName("idx_application_id_timestamp").
			SetBackground(true),
	}

	// Index on log level for filtering by severity
	levelIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "level", Value: 1}},
		Options: options.Index().
			SetName("idx_level").
			SetBackground(true),
	}

	indexes := []mongo.IndexModel{
		applicationIDIndex,
		timestampIndex,
		compoundIndex,
		levelIndex,
	}

	if _, err := collection.Indexes().CreateMany(ctx, indexes); err != nil {
		// Log error but don't fail - the repository can still function without indexes
		fmt.Printf("Warning: failed to create indexes: %v\n", err)
	}
}

func (r *LogRepository) Create(ctx context.Context, l *log.Log) error {
	_, err := r.collection.InsertOne(ctx, l)

	if err != nil {
		return fmt.Errorf("mongodb: failed to insert log: %w", err)
	}

	return nil
}
