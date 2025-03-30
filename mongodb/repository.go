package mongodb

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/idrinkyourmilkshake/data"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	conn *Conn
}

type HealthStatus struct {
	OK             bool          `bson:"ok" json:"ok"`
	CircuitBreaker bool          `bson:"circuit_breaker" json:"circuit_breaker"`
	Latency        time.Duration `bson:"latency" json:"latency"`
	Error          string        `bson:"error,omitempty" json:"error,omitempty"`
}

func NewRepository(conn *Conn) *Repository {
	return &Repository{conn: conn}
}

func (repository *Repository) Find(ctx context.Context, query *data.Query) {
	collection := repository.conn.collection(query.Collection)
	if collection == nil {
		query.WithResult(nil)
		return
	}

	var result interface{}
	findOpts := options.Find()
	for k, v := range query.Opts {
		switch k {
		case "limit":
			if limit, ok := v.(int64); ok {
				findOpts.SetLimit(limit)
			}
		case "skip":
			if skip, ok := v.(int64); ok {
				findOpts.SetSkip(skip)
			}
		case "sort":
			if sort, ok := v.(map[string]interface{}); ok {
				findOpts.SetSort(sort)
			}
		}
	}

	cursor, err := collection.Find(ctx, query.Filter, findOpts)
	if err != nil {
		log.Error("error finding documents", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result); err != nil {
		log.Error("error converting cursor to result", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(result)
}

func (repository *Repository) Create(ctx context.Context, query *data.Query) {
	collection := repository.conn.collection(query.Collection)
	if collection == nil {
		query.WithResult(nil)
		return
	}

	result, err := collection.InsertOne(ctx, query.Filter)
	if err != nil {
		log.Error("error inserting document", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(result.InsertedID)
}

func (repository *Repository) Update(ctx context.Context, query *data.Query) {
	collection := repository.conn.collection(query.Collection)
	if collection == nil {
		query.WithResult(nil)
		return
	}

	result, err := collection.UpdateOne(ctx, query.Filter, query.Opts["update"])
	if err != nil {
		log.Error("error updating document", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(result.ModifiedCount)
}

func (repository *Repository) Upsert(ctx context.Context, query *data.Query) {
	collection := repository.conn.collection(query.Collection)
	if collection == nil {
		query.WithResult(nil)
		return
	}

	result, err := collection.UpdateOne(
		ctx,
		query.Filter,
		query.Opts["update"],
		options.UpdateOne().SetUpsert(true),
	)

	if err != nil {
		log.Error("error upserting document", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(map[string]interface{}{
		"matched":  result.MatchedCount,
		"modified": result.ModifiedCount,
		"upserted": result.UpsertedID,
	})
}

func (repository *Repository) Delete(ctx context.Context, query *data.Query) {
	collection := repository.conn.collection(query.Collection)
	if collection == nil {
		query.WithResult(nil)
		return
	}

	result, err := collection.DeleteOne(ctx, query.Filter)
	if err != nil {
		log.Error("error deleting document", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(result.DeletedCount)
}

func (repository *Repository) Transaction(ctx context.Context, query *data.Query) {
	if repository.conn.client == nil {
		query.WithResult(nil)
		return
	}

	session, err := repository.conn.client.StartSession()
	if err != nil {
		log.Error("error starting transaction", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		// Execute the operations defined in the query options
		operations := query.Opts["operations"].([]interface{})
		results := make([]interface{}, len(operations))

		for i, op := range operations {
			operation := op.(map[string]interface{})
			collection := repository.conn.collection(operation["collection"].(string))

			switch operation["type"].(string) {
			case "insertOne":
				res, err := collection.InsertOne(sessCtx, operation["document"])
				if err != nil {
					log.Error("error inserting document", "error", err)
					query.WithResult(nil)
					query.WithError(err)
					return nil, err
				}
				results[i] = res.InsertedID
			case "updateOne":
				res, err := collection.UpdateOne(sessCtx, operation["filter"], operation["update"])
				if err != nil {
					log.Error("error updating document", "error", err)
					query.WithResult(nil)
					query.WithError(err)
					return nil, err
				}
				results[i] = res.ModifiedCount
			case "deleteOne":
				res, err := collection.DeleteOne(sessCtx, operation["filter"])
				if err != nil {
					log.Error("error deleting document", "error", err)
					query.WithResult(nil)
					query.WithError(err)
					return nil, err
				}
				results[i] = res.DeletedCount
			}
		}
		return results, nil
	})

	if err != nil {
		log.Error("error executing transaction", "error", err)
		query.WithResult(nil)
		query.WithError(err)
		return
	}

	query.WithResult(result)
}
