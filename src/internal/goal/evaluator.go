package goal

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Evaluator struct {
	manager *Manager
	log     *zap.SugaredLogger
}

func (e *Evaluator) EvaluateGoalStatus(ctx context.Context, goal *Goal) (GoalStatus, error) {
	if goal.Progress >= 1.0 {
		e.log.Infow("Goal completed",
			"goal", goal.ID,
			"progress", goal.Progress)
		return GoalStatusCompleted, nil
	}

	if goal.Deadline.Before(time.Now()) && goal.Progress < 1.0 {
		e.log.Warnw("Goal failed due to deadline",
			"goal", goal.ID,
			"progress", goal.Progress)
		return GoalStatusFailed, nil
	}

	return goal.Status, nil
}
