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

// Config holds MongoDB connection configuration
type Config struct {
	URI                    string
	MaxPoolSize            *uint64
	MinPoolSize            *uint64
	MaxConnIdleTime        *time.Duration
	ServerSelectionTimeout *time.Duration
	ConnectTimeout         *time.Duration
	PingTimeout            *time.Duration
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig(uri string) *Config {
	maxPool := uint64(100)
	minPool := uint64(10)
	maxIdle := 60 * time.Second
	serverTimeout := 5 * time.Second
	connectTimeout := 10 * time.Second
	pingTimeout := 10 * time.Second

	return &Config{
		URI:                    uri,
		MaxPoolSize:            &maxPool,
		MinPoolSize:            &minPool,
		MaxConnIdleTime:        &maxIdle,
		ServerSelectionTimeout: &serverTimeout,
		ConnectTimeout:         &connectTimeout,
		PingTimeout:            &pingTimeout,
	}
}

// Client represents a MongoDB client wrapper
type Client struct {
	*mongo.Client
	config *Config
}

// Connect establishes a connection to MongoDB with the provided configuration
func Connect(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.URI == "" {
		return nil, fmt.Errorf("URI cannot be empty")
	}

	connectTimeout := 10 * time.Second
	if config.ConnectTimeout != nil {
		connectTimeout = *config.ConnectTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().ApplyURI(config.URI)

	if config.MaxPoolSize != nil {
		clientOptions.SetMaxPoolSize(*config.MaxPoolSize)
	}
	if config.MinPoolSize != nil {
		clientOptions.SetMinPoolSize(*config.MinPoolSize)
	}
	if config.MaxConnIdleTime != nil {
		clientOptions.SetMaxConnIdleTime(*config.MaxConnIdleTime)
	}
	if config.ServerSelectionTimeout != nil {
		clientOptions.SetServerSelectionTimeout(*config.ServerSelectionTimeout)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Validate connection
	pingTimeout := 10 * time.Second
	if config.PingTimeout != nil {
		pingTimeout = *config.PingTimeout
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), pingTimeout)
	defer pingCancel()

	if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
		if discErr := client.Disconnect(ctx); discErr != nil {
			log.Printf("Error during client cleanup after ping failure: %v", discErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected and pinged MongoDB")

	return &Client{
		Client: client,
		config: config,
	}, nil
}

// ConnectSimple is a convenience function for simple connections
func ConnectSimple(uri string) (*Client, error) {
	return Connect(DefaultConfig(uri))
}

// Disconnect closes the MongoDB connection
func (c *Client) Disconnect() error {
	if c.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("error during MongoDB disconnection: %w", err)
	}

	log.Println("MongoDB connection closed")
	return nil
}

// HealthCheck performs a health check on the connection
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.Client == nil {
		return fmt.Errorf("client is nil")
	}

	return c.Client.Ping(ctx, readpref.Primary())
}

// GetConfig returns the current configuration
func (c *Client) GetConfig() *Config {
	return c.config
}
