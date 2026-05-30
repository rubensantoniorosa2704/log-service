package log

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
)

const LogsCollection = "logs"

// ttlSeconds defines how long log entries are retained (30 days).
const ttlSeconds = int32(30 * 24 * 60 * 60)

type LogRepository struct {
	collection *mongo.Collection
}

func NewLogRepository(client *mongo.Client, databaseName string) *LogRepository {
	collection := client.Database(databaseName).Collection(LogsCollection)
	repo := &LogRepository{collection: collection}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *LogRepository) ensureIndexes(ctx context.Context) {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "application_id", Value: 1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("application_id_timestamp"),
		},
		{
			Keys:    bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetName("timestamp_ttl").SetExpireAfterSeconds(ttlSeconds),
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		fmt.Printf("warning: failed to create indexes: %v\n", err)
	}
}

func (r *LogRepository) Create(ctx context.Context, l *log.Log) error {
	_, err := r.collection.InsertOne(ctx, l)
	if err != nil {
		return fmt.Errorf("mongodb: failed to insert log: %w", err)
	}
	return nil
}
