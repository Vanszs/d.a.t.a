package tasks

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
)

type TaskStatus string

const (
	TaskStatusActive    TaskStatus = "active"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID               string
	Title            string
	Description      string
	Priority         float64
	Status           TaskStatus
	Metrics          map[string]Metric
	Deadline         *time.Time
	RequiresApproval bool
	AutoRenew        bool
	AutoTrigger      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Metric struct {
	Threshold float64
	Current   float64
}

type Manager struct {
	tasks      map[string]*Task
	governance *token.Governance
	settings   TaskSettings
}

type TaskSettings struct {
	AutoSuggest            bool
	SuggestThreshold       float64
	MaxConcurrent          int
	MinStakeholderApproval float64
}

func (m *Manager) SuggestTasks(ctx context.Context, data []data.Data) ([]Task, error) {
	if !m.settings.AutoSuggest {
		return nil, nil
	}

	var suggestions []Task
	// Analyze data and suggest new tasks
	// Submit task proposals if needed
	return suggestions, nil
}

func (m *Manager) AddTask(ctx context.Context, task Task) error {
	if task.RequiresApproval {
		return m.proposeTask(ctx, task)
	}
	return m.addTaskDirectly(ctx, task)
}

func (m *Manager) GetTasks(ctx context.Context) []*Task {
	// Convert map to slice
	var taskSlice []*Task
	for _, task := range m.tasks {
		taskSlice = append(taskSlice, task)
	}
	return taskSlice
}
