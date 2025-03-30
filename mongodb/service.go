package mongodb

import (
	"context"

	"github.com/theapemachine/idrinkyourmilkshake/data"
)

type Service struct {
	conn       *Conn
	repository *Repository
}

func NewService(conn *Conn) *Service {
	return &Service{conn: conn, repository: NewRepository(conn)}
}

func (service *Service) ListCollections() ([]string, error) {
	return service.conn.ListCollections()
}

func (service *Service) GetCollectionSchema(collection string) (string, error) {
	return service.conn.GetCollectionSchema(collection)
}

func (service *Service) Find(ctx context.Context, query *data.Query) {
	service.repository.Find(ctx, query)
}

func (service *Service) Create(ctx context.Context, query *data.Query) {
	service.repository.Create(ctx, query)
}

func (service *Service) Update(ctx context.Context, query *data.Query) {
	service.repository.Update(ctx, query)
}

func (service *Service) Upsert(ctx context.Context, query *data.Query) {
	service.repository.Upsert(ctx, query)
}

func (service *Service) Delete(ctx context.Context, query *data.Query) {
	service.repository.Delete(ctx, query)
}

func (service *Service) Transaction(ctx context.Context, query *data.Query) {
	service.repository.Transaction(ctx, query)
}
