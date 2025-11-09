package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Validate connection with ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Extract database name from URI
	dbName := extractDatabaseName(uri)
	
	return client.Database(dbName), nil
}

func extractDatabaseName(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "siha" // fallback
	}
	
	// Remove leading slash from path
	dbName := strings.TrimPrefix(parsed.Path, "/")
	
	// Remove query parameters if any
	if idx := strings.Index(dbName, "?"); idx != -1 {
		dbName = dbName[:idx]
	}
	
	if dbName == "" {
		return "siha" // fallback
	}
	
	return dbName
}
