package core

import (
	"context"
	"plugin"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
)

// AgentSystem orchestrates the entire agent operation
type AgentSystem struct {
	agents        map[string]*Agent
	stakeholders  *token.StakeholderManager
	taskScheduler *Scheduler
	socialClient  SocialClient
	messageQueue  chan SocialMessage
	ctx           context.Context
	cancel        context.CancelFunc
}

// SystemState represents the complete state of the agent system
type SystemState struct {
	// General system information
	Timestamp     time.Time
	AgentStates   map[string]*AgentState
	GlobalMetrics GlobalMetrics

	// Token and stakeholder information
	// TokenState             *token.TokenState
	StakeholderPreferences map[string]interface{}
	// ActiveVotes            map[string][]token.Vote

	// Task and action information
	ActiveTasks    map[string]*tasks.Task
	PendingActions []Action
	ExecutionQueue []string // Task IDs in execution queue
}

// GlobalMetrics tracks system-wide metrics
type GlobalMetrics struct {
	TotalTransactions   uint64
	ActiveStakeholders  uint64
	SystemUptime        time.Duration
	TotalTasksProcessed uint64
	SuccessRate         float64
}

type AgentSystemConfig struct {
	TokenManager       *token.TokenManager
	StakeholderManager *token.StakeholderManager
}

func NewAgentSystem(config AgentSystemConfig) *AgentSystem {
	ctx, cancel := context.WithCancel(context.Background())

	system := &AgentSystem{
		agents:        make(map[string]*Agent),
		stakeholders:  config.StakeholderManager,
		taskScheduler: NewScheduler(),
		socialClient:  NewSocialClient(),
		messageQueue:  make(chan SocialMessage, 1000),
		ctx:           ctx,
		cancel:        cancel,
	}

	return system
}

func (s *AgentSystem) AddAgent(agentConfig AgentConfig) {
	s.agents[agentConfig.ID.String()] = newAgent(agentConfig)
}

// Main system routines
func (s *AgentSystem) Start() error {
	var wg sync.WaitGroup

	// Start periodic task evaluation
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runPeriodicEvaluation()
	}()

	// Start social media monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.monitorSocialInputs()
	}()

	wg.Wait()
	return nil
}

func (s *AgentSystem) RegisterPlugin(p *plugin.Plugin) {
}

// Periodic task evaluation
func (s *AgentSystem) runPeriodicEvaluation() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.evaluateAndExecuteTasks()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *AgentSystem) evaluateAndExecuteTasks() {
	// Get current system state
	state := s.getCurrentState()

	prefs, _ := s.stakeholders.GetAggregatedPreferences(s.ctx)

	// Evaluate tasks for each agent
	for _, agent := range s.agents {
		tasks := agent.EvaluateTasks(state)

		for _, task := range tasks {
			// Check if stakeholder input is needed
			if task.RequiresStakeholderInput {
				s.requestStakeholderFeedback(task)
				continue
			}

			// Execute task
			result, err := agent.ExecuteTask(s.ctx, task, prefs)
			if err != nil {
				return err
			}

			// Report results
			s.reportTaskResults(task, result)
		}
	}
}

// In your agent_system.go
func (s *AgentSystem) getCurrentState() *SystemState {
	state := NewSystemState()

	// Update agent states
	for id, agent := range s.agents {
		state.UpdateAgentState(id, agent.GetState())
	}

	// Update global metrics
	state.UpdateMetrics(s.calculateGlobalMetrics())

	// Update token state
	state.TokenState = s.stakeholders.GetTokenState()

	// Update tasks
	state.ActiveTasks = s.taskScheduler.GetActiveTasks()

	return state
}

// Stakeholder feedback collection
func (s *AgentSystem) requestStakeholderFeedback(task *tasks.Task) {
	// Prepare feedback request message
	msg := token.SocialMessage{
		Type:     TypeFeedbackRequest,
		Content:  s.generateFeedbackRequest(task),
		Platform: "all",
		Context: map[string]interface{}{
			"task_id":  task.ID,
			"agent_id": task.AgentID,
			"deadline": time.Now().Add(30 * time.Minute),
		},
	}

	// Send to social platforms
	s.socialConnector.SendMessage(msg)

	// Set reminder to check for feedback
	s.taskScheduler.ScheduleReminder(task.ID, 30*time.Minute)
}

// Social media monitoring
func (s *AgentSystem) monitorSocialInputs() {
	for {
		select {
		case msg := <-s.messageQueue:
			s.processStakeholderMessage(msg)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *AgentSystem) processStakeholderMessage(msg token.SocialMessage) error {
	// Process message and update stakeholder state
	input, err := s.stakeholders.ProcessMessage(s.ctx, msg)
	if err != nil {
		return err
	}

	// Update affected agents
	if input.Intent.Type == "feedback" {
		s.updateAgentWithFeedback(input)
	}

	// Check if we need to respond
	if input.RequiresResponse {
		s.generateAndSendResponse(input)
	}

	// Convert message to reward signal
	reward := s.calculateMessageReward(msg)

	// Update agent learning
	agent := s.agents[msg.AgentID]
	agent.cognitive.UpdateFromFeedback(context.Background(), LearningEntry{
		Feedback: msg,
		Reward:   reward,
	})
}

// Task execution with stakeholder influence
func (s *AgentSystem) executeTask(task tasks.Task) error {
	// Get current stakeholder preferences
	prefs, err := s.stakeholders.GetAggregatedPreferences(s.ctx)
	if err != nil {
		return err
	}

	// Execute task with preferences
	agent := s.agents[task.AgentID]
	result, err := agent.ExecuteTask(s.ctx, task, prefs)
	if err != nil {
		return err
	}

	// Report results
	s.reportTaskResults(task, result)
}

// Result reporting
func (s *AgentSystem) reportTaskResults(task Task, result TaskResult) {
	// Generate status update
	msg := SocialMessage{
		Type:     TypeStatusUpdate,
		Content:  s.generateStatusUpdate(task, result),
		Platform: "all",
	}

	// Send update to social platforms
	s.socialClient.SendMessage(msg)

	// Store results
	s.storeTaskResult(task, result)

	// Update agent learning
	s.updateAgentLearning(task.AgentID, result)
}

// Social message handling
type SocialConnector struct {
	twitterClient *TwitterClient
	discordBot    *DiscordBot
	outputQueue   chan SocialMessage
}

func (sc *SocialConnector) SendMessage(msg SocialMessage) error {
	switch msg.Platform {
	case "twitter":
		return sc.twitterClient.Tweet(msg.Content)
	case "discord":
		return sc.discordBot.SendMessage(msg.Content)
	case "all":
		// Send to all platforms
		if err := sc.twitterClient.Tweet(msg.Content); err != nil {
			return err
		}
		return sc.discordBot.SendMessage(msg.Content)
	}
	return nil
}

func (s *AgentSystem) Shutdown(ctx context.Context) {
	s.cancel()
}
