package goal

import (
	"time"
)

type Goal struct {
	ID          string
	Name        string
	Description string
	Priority    int
	Status      GoalStatus
	Progress    float64
	Deadline    time.Time
	Created     time.Time
	Updated     time.Time
	Metrics     map[string]Metric
}

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "ACTIVE"
	GoalStatusCompleted GoalStatus = "COMPLETED"
	GoalStatusFailed    GoalStatus = "FAILED"
	GoalStatusPaused    GoalStatus = "PAUSED"
)

type Metric struct {
	Name      string
	Value     float64
	Target    float64
	Weight    float64
	UpdatedAt time.Time
}
