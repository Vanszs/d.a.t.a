// internal/core/agent.go

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/action"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/goal"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
)

type InputType string

const (
	Unknown    InputType = "Unknown"
	Autonomous InputType = "Autonomous"
)

// Agent represents the autonomous AI agent
type Agent struct {
	ID          uuid.UUID
	cognitive   *CognitiveEngine
	memory      memory.Manager
	character   *characters.Character
	goalManager *goal.Manager
	dataManager *data.Manager
	scheduler   *Scheduler
	plugins     *plugins.PluginRegistry // Inject plugin registry
}

type Runtime interface {
	GetMemory() memory.Manager
	GetData() *data.Manager
	GetGoals() *goal.Manager
	GetCharacter() *characters.Character
}

type Input struct {
	Type      InputType
	Content   string
	Timestamp time.Time
}

// ProcessInput handles both user inputs and autonomous checks
func (a *Agent) ProcessInput(ctx context.Context, input Input) error {
	runtime := a.createRuntime()
	// 1. Update global context with new input
	newContext := a.contextBuilder.BuildContext(ctx, runtime, input)

	// 2. Cognitive analysis
	analysis, err := a.cognitive.Analyze(ctx, AnalysisRequest{
		Input:     input,
		Context:   newContext,
		Character: a.character,
		Goals:     a.goalManager.GetActiveGoals(),
	})
	if err != nil {
		return err
	}

	// 3. Execute actions through runtime interface
	for _, actionResult := range analysis.Actions {
		action, exists := a.cognitive.actionRegistry.Get(actionResult.Name)
		if !exists {
			continue
		}

		data := map[string]interface{}{"action": action, "analysis": analysis}
		if err := a.plugins.Execute(ctx, "filter_action", data); err != nil {
			return err
		}

		if err := action.Execute(ctx, runtime); err != nil {
			a.memory.StoreFailure(ctx, actionResult, err)
			continue
		}
	}

	// 4. Learn from interaction
	a.learn(ctx, input, analysis)

	return nil
}

func (a *Agent) StartAutonomousLoop(ctx context.Context) {
	a.scheduler.SchedulePeriodic(ctx, time.Hour, func() {
		a.ProcessInput(ctx, Input{
			Type:      Autonomous,
			Content:   "Periodic goal evaluation",
			Timestamp: time.Now(),
		})
	})
}

// executeActions performs the selected actions
func (a *Agent) executeActions(ctx context.Context, actions []action.Action) (Response, error) {
	var results []action.Result

	// Execute each action in sequence
	for _, act := range actions {
		result, err := act.Execute(ctx)
		if err != nil {
			// Handle failed action
			return a.handleFailedAction(act, err)
		}
		results = append(results, result)
	}

	// Generate response based on action results
	return a.generateResponse(results)
}

func (a *Agent) learn(ctx context.Context, input Input, analysis Analysis) {
	// Store interaction memory
	a.memory.Store(ctx, memory.Entry{
		Input:    input,
		Analysis: analysis,
		Time:     time.Now(),
	})

	// Update goal progress
	a.goalManager.UpdateProgress(ctx, analysis.GoalImpact)

	// Evolve character preferences
	a.character.Learn(analysis.Outcomes)
}

func (a *Agent) buildContext(ctx context.Context, input Input) Context {
	return CognitiveContext{
		Goals:      a.goalManager.GetActiveGoals(),
		TokenState: a.dataManager.GetTokenState(),
		MarketData: a.dataManager.GetMarketData(),
		Memory:     a.memory.GetRecent(10),
	}
}

func (a *Agent) createRuntime() Runtime {
	return &agentRuntime{
		agent: a,
	}
}

// BuildContext collects and preprocesses input, generating context for analysis
func (a *Agent) BuildContext(ctx context.Context, runtime Runtime, input Input) (CognitiveContext, error) {
	dataList := []data.Data{}
	for _, source := range c.sources {
		// Call BeforeFetch hook
		if err := source.BeforeFetch(ctx, runtime, input); err != nil {
			return CognitiveContext{}, fmt.Errorf("before fetch failed: %w", err)
		}

		data, err := source.Fetch(ctx, runtime, input) // Fetch data from source
		if err != nil {
			return CognitiveContext{}, fmt.Errorf("fetch data failed: %w", err)
		}

		dataList = append(dataList, data)
	}
	return CognitiveContext{
		Goals:      a.goalManager.GetActiveGoals(),
		TokenState: a.dataManager.GetTokenState(),
		DataList:   dataList,
		Memory:     a.memory.GetRecent(10),
	}, nil
}
