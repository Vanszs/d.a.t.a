package core

import (
	"context"
	"plugin"
	"sync"
	"time"
)

// Main system routines
func (a *Agent) Start() error {
	a.logger.Info("Starting agent system")
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
	// a.logger.Infof("First evaluation in %d seconds", 5)
	for {
		select {
		case <-ticker.C:
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

	return &SystemState{
		Character:              a.character,
		AvailableTools:         a.toolManager.AvailableTools(),
		AvailableActions:       a.actionManager.GetAvailableActions(),
		Timestamp:              time.Now(),
		AgentStates:            a.GetState(),
		StakeholderPreferences: pref,
		ActiveTasks:            a.taskManager.GetTasks(a.ctx),
	}
}

// Social media monitoring
func (a *Agent) monitorSocialInputs() {
	a.processMessage(&SocialMessage{
		Type:     "post",
		Content:  "Hey there! Can you tell me what you are able to do to boost the token price?",
		Platform: "twitter",
	})
	msgQueue := a.socialClient.GetMessageChannel()
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
	// Build prompt for LLM

	processedMsg, err := a.cognitive.processMessage(a.ctx, state, msg)
	if err != nil {
		a.logger.Errorw("Error processing message", "error", err)
		return err
	}

	a.logger.Infof("Processed message: %+v", processedMsg)
	// TODO: process task generation and action taking
	if processedMsg.ShouldReply {
		// Send response
		a.socialClient.SendMessage(SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  processedMsg.ResponseMsg,
		})
	}

	return nil
}

func (a *Agent) Shutdown(ctx context.Context) {
	a.cancel()
}
