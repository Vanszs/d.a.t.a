package database

import (
	"context"
)

type Store interface {
	Connect(ctx context.Context) error
	CreateTable(ctx context.Context, tableName string, schema string) error
	Insert(ctx context.Context, tableName string, data map[string]interface{}) error
	Update(ctx context.Context, tableName string, id string, data map[string]interface{}) error
	GetAll(ctx context.Context, tableName string) ([]map[string]interface{}, error)
	Delete(ctx context.Context, tableName string, id string) error
	Get(ctx context.Context, tableName string, id string) (map[string]interface{}, error)
	Close() error
}
