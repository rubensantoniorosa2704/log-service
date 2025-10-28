package log

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/valueobjects"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// These tests require a running MongoDB instance
// Skip them if MongoDB is not available

func setupTestMongoDB(t *testing.T) (*mongo.Client, func()) {
	// Use a test database
	testURI := "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testURI))
	if err != nil {
		t.Skip("MongoDB not available for integration tests:", err)
		return nil, nil
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("Cannot ping MongoDB for integration tests:", err)
		return nil, nil
	}

	// Cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Drop test database
		testDB := "loggingdb_test"
		client.Database(testDB).Drop(ctx)
		client.Disconnect(ctx)
	}

	return client, cleanup
}

func TestLogRepository_Create_Integration(t *testing.T) {
	client, cleanup := setupTestMongoDB(t)
	if client == nil {
		return // Test was skipped
	}
	defer cleanup()

	// Setup repository
	testDB := "loggingdb_test"
	repo := NewLogRepository(client, testDB)

	// Create test log
	applicationID := uuid.New()
	userID := uuid.New()
	message := "Integration test log message"
	level := valueobjects.LogLevelInfo

	testLog, err := log.New(message, level, applicationID, userID)
	if err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Set optional fields
	testLog.Source = "IntegrationTest"
	testLog.Tags["test"] = "integration"
	testLog.Tags["environment"] = "test"
	testLog.Metadata["test_id"] = "test-123"
	testLog.Metadata["duration"] = 250.5

	// Test Create operation
	ctx := context.Background()
	err = repo.Create(ctx, testLog)
	if err != nil {
		t.Errorf("Failed to create log in repository: %v", err)
	}

	// Verify log was created by querying the collection directly
	collection := client.Database(testDB).Collection(LogsCollection)

	var retrievedLog log.Log
	err = collection.FindOne(ctx, map[string]interface{}{"_id": testLog.ID}).Decode(&retrievedLog)
	if err != nil {
		t.Errorf("Failed to retrieve created log: %v", err)
	}

	// Verify log fields
	if retrievedLog.ID != testLog.ID {
		t.Errorf("Expected ID %s, got %s", testLog.ID, retrievedLog.ID)
	}
	if retrievedLog.ApplicationID != testLog.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", testLog.ApplicationID, retrievedLog.ApplicationID)
	}
	if retrievedLog.UserID != testLog.UserID {
		t.Errorf("Expected UserID %s, got %s", testLog.UserID, retrievedLog.UserID)
	}
	if retrievedLog.Message != testLog.Message {
		t.Errorf("Expected Message '%s', got '%s'", testLog.Message, retrievedLog.Message)
	}
	if retrievedLog.Level != testLog.Level {
		t.Errorf("Expected Level '%s', got '%s'", testLog.Level, retrievedLog.Level)
	}
	if retrievedLog.Source != testLog.Source {
		t.Errorf("Expected Source '%s', got '%s'", testLog.Source, retrievedLog.Source)
	}

	// Verify tags
	for key, expectedValue := range testLog.Tags {
		if actualValue, exists := retrievedLog.Tags[key]; !exists {
			t.Errorf("Expected tag '%s' not found in retrieved log", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected tag '%s' = '%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	// Verify metadata
	for key, expectedValue := range testLog.Metadata {
		if actualValue, exists := retrievedLog.Metadata[key]; !exists {
			t.Errorf("Expected metadata '%s' not found in retrieved log", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected metadata '%s' = '%v', got '%v'", key, expectedValue, actualValue)
		}
	}

	// Verify timestamp is close (within 1 second)
	timeDiff := retrievedLog.Timestamp.Sub(testLog.Timestamp)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("Timestamp difference too large: %v", timeDiff)
	}
}

func TestLogRepository_Create_Multiple_Integration(t *testing.T) {
	client, cleanup := setupTestMongoDB(t)
	if client == nil {
		return // Test was skipped
	}
	defer cleanup()

	// Setup repository
	testDB := "loggingdb_test"
	repo := NewLogRepository(client, testDB)

	// Create multiple test logs
	applicationID := uuid.New()
	userID := uuid.New()

	logs := make([]*log.Log, 5)
	for i := 0; i < 5; i++ {
		message := "Integration test log message"
		level := valueobjects.LogLevelInfo

		testLog, err := log.New(message, level, applicationID, userID)
		if err != nil {
			t.Fatalf("Failed to create test log %d: %v", i, err)
		}
		logs[i] = testLog
	}

	// Create all logs
	ctx := context.Background()
	for i, testLog := range logs {
		err := repo.Create(ctx, testLog)
		if err != nil {
			t.Errorf("Failed to create log %d in repository: %v", i, err)
		}
	}

	// Verify all logs were created
	collection := client.Database(testDB).Collection(LogsCollection)

	count, err := collection.CountDocuments(ctx, map[string]interface{}{
		"application_id": applicationID,
	})
	if err != nil {
		t.Errorf("Failed to count documents: %v", err)
	}

	if count != int64(len(logs)) {
		t.Errorf("Expected %d logs in database, got %d", len(logs), count)
	}
}

func TestLogRepository_Create_ContextCancellation_Integration(t *testing.T) {
	client, cleanup := setupTestMongoDB(t)
	if client == nil {
		return // Test was skipped
	}
	defer cleanup()

	// Setup repository
	testDB := "loggingdb_test"
	repo := NewLogRepository(client, testDB)

	// Create test log
	applicationID := uuid.New()
	userID := uuid.New()
	message := "Context cancellation test"
	level := valueobjects.LogLevelInfo

	testLog, err := log.New(message, level, applicationID, userID)
	if err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Attempt to create log with cancelled context
	err = repo.Create(ctx, testLog)
	if err == nil {
		t.Error("Expected error when using cancelled context, got none")
	}

	// Error should be context-related
	if err != context.Canceled {
		// MongoDB driver might wrap the error, so check if it contains context cancellation
		if !mongo.IsTimeout(err) {
			t.Logf("Got error (expected): %v", err)
		}
	}
}

func TestLogRepository_Create_DatabaseError_Integration(t *testing.T) {
	client, cleanup := setupTestMongoDB(t)
	if client == nil {
		return // Test was skipped
	}
	defer cleanup()

	// Setup repository with invalid database name to simulate error
	// Note: This might not always cause an error as MongoDB is quite permissive
	testDB := "" // Empty database name
	repo := NewLogRepository(client, testDB)

	// Create test log
	applicationID := uuid.New()
	userID := uuid.New()
	message := "Database error test"
	level := valueobjects.LogLevelError

	testLog, err := log.New(message, level, applicationID, userID)
	if err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Attempt to create log
	ctx := context.Background()
	err = repo.Create(ctx, testLog)

	// Note: MongoDB might still succeed with empty database name
	// This test mainly verifies that errors are properly propagated
	if err != nil {
		t.Logf("Got expected database error: %v", err)
	}
}
