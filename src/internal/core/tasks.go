package core

import (
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
)

type TaskResult struct {
	TaskID    string
	Task      tasks.Task
	Actions   []Action
	Timestamp time.Time
}

type TaskGeneration struct {
	Chain *ThoughtChain
	Tasks []*tasks.Task
}

// convertThoughtChainToTasks converts a thought chain into concrete tasks
func convertThoughtChainToTasks(chain *ThoughtChain) ([]*tasks.Task, error) {
	var relevantSteps []*ThoughtStep

	// 1. Collect only contributing steps
	for _, step := range chain.Steps {
		if step.ContributesToOutcome {
			relevantSteps = append(relevantSteps, step)
		}
	}

	return []*tasks.Task{}, nil
}
