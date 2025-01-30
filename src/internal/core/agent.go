// agent.go
package core

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

type Agent struct {
	ID            uuid.UUID
	cognitive     *CognitiveEngine
	memoryManager memory.Manager
	character     *characters.Character
	taskManager   *tasks.Manager
	dataManager   data.Manager
	scheduler     *Scheduler
	plugins       *plugins.PluginRegistry
	config        AgentConfig
	log           *zap.SugaredLogger
}

type AgentConfig struct {
	ID            uuid.UUID
	Character     *characters.Character
	LLMClient     llm.Client
	DataManager   data.Manager
	MemoryManager memory.Manager
	Training      struct {
		Enabled       bool
		MaxIterations int
		BatchSize     int
		StopThreshold float64
	}
	Inference struct {
		Temperature    float64
		MaxChainLength int
		MinConfidence  float64
	}
}

type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "IDLE"
	AgentStatusProcessing AgentStatus = "PROCESSING"
	AgentStatusPaused     AgentStatus = "PAUSED"
	AgentStatusError      AgentStatus = "ERROR"
)

// AgentState represents the state of an individual agent
type AgentState struct {
	ID          string
	Status      AgentStatus
	CurrentTask *tasks.Task
	Memory      MemoryMetrics
	Goals       []Goal
	Performance AgentPerformance
	LastAction  time.Time
}

func newAgent(config AgentConfig) *Agent {
	return &Agent{
		ID:            config.ID,
		cognitive:     NewCognitiveEngine(config.LLMClient),
		memoryManager: config.MemoryManager,
		character:     config.Character,
		dataManager:   config.DataManager,
		config:        config,
		log:           zap.S(),
	}
}

func (a *Agent) EvaluateTasks(state *SystemState) []tasks.Task {
	return nil
}

func (a *Agent) ExecuteTask(ctx context.Context, task tasks.Task, prefs map[string]interface{}) (TaskResult, error) {
	// 1. Generate samples using GRPO
	samples := a.cognitive.generateSamples(ctx, task, prefs)

	// 2. Calculate rewards
	rewards := a.cognitive.evaluateSamples(ctx, samples, prefs)

	// 3. Select best action
	selectedAction := a.cognitive.selectBestAction(samples, rewards)

	// 4. Execute and learn
	result := a.executeAction(ctx, selectedAction)
	a.learn(ctx, task, result, prefs)

	return result, nil
}

func (a *Agent) learn(ctx context.Context, task Task, result TaskResult, prefs StakeholderPreferences) error {
	// Calculate combined reward
	baseReward := a.rewardModel.calculateBaseReward(result)
	stakeholderReward := a.calculateStakeholderReward(result, prefs)
	finalReward := baseReward*0.6 + stakeholderReward*0.4

	// Update using GRPO
	return a.cognitive.updateModel(ctx, LearningEntry{
		Task:     task,
		Result:   result,
		Reward:   finalReward,
		Feedback: stakeholderReward,
	})
}

func (a *Agent) ProcessInput(ctx context.Context, input Input) error {
	a.log.Infow("Processing input", "type", input.Type)

	// Build cognitive context
	cognitiveContext, err := a.buildContext(ctx, input)
	if err != nil {
		return err
	}

	// Run training if enabled
	if a.config.Training.Enabled {
		if err := a.cognitive.Train(ctx, cognitiveContext); err != nil {
			a.log.Warnw("Training failed", "error", err)
		}
	}

	// Generate thought chain
	thoughtChain, err := a.cognitive.GenerateThoughtChain(ctx, AnalysisRequest{
		Input:     input,
		Context:   cognitiveContext,
		Character: a.character,
		Tasks:     a.taskManager.GetTasks(ctx),
	})
	if err != nil {
		return err
	}

	// Execute actions from thought chain
	for _, action := range thoughtChain.Actions {
		// Plugin pre-execution hook
		data := map[string]interface{}{
			"action":       action,
			"thoughtChain": thoughtChain,
		}
		if err := a.plugins.Execute(ctx, "before_action", data); err != nil {
			a.log.Warnw("Plugin pre-execution failed", "error", err)
			continue
		}

		// Execute action
		result, err := a.executeAction(ctx, action)
		if err != nil {
			a.storeFailure(ctx, thoughtChain, action, err)
			continue
		}

		// Store success and learn
		a.storeSuccess(ctx, thoughtChain, action)
		a.learn(ctx, input, thoughtChain, result)

		// Plugin post-execution hook
		if err := a.plugins.Execute(ctx, "after_action", data); err != nil {
			a.log.Warnw("Plugin post-execution failed", "error", err)
		}
	}

	return nil
}

func (a *Agent) executeAction(ctx context.Context, action Action) (ActionResult, error) {
	runtime := a.createRuntime()

	a.log.Infow("Executing action", "type", action.Type)

	// Validate action parameters
	if err := action.ValidateParams(action.Params); err != nil {
		return ActionResult{}, err
	}

	// Check if action requires approval
	if action.RequiresApproval() {
		approved, err := a.checkGovernanceApproval(ctx, action)
		if err != nil || !approved {
			return ActionResult{}, ErrActionNotApproved
		}
	}

	// Execute action
	if err := action.Execute(ctx, runtime); err != nil {
		return ActionResult{}, err
	}

	// Calculate impact
	impact := action.EstimateImpact(runtime.GetContext())

	return ActionResult{
		Action: action,
		Impact: impact,
		Time:   time.Now(),
	}, nil
}

func (a *Agent) learn(ctx context.Context, input Input, chain *ThoughtChain, result ActionResult) {
	// Store interaction memory
	a.memoryManager.Store(ctx, memory.Entry{
		Type: "interaction",
		Content: LearningEntry{
			Input:     input,
			Chain:     chain,
			Result:    result,
			Timestamp: time.Now(),
		},
	})

	// Update character knowledge
	a.character.Learn(CharacterLearning{
		ThoughtChain: chain,
		Result:       result,
	})

	// Update cognitive model
	a.cognitive.Learn(ctx, LearningEntry{
		Input:  input,
		Chain:  chain,
		Result: result,
	})
}

func (a *Agent) StartAutonomousLoop(ctx context.Context) {
	a.scheduler.SchedulePeriodic(ctx, time.Hour, func() {
		if err := a.ProcessInput(ctx, Input{
			Type:      Autonomous,
			Content:   "Periodic goal evaluation",
			Timestamp: time.Now(),
		}); err != nil {
			a.log.Errorw("Autonomous processing failed", "error", err)
		}
	})
}

func (a *Agent) createRuntime() Runtime {
	return &agentRuntime{
		agent:     a,
		memory:    a.memoryManager,
		data:      a.dataManager,
		character: a.character,
	}
}
