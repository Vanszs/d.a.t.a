package tasks

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
)

// Implement me
type TaskStore struct {
	db database.Store
}

func NewTaskStore(db database.Store) *TaskStore {
	return &TaskStore{
		db: db,
	}
}

func (t *TaskStore) AddTask(ctx context.Context, task Task) error {
	return nil
}

func (t *TaskStore) GetTask(ctx context.Context, taskID string) (Task, error) {
	return Task{}, nil
}

func (t *TaskStore) UpdateTask(ctx context.Context, task Task) error {
	return nil
}

func (t *TaskStore) GetAllTasks(ctx context.Context) ([]*Task, error) {
	return nil, nil
}
