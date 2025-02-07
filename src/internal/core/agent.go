// agent.go
package core

import (
	"context"
	"fmt"
	"plugin"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Agent struct {
	ID            uuid.UUID
	cognitive     *CognitiveEngine
	character     *characters.Character
	taskManager   TaskManager
	actionManager ActionManager
	logger        *zap.SugaredLogger
	toolManager   ToolManager
	stakeholders  StakeholderManager
	TokenManager  TokenManager
	socialClient  SocialClient
	Goals         []Goal
	ctx           context.Context
	cancel        context.CancelFunc
}

// SystemState represents the complete state of the agent system
type SystemState struct {
	// General system information
	Timestamp   time.Time
	AgentStates *AgentState

	// Token and stakeholder information
	// TokenState             *TokenState
	StakeholderPreferences map[string]interface{}
	// ActiveVotes            map[string][]Vote

	Character        *characters.Character
	AvailableTools   []Tool
	AvailableActions []Action
	// Task and action information
	ActiveTasks     []*Task
	PendingActions  []Action
	NativeTokenInfo *TokenInfo
}

type Goal struct {
	ID          string
	Name        string
	Description string
	Weight      float64
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
	ID             string
	Status         AgentStatus
	CurrentTask    *Task
	Goals          []Goal
	LastActionTime time.Time
}

// Main system routines
func (a *Agent) Start() error {
	a.logger.Info("Starting agent system")

	for _, account := range a.character.PriorityAccounts {
		_, err := a.stakeholders.FetchOrCreateStakeholder(
			a.ctx,
			account.ID,
			account.Platform,
			StakeholderTypePriority,
		)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	// Start periodic task evaluation
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runPeriodicEvaluation()
	}()

	// Start social media monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.monitorSocialInputs()
	}()

	wg.Wait()
	return nil
}

func (a *Agent) RegisterPlugin(p *plugin.Plugin) {
	// TODO: implement me
}

// Periodic task evaluation
func (a *Agent) runPeriodicEvaluation() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// a.evaluateAndExecuteTasks()
	for {
		select {
		case <-ticker.C:
			// TODO: enable execution
			// a.evaluateAndExecuteTasks()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) evaluateAndExecuteTasks() error {
	a.logger.Info("Evaluating and executing tasks")

	// Get current system state
	state := a.getCurrentState()

	tasks, _ := a.GenerateTasks(context.Background(), state)
	a.logger.Infof("Generated tasks: %d", len(tasks))

	for _, task := range tasks {
		// Check if stakeholder input is needed
		// if task.RequiresStakeholderInput {
		// 	a.requestStakeholderFeedback(task)
		// 	continue
		// }

		// Execute task
		_, err := a.ExecuteTask(a.ctx, task, state)
		if err != nil {
			return err
		}

		// Report results
		// a.reportTaskResults(task, result)
	}

	return nil
}

// In your agent_system.go
func (a *Agent) getCurrentState() *SystemState {
	pref, _ := a.stakeholders.GetAggregatedPreferences(a.ctx)

	nativeToken, _ := a.TokenManager.NativeTokenInfo(a.ctx)
	tasks, _ := a.taskManager.GetTasks(a.ctx)

	return &SystemState{
		Character:              a.character,
		AvailableTools:         a.toolManager.AvailableTools(),
		AvailableActions:       a.actionManager.GetAvailableActions(),
		Timestamp:              time.Now(),
		AgentStates:            a.GetState(),
		StakeholderPreferences: pref,
		ActiveTasks:            tasks,
		NativeTokenInfo:        nativeToken,
	}
}

// Social media monitoring
func (a *Agent) monitorSocialInputs() {
	msgQueue := a.socialClient.GetMessageChannel()
	// TODO graceful shutdown
	go a.socialClient.MonitorMessages(a.ctx)
	for {
		select {
		case msg := <-msgQueue:
			a.processMessage(&msg)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) processMessage(msg *SocialMessage) error {
	state := a.getCurrentState()

	stakeholder, err := a.stakeholders.FetchOrCreateStakeholder(
		a.ctx,
		msg.FromUser,
		msg.Platform,
		StakeholderTypeUser,
	)
	if err != nil {
		a.logger.Errorw("Error fetching stakeholder", "error", err)
		return err
	}

	// a.logger.Infof("Historical message: %+v", stakeholder.HistoricalMsgs)
	a.logger.Infof("Priority accounts: %t", stakeholder.Type == StakeholderTypePriority)

	balance, _ := a.TokenManager.FetchNativeTokenBalance(a.ctx, msg.FromUser, msg.Platform)
	// TODO error handling

	if balance != nil {
		a.logger.Infof("Native token balance: %f", balance.Balance)
		stakeholder.TokenBalance = balance
	}

	processedMsg, err := a.cognitive.processMessage(a.ctx, state, msg, stakeholder)
	if err != nil {
		a.logger.Errorw("Error processing message", "error", err)
		return err
	}

	// TODO fetch or create stakeholder
	// a.stakeholders.FetchOrCreateStakeholder(a.ctx, msg.FromUser)

	a.logger.Infof("Processed message: %+v", processedMsg)
	err = a.stakeholders.AddHistoricalMsg(
		a.ctx,
		msg.FromUser,
		msg.Platform,
		[]string{
			fmt.Sprintf("%s: %s", msg.FromUser, msg.Content),
			fmt.Sprintf("%s: %s", state.Character.Name, processedMsg.ResponseMsg),
		},
	)
	if err != nil {
		a.logger.Errorw("Error adding historical message", "error", err)
		return err
	}

	// TODO: process task generation and action taking
	if processedMsg.ShouldReply {
		// Send response
		a.socialClient.SendMessage(a.ctx, SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  processedMsg.ResponseMsg,
			Metadata: msg.Metadata,
		})
	}

	if processedMsg.ShouldGenerateTask && stakeholder.Type == StakeholderTypePriority {
		a.evaluateAndExecuteTasks()
	}

	return nil
}

func (a *Agent) Shutdown(ctx context.Context) error {
	a.cancel()
	return nil
}

func NewAgent(config AgentConfig) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	logger, _ := zap.NewProduction()

	return &Agent{
		ID:            config.ID,
		cognitive:     NewCognitiveEngine(config.LLMClient, config.Character, logger.Sugar()),
		character:     config.Character,
		logger:        logger.Sugar(),
		toolManager:   config.ToolsManager,
		stakeholders:  config.Stakeholders,
		ctx:           ctx,
		socialClient:  config.SocialClient,
		cancel:        cancel,
		taskManager:   config.TaskManager,
		actionManager: config.ActionManager,
		TokenManager:  config.TokenManager,
	}
}

func (a *Agent) GenerateTasks(ctx context.Context, state *SystemState) ([]*Task, error) {
	tasks, err := a.cognitive.GenerateTasks(ctx, state)
	if err != nil {
		a.logger.Errorw("Failed to evaluate task", "error", err)
		return nil, err
	}

	return tasks.Tasks, nil
}

func (a *Agent) ExecuteTask(ctx context.Context, task *Task, state *SystemState) (*TaskResult, error) {
	// Generate actions using cognitive engine
	actionGen, err := a.cognitive.GenerateActions(ctx, task, state)
	if err != nil {
		return nil, fmt.Errorf("failed to generate actions: %w", err)
	}

	// Execute actions with continuous verification
	var results []error
	for _, action := range actionGen.Actions {
		// Execute action
		err := a.executeAction(ctx, action)
		results = append(results, err)

		// results = append(results, result)

		// Update stakeholders on significant progress
		// if a.isSignificantProgress(result) {
		// 	a.notifyStakeholders(ctx, task, result)
		// }
	}

	return &TaskResult{
		TaskID:    task.ID,
		Task:      task,
		Actions:   actionGen.Actions,
		Timestamp: time.Now(),
		Result:    results,
	}, nil
}

func (a *Agent) GetState() *AgentState {
	return &AgentState{
		ID:             a.ID.String(),
		Status:         AgentStatusIdle,
		CurrentTask:    nil,
		Goals:          a.Goals,
		LastActionTime: time.Now(),
	}
}

func (a *Agent) executeAction(ctx context.Context, action Action) error {
	a.logger.Infow("Executing action", "type", action.Type)

	// Execute action
	if err := action.Execute(); err != nil {
		return err
	}

	return nil
}
