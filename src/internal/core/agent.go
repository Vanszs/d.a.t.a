// agent.go
package core

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Agent struct {
	ID             uuid.UUID
	cognitive      *CognitiveEngine
	character      *characters.Character
	logger         *zap.SugaredLogger
	stakeholders   StakeholderManager
	tokenManager   TokenManager
	socialClient   SocialClient
	pluginRegistry *plugins.Registry
	ctx            context.Context
	cancel         context.CancelFunc
}

// SystemState represents the complete state of the agent system
type SystemState struct {
	// General system information
	Timestamp time.Time

	Character        *characters.Character
	AvailableActions []actions.IAction
	AvailablePlugins []plugins.Plugin
	NativeTokenInfo  *TokenInfo
	ProviderStates   []*plugins.ProviderState
}

func NewAgent(config AgentConfig) (*Agent, error) {
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		ID:             config.ID,
		character:      config.Character,
		cognitive:      NewCognitiveEngine(config.LLMClient, config.Model, config.Character, config.PromptTemplates),
		logger:         logger.GetLogger(),
		stakeholders:   config.Stakeholders,
		tokenManager:   config.TokenManager,
		socialClient:   config.SocialClient,
		pluginRegistry: config.PluginRegistry,
		ctx:            ctx,
		cancel:         cancel,
	}

	return agent, nil
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

	// Start social media monitoring
	go func() {
		a.monitorSocialInputs()
	}()

	a.socialClient.SendMessage(a.ctx, SocialMessage{
		Platform: "Twitter",
		Type:     "Response",
		Content:  "Hello, world!",
	})
	return nil
}

// In your agent_system.go
func (a *Agent) getCurrentState() *SystemState {
	nativeToken, _ := a.tokenManager.NativeTokenInfo(a.ctx)

	// Get plugin actions and provider states
	var pluginActions []actions.IAction
	var providerStates []*plugins.ProviderState

	if a.pluginRegistry != nil {
		// Collect actions from plugins
		for _, plugin := range a.pluginRegistry.GetPlugins() {
			for _, action := range plugin.Actions() {
				pluginActions = append(pluginActions, action)
			}
		}

		// Collect provider states
		for _, provider := range a.pluginRegistry.GetProviders() {
			if state, err := provider.GetProviderState(a.ctx); err == nil {
				providerStates = append(providerStates, state)
			} else {
				a.logger.Warnw("Failed to get provider state",
					"provider", provider.Name(),
					"error", err,
				)
			}
		}
	}

	// print all available actions
	for _, action := range pluginActions {
		a.logger.Infof("Available action: %s", action.Name())
	}

	// print all provider states
	for _, state := range providerStates {
		a.logger.Infof("Provider state: %+v", state)
	}

	return &SystemState{
		Character:        a.character,
		AvailableActions: pluginActions,
		Timestamp:        time.Now(),
		NativeTokenInfo:  nativeToken,
		ProviderStates:   providerStates,
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

// executeAction executes a generic action
func (a *Agent) executeAction(ctx context.Context, action actions.IAction, params map[string]interface{}) error {
	a.logger.Infow("Executing action", "type", action.Type(), "params", params)
	err := action.Execute(ctx, params)
	if err != nil {
		return err
	}

	if rp, ok := action.(actions.ResultProvider); ok {
		res := rp.LastResult()
		if res != "" {
			meta, _ := params["metadata"].(map[string]interface{})
			platform, _ := params["platform"].(string)
			a.socialClient.SendMessage(ctx, SocialMessage{
				Platform: platform,
				Type:     "Response",
				Content:  res,
				Metadata: meta,
			})
		}
	}
	return nil
}

func (a *Agent) processMessage(msg *SocialMessage) error {
	var err error
	defer func() {
		if err != nil {
			a.logger.Errorw("Error processing message", "error", err)
			a.socialClient.SendMessage(a.ctx, SocialMessage{
				Platform: msg.Platform,
				Type:     "Response",
				Content:  "Something went wrong. Please try again later.",
				Metadata: msg.Metadata,
			})
		}
	}()

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

	a.logger.Infof("Priority accounts: %t", stakeholder.Type == StakeholderTypePriority)

	balance, _ := a.tokenManager.FetchNativeTokenBalance(a.ctx, msg.FromUser, msg.Platform)
	if balance != nil {
		a.logger.Infof("Native token balance: %f", balance.Balance)
		stakeholder.TokenBalance = balance
	}

	processedMsg, err := a.cognitive.processMessage(a.ctx, state, msg, stakeholder)
	if err != nil {
		a.logger.Errorw("Error processing message", "error", err)
		return err
	}

	if processedMsg.ShouldGenerateAction {
		for _, action := range processedMsg.Actions {
			var actionImpl actions.IAction
			if a.pluginRegistry != nil {
				for _, plugin := range a.pluginRegistry.GetPlugins() {
					for _, pluginAction := range plugin.Actions() {
						if pluginAction.Type() == action.ActionType && pluginAction.Name() == action.ActionName {
							actionImpl = pluginAction
							break
						}
					}
					if actionImpl != nil {
						break
					}
				}
			}

			if actionImpl == nil {
				a.logger.Errorw("Error getting action", "error", err)
				return err
			}
			a.logger.Infof("Action found in pluginRegistry: %s", actionImpl.Name())

			params, err := a.cognitive.generateActionParameters(a.ctx, state, msg, stakeholder, actionImpl)
			if err != nil {
				a.logger.Errorw("Error generating action parameters", "error", err)
				return err
			}

			// include message metadata so actions can respond
			if params == nil {
				params = map[string]interface{}{}
			}
			params["metadata"] = msg.Metadata
			params["platform"] = msg.Platform
			params["from_user"] = msg.FromUser

			if moreInfoNeeded, ok := params["more_info_needed"].(bool); ok && moreInfoNeeded {
				a.logger.Infof("More info needed, relying on message: %s", params["rely_message"])
				processedMsg.ResponseMsg = params["rely_message"].(string)
				processedMsg.ShouldReply = true
				continue
			}

			if err = a.executeAction(a.ctx, actionImpl, params); err != nil {
				a.logger.Errorw("Error executing action", "error", err)
				return err
			}
		}
	}

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

	if processedMsg.ShouldReply {
		// If we didn't send a response with analysis, send the original response
		a.socialClient.SendMessage(a.ctx, SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  processedMsg.ResponseMsg,
			Metadata: msg.Metadata,
		})
	}

	// if processedMsg.ShouldGenerateTask && stakeholder.Type == StakeholderTypePriority {
	// 	a.evaluateAndExecuteTasks()
	// }

	return nil
}

func (a *Agent) Shutdown(ctx context.Context) error {
	a.cancel()
	return nil
}
