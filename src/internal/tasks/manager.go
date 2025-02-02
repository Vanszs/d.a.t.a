package tasks

import (
	"context"
	"time"
)

type TaskStatus string

const (
	TaskStatusActive    TaskStatus = "active"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID                       string
	Name                     string
	Description              string
	Priority                 float64
	ExecutionSteps           []string
	Status                   TaskStatus
	Deadline                 *time.Time
	RequiresApproval         bool
	RequiresStakeholderInput bool
	Tools                    []string
	CreatedBy                string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type Metric struct {
	Threshold float64
	Current   float64
}

type Manager struct {
	store *TaskStore
}

type TaskSettings struct {
	AutoSuggest            bool
	SuggestThreshold       float64
	MaxConcurrent          int
	MinStakeholderApproval float64
}

func NewManager(store *TaskStore) *Manager {
	return &Manager{
		store: store,
	}
}

func (m *Manager) AddTask(ctx context.Context, task Task) error {
	return m.store.AddTask(ctx, task)
}

func (m *Manager) GetTasks(ctx context.Context) []*Task {
	taskSlice, _ := m.store.GetAllTasks(ctx)
	return taskSlice
}
