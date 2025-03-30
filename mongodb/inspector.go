package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBInspector struct {
	models.BaseTool
}

func NewMongoDBInspector() models.ToolType {
	return &MongoDBInspector{
		BaseTool: models.BaseTool{
			ToolName:        "mongodb_inspector",
			ToolDescription: "Inspects MongoDB collections and their schemas",
			ToolParameters: models.Parameter{
				Type: "object",
				Properties: []models.Property{
					{
						Name:        "collection",
						Type:        "string",
						Description: "Optional: The collection name to inspect schema (if not provided, will list all collections)",
					},
					{
						Name:        "sample_size",
						Type:        "integer",
						Description: "Optional: Number of documents to sample for schema inference (default: 10)",
					},
				},
				Required: true,
			},
			Required: []string{"connection_string", "database"},
		},
	}
}

func (mi *MongoDBInspector) Name() string {
	return mi.ToolName
}

func (mi *MongoDBInspector) Description() string {
	return mi.ToolDescription
}

func (mi *MongoDBInspector) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"collection": map[string]any{
				"type":        "string",
				"description": "Optional: The collection name to inspect schema (if not provided, will list all collections)",
			},
			"sample_size": map[string]any{
				"type":        "integer",
				"description": "Optional: Number of documents to sample for schema inference (default: 100)",
			},
		},
		"required": []string{"connection_string", "database"},
	}
}

// Only need to implement Execute and Schema methods
func (mi *MongoDBInspector) Execute(args map[string]any) (string, error) {
	// Extract arguments
	connectionString := os.Getenv("MONGODB_URI")

	database := "FanAppDev2"

	collection, _ := args["collection"].(string)

	// Default sample size to 10 if not provided
	sampleSize := 10
	if val, ok := args["sample_size"].(float64); ok {
		sampleSize = int(val)
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Connecting to MongoDB", "uri", connectionString)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Error("Failed to connect to MongoDB", "error", err)
		return "", fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Error("Failed to ping MongoDB", "error", err)
		return "", fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Info("Successfully connected to MongoDB")

	db := client.Database(database)

	// If collection is specified, get its schema
	if collection != "" {
		return mi.getCollectionSchema(ctx, db, collection, sampleSize)
	}

	// Otherwise, list all collections
	return mi.listCollections(ctx, db)
}

func (mi *MongoDBInspector) listCollections(ctx context.Context, db *mongo.Database) (string, error) {
	log.Info("Listing collections in database", "database", db.Name())

	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Error("Failed to list collections", "error", err)
		return "", fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collections) == 0 {
		return "No collections found in database", nil
	}

	result := fmt.Sprintf("Collections in database %s:\n", db.Name())
	for _, coll := range collections {
		result += fmt.Sprintf("- %s\n", coll)
	}

	log.Info("Successfully listed collections", "count", len(collections))
	return result, nil
}

func (mi *MongoDBInspector) getCollectionSchema(ctx context.Context, db *mongo.Database, collectionName string, sampleSize int) (string, error) {
	log.Info("Getting schema for collection", "collection", collectionName, "sampleSize", sampleSize)

	coll := db.Collection(collectionName)

	// Get document count
	count, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Error("Failed to count documents", "error", err)
		return "", fmt.Errorf("failed to count documents: %w", err)
	}

	if count == 0 {
		return fmt.Sprintf("Collection %s is empty", collectionName), nil
	}

	// Sample documents to infer schema
	findOptions := options.Find().
		SetLimit(int64(sampleSize)).
		SetSort(bson.D{{Key: "Created", Value: -1}})
	cursor, err := coll.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		log.Error("Failed to query collection", "error", err)
		return "", fmt.Errorf("failed to query collection: %w", err)
	}
	defer cursor.Close(ctx)

	// Extract schema from documents
	var documents []bson.M
	if err := cursor.All(ctx, &documents); err != nil {
		log.Error("Failed to decode documents", "error", err)
		return "", fmt.Errorf("failed to decode documents: %w", err)
	}

	if len(documents) == 0 {
		return fmt.Sprintf("No documents found in collection %s", collectionName), nil
	}

	// Infer schema from documents
	schema := inferSchema(documents)

	// Convert schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Error("Failed to marshal schema to JSON", "error", err)
		return "", fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	log.Info("Successfully inferred schema", "collection", collectionName)
	return fmt.Sprintf("Schema for collection %s (inferred from %d documents):\n%s",
		collectionName, len(documents), string(schemaJSON)), nil
}

// inferSchema analyzes documents to determine their structure
func inferSchema(documents []bson.M) map[string]any {
	schema := make(map[string]any)

	// Process each document
	for _, doc := range documents {
		for key, value := range doc {
			// If key doesn't exist in schema yet, add it
			if _, exists := schema[key]; !exists {
				schema[key] = getValueType(value)
			}
		}
	}

	return schema
}

// getValueType determines the type of a value
func getValueType(value any) any {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return "string"
	case int, int32, int64, float32, float64:
		return "number"
	case bool:
		return "boolean"
	case bson.M, map[string]any:
		// For nested objects, recursively determine their structure
		nestedSchema := make(map[string]any)
		for k, val := range v.(bson.M) {
			nestedSchema[k] = getValueType(val)
		}
		return nestedSchema
	case []any:
		// For arrays, determine the type of the first element (if any)
		if len(v) > 0 {
			return []any{getValueType(v[0])}
		}
		return []any{}
	default:
		return fmt.Sprintf("unknown (%T)", value)
	}
}
