package log

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
)

const LogsCollection = "logs"

type LogRepository struct {
	collection *mongo.Collection
}

func NewLogRepository(client *mongo.Client, databaseName string) *LogRepository {
	collection := client.Database(databaseName).Collection(LogsCollection)

	return &LogRepository{
		collection: collection,
	}
}

func (r *LogRepository) Create(ctx context.Context, l *log.Log) error {
	_, err := r.collection.InsertOne(ctx, l)

	if err != nil {
		return fmt.Errorf("mongodb: failed to insert log: %w", err)
	}

	return nil
}
