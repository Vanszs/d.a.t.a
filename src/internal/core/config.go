package core

import (
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/google/uuid"
)

type AgentConfig struct {
	ID            uuid.UUID
	Character     *characters.Character
	LLMClient     llm.Client
	Model         string
	TaskManager   TaskManager
	Stakeholders  StakeholderManager
	ActionManager ActionManager
	SocialClient  SocialClient
	ToolsManager  ToolManager
	TokenManager  TokenManager
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

	SystemConfig struct {
		MaxConcurrentTasks int
		TaskTimeout        time.Duration
		MessageBatchSize   int
		MonitorInterval    time.Duration
		Temperature        float64
		MaxChainLength     int
	}
}

func validateConfig(config *AgentConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ID == uuid.Nil {
		return fmt.Errorf("agent ID is required")
	}
	if config.Model == "" {
		return fmt.Errorf("model name is required")
	}
	// Add more validation as needed
	return nil
}
