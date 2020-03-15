package base

import (
	"context"
	"gokitCouchDB/db"
)

type Service interface {
	Check(ctx context.Context) (bool, error)
	GetDocument(ctx context.Context, req string) (response interface{}, err error)
	PutDocument(ctx context.Context) (bool, error)
}

type baseService struct {
	databases *db.RoundRobin
}

func NewService(roundRobin *db.RoundRobin) Service {
	return &baseService{databases: roundRobin}
}

func (b *baseService) Check(ctx context.Context) (bool, error) {
	return true, nil
}

func (b *baseService) GetDocument(ctx context.Context, req string) (response interface{}, err error) {
	db, err := b.databases.DB()
	if err != nil {
		return nil, err
	}
	return db.GetDocument(ctx, req)
}

func (b *baseService) PutDocument(ctx context.Context) (bool, error) {
	db, err := b.databases.DB()
	if err != nil {
		return false, err
	}
	return db.PutDocument(ctx)
}
