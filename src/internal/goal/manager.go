package goal

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Manager struct {
	goals map[string]*Goal
	mu    sync.RWMutex
	log   *zap.SugaredLogger
}

func (m *Manager) UpdateGoalProgress(ctx context.Context, goalID string, metrics map[string]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	goal, exists := m.goals[goalID]
	if !exists {
		return ErrGoalNotFound
	}

	m.log.Debugw("Updating goal progress",
		"goal", goalID,
		"metrics", metrics)

	// Update metrics
	for name, value := range metrics {
		if metric, exists := goal.Metrics[name]; exists {
			metric.Value = value
			metric.UpdatedAt = time.Now()
			goal.Metrics[name] = metric
		}
	}

	// Calculate overall progress
	var totalProgress, totalWeight float64
	for _, metric := range goal.Metrics {
		progress := metric.Value / metric.Target
		totalProgress += progress * metric.Weight
		totalWeight += metric.Weight
	}

	if totalWeight > 0 {
		goal.Progress = totalProgress / totalWeight
	}

	goal.Updated = time.Now()
	return nil
}

func (m *Manager) AdjustGoal(ctx context.Context, goalID string, adjustments map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	goal, exists := m.goals[goalID]
	if !exists {
		return ErrGoalNotFound
	}

	m.log.Infow("Adjusting goal",
		"goal", goalID,
		"adjustments", adjustments)

	// Apply adjustments
	if priority, ok := adjustments["priority"].(int); ok {
		goal.Priority = priority
	}
	if deadline, ok := adjustments["deadline"].(time.Time); ok {
		goal.Deadline = deadline
	}
	if metrics, ok := adjustments["metrics"].(map[string]Metric); ok {
		for name, metric := range metrics {
			goal.Metrics[name] = metric
		}
	}

	goal.Updated = time.Now()
	return nil
}
