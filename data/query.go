package data

import "go.mongodb.org/mongo-driver/v2/bson"

/*
Query defines a generic wrapper around data lookup queries, making it loosely
compatible with most data stores.
*/
type Query struct {
	Collection string
	Filter     bson.M
	Opts       map[string]any
	Payload    any
	Result     any
	err        error
}

func NewQuery(collection string) *Query {
	return &Query{
		Collection: collection,
		Filter:     bson.M{},
		Opts:       make(map[string]any),
		Payload:    nil,
		Result:     nil,
		err:        nil,
	}
}

// WithPayload sets the payload for the query
func (query *Query) WithPayload(payload any) *Query {
	query.Payload = payload
	return query
}

// SetFilter sets the filter for the query
func (query *Query) WithFilter(filter bson.M) *Query {
	query.Filter = filter
	return query
}

// SetOpts sets the options for the query
func (query *Query) WithOpts(opts map[string]any) *Query {
	query.Opts = opts
	return query
}

// SetResult sets the result for the query
func (query *Query) WithResult(result any) *Query {
	query.Result = result
	return query
}

// SetError sets the error for the query
func (query *Query) WithError(err error) *Query {
	query.err = err
	return query
}

// GetError returns the error for the query
func (query *Query) Error() string {
	return query.err.Error()
}
