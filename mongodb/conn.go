package mongodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

/*
Conn is a MongoDB connection structure that provides methods for connecting to and interacting with a MongoDB server.
*/
type Conn struct {
	mu               sync.RWMutex  // Provides concurrent-safe access to the connection fields
	client           *mongo.Client // The MongoDB client instance
	connStr          string        // Connection string to MongoDB
	database         string        // The default database name
	err              error         // Stores any connection-related error
	config           ConnConfig
	circuitBreaker   *atomic.Bool
	lastCircuitBreak time.Time
}

type ConnConfig struct {
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	SocketTimeout          time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	CircuitBreakerTimeout  time.Duration
	HeartbeatInterval      time.Duration
	RetryInterval          time.Duration
	MaxRetryAttempts       int
}

/*
NewConn creates a new MongoDB connection and returns a pointer to the Conn structure
*/
func NewConn(connStr, database string, config *ConnConfig) *Conn {
	log.Info("connecting to mongodb", "database", database)
	if config == nil {
		defaultConfig := DefaultConnConfig()
		config = &defaultConfig
	}

	// Create connection with circuit breaker initialized
	conn := &Conn{
		connStr:        connStr,
		database:       database,
		config:         *config,
		circuitBreaker: &atomic.Bool{},
	}

	// Only attempt initial connection if circuit breaker is not active
	if !conn.isCircuitBreaker() {
		conn.connect()
	}

	return conn
}

/*
collection returns a reference to a MongoDB collection under the connection's database
*/
func (conn *Conn) collection(collection string) *mongo.Collection {
	if err := conn.ensureConnected(); err != nil {
		log.Error(fmt.Errorf("failed to ensure connection: %w", err))
		return nil
	}

	conn.mu.RLock()
	defer conn.mu.RUnlock()

	if conn.client == nil {
		log.Error(errors.New("no MongoDB client available when accessing collection"))
		return nil
	}

	return conn.client.Database(conn.database).Collection(collection)
}

/*
connect initializes the MongoDB client connection and sets up monitors and logging
*/
func (conn *Conn) connect() *Conn {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	log.Info("attempting to connect to mongodb",
		"endpoint", conn.connStr,
		"database", conn.database,
	)

	clientOptions := options.Client().ApplyURI(conn.connStr).
		SetRetryWrites(true).
		SetRetryReads(true)

	log.Info("connecting to mongodb with options")

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		conn.err = err
		return conn
	}

	conn.client = client
	conn.err = nil
	log.Info("successfully connected to mongodb",
		"database", conn.database,
		"endpoint", conn.connStr,
	)
	return conn
}

/*
ensureConnected checks if the connection is alive by pinging the MongoDB server, and attempts to reconnect if necessary
*/
func (conn *Conn) ensureConnected() error {
	conn.mu.RLock()
	healthy := (conn.err == nil && conn.client != nil)
	conn.mu.RUnlock()

	if healthy {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := conn.client.Ping(ctx, readpref.Primary()); err == nil {
			return nil
		}
	}

	// If not healthy, attempt to reconnect
	log.Warn("MongoDB connection not healthy, attempting to reconnect...")
	return conn.reconnect()
}

/*
reconnect attempts to re-establish the MongoDB connection with an exponential backoff, up to 5 times
*/
func (conn *Conn) reconnect() error {
	var (
		attempt     = 1
		backoff     = conn.config.RetryInterval
		maxAttempts = conn.config.MaxRetryAttempts
	)

	for attempt <= maxAttempts {
		if conn.isCircuitBreaker() {
			return errors.New("circuit breaker is open")
		}

		conn.connect()

		conn.mu.RLock()
		err := conn.err
		conn.mu.RUnlock()

		if err == nil {
			log.Info("Successfully reconnected to MongoDB")
			return nil
		}

		log.Warn(fmt.Sprintf("Reconnection attempt %d/%d failed: %v", attempt, maxAttempts, err))
		time.Sleep(backoff)
		backoff *= 2
		attempt++
	}

	err := fmt.Errorf("failed to reconnect after %d attempts", maxAttempts)
	conn.triggerCircuitBreaker()
	return err
}

/*
Close disconnects the MongoDB client with a timeout of 5 seconds
*/
func (conn *Conn) Close() {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := conn.client.Disconnect(ctx)
		if err != nil {
			log.Error(fmt.Errorf("error occurred while disconnecting: %w", err))
		}
		conn.client = nil
	}
}

func (conn *Conn) isCircuitBreaker() bool {
	if !conn.circuitBreaker.Load() {
		return false
	}

	// Check if circuit breaker timeout has elapsed
	if time.Since(conn.lastCircuitBreak) > conn.config.CircuitBreakerTimeout {
		conn.circuitBreaker.Store(false)
		return false
	}
	return true
}

func (conn *Conn) triggerCircuitBreaker() {
	conn.circuitBreaker.Store(true)
	conn.lastCircuitBreak = time.Now()
	log.Warn("Circuit breaker triggered - temporarily suspending connection attempts")
}

func DefaultConnConfig() ConnConfig {
	return ConnConfig{}
}

/*
ListCollections returns a list of all collections in the database, which is used by the
AI to determine which collections most likely are involved in the integration of a given
API endpoint.
*/
func (conn *Conn) ListCollections() ([]string, error) {
	collections, err := conn.client.Database(conn.database).ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	return collections, nil
}

/*
GetCollectionSchema returns the schema of a given collection, which is used by the
AI to determine which fields are involved in the integration of a given API endpoint.
*/
func (conn *Conn) GetCollectionSchema(collection string) (string, error) {
	schema, err := conn.client.Database(
		conn.database,
	).Collection(
		collection,
	).Indexes().ListSpecifications(
		context.Background(),
	)

	if err != nil {
		return "", err
	}

	schemaJSON, err := json.Marshal(schema)

	if err != nil {
		return "", err
	}

	return string(schemaJSON), nil
}
