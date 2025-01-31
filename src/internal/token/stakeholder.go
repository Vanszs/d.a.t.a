package token

import (
	"context"
	"math"
	"math/big"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/google/uuid"
)

// StakeholderManager manages stakeholder interactions and influences
type StakeholderManager struct {
	tokenManager  *TokenManager
	memoryManager memory.Manager
	messageParser *MessageParser
	store         *StakeholderStore
}

// StakeholderInput represents processed input from social media
type StakeholderInput struct {
	StakeholderID string
	InputIntent   InputIntent
	Platform      string // "twitter", "discord"
	MessageID     string
	Content       string
	Timestamp     time.Time
	TokenBalance  *big.Int
	Preferences   map[string]interface{}
}

// InputIntent classifies the type of stakeholder input
type InputIntent struct {
	Type         string  // "feedback", "preference", "governance"
	TargetAgent  string  // Which agent this affects
	TargetAspect string  // What aspect to modify
	Sentiment    float64 // -1 to 1 sentiment score
	Confidence   float64 // How confident we are in the interpretation
}

// Message types for different interactions
type MessageType string

const (
	TypeFeedbackRequest MessageType = "FEEDBACK_REQUEST"
	TypeStatusUpdate    MessageType = "STATUS_UPDATE"
	TypeAlert           MessageType = "ALERT"
)

type SocialMessage struct {
	UserID      string
	Type        MessageType
	Content     string
	Platform    string
	TargetAgent string
	Timestamp   time.Time
	Context     map[string]interface{}
}

// Preference represents a stakeholder's preference for an aspect
type Preference struct {
	Value     interface{}
	Weight    float64 // Based on token holdings
	UpdatedAt time.Time
	Source    string // Which platform/message set this
}

func NewStakeholderManager(memoryManager memory.Manager, tokenManager *TokenManager) *StakeholderManager {
	return &StakeholderManager{
		tokenManager:  tokenManager,
		messageParser: NewMessageParser(),
		memoryManager: memoryManager,
	}
}

// ProcessMessage handles new input from social media
func (sm *StakeholderManager) ProcessMessage(ctx context.Context, msg *SocialMessage) (*StakeholderInput, error) {
	// 1. Parse message to understand intent
	input, err := sm.messageParser.Parse(msg)
	if err != nil {
		return nil, err
	}

	// 2. Validate stakeholder token holdings
	holdings, err := sm.tokenManager.GetHoldings(input.StakeholderID)
	if err != nil {
		return nil, err
	}
	input.TokenBalance = holdings

	// 3. Update stakeholder state
	if err := sm.updateStakeholderState(ctx, input); err != nil {
		return nil, err
	}

	return input, nil
}

// updateStakeholderState persists changes from new input
func (sm *StakeholderManager) updateStakeholderState(ctx context.Context, input *StakeholderInput) error {
	// Get current state
	state, err := sm.store.GetStakeholderState(ctx, input.StakeholderID)
	if err != nil {
		state = &StakeholderState{
			ID:          input.StakeholderID,
			LastUpdated: time.Now(),
		}
	}

	// Update preferences based on input
	for k, v := range input.Preferences {
		state.Preferences[k] = Preference{
			Value:     v,
			Weight:    calculateWeight(input.TokenBalance),
			UpdatedAt: input.Timestamp,
			Source:    input.Platform,
		}
	}

	// Persist to database
	return sm.store.SaveStakeholderState(ctx, state)
}

// GetAggregatedPreferences gets current preferences weighted by stake
func (sm *StakeholderManager) GetAggregatedPreferences(ctx context.Context) (map[string]interface{}, error) {
	// Get all stakeholder states
	states, err := sm.store.GetAllStates(ctx)
	if err != nil {
		return nil, err
	}

	// Aggregate preferences weighted by token holdings
	aggregated := make(map[string]interface{})
	for _, state := range states {
		weight := calculateWeight(state.TokenBalance)
		for k, pref := range state.Preferences {
			aggregated[k] = aggregatePreference(aggregated[k], pref.Value, weight)
		}
	}

	return aggregated, nil
}

// aggregatePreference combines two preference values based on weight
// The exact implementation depends on the type of preference value
func aggregatePreference(existing, new interface{}, weight float64) interface{} {
	switch v := new.(type) {
	case float64:
		// For numeric preferences, do a weighted average
		if e, ok := existing.(float64); ok {
			return e*(1-weight) + v*weight
		}
		return v

	case bool:
		// For boolean preferences, use weight as probability threshold
		if e, ok := existing.(bool); ok {
			// If weights heavily favor the new value, use it
			if weight > 0.7 {
				return v
			}
			// Otherwise keep existing
			return e
		}
		return v

	case string:
		// For string preferences, keep the value from the higher weight
		if e, ok := existing.(string); ok {
			if weight > 0.5 {
				return v
			}
			return e
		}
		return v

	case map[string]interface{}:
		// For nested preferences, recursively aggregate each field
		result := make(map[string]interface{})
		if e, ok := existing.(map[string]interface{}); ok {
			// Start with existing values
			for k, val := range e {
				result[k] = val
			}
			// Update with new values using weight
			for k, val := range v {
				if existingVal, exists := result[k]; exists {
					result[k] = aggregatePreference(existingVal, val, weight)
				} else {
					result[k] = val
				}
			}
			return result
		}
		return v

	case []interface{}:
		// For array preferences, combine arrays with weight-based selection
		if e, ok := existing.([]interface{}); ok {
			// Calculate how many items to take from each array based on weight
			newCount := int(float64(len(v)) * weight)
			oldCount := len(e) - newCount
			if oldCount < 0 {
				oldCount = 0
			}

			// Create combined result
			result := make([]interface{}, 0, oldCount+newCount)
			result = append(result, e[:oldCount]...)
			result = append(result, v[:newCount]...)
			return result
		}
		return v

	default:
		// For unsupported types, just return the new value
		return v
	}
}

// calculateWeight determines a stakeholder's voting weight based on their token balance
// Returns a normalized weight between 0 and 1
func calculateWeight(balance *big.Int) float64 {
	if balance == nil || balance.Sign() <= 0 {
		return 0.0
	}

	// Convert balance to float64 for calculations
	balanceFloat := new(big.Float).SetInt(balance)
	weightFloat, _ := balanceFloat.Float64()

	// Apply logarithmic scaling to prevent large token holders from having too much influence
	// while still maintaining meaningful weight differences
	// Using log base 10 plus 1 to ensure positive weights and handle small balances
	weight := math.Log10(weightFloat + 1)

	// Normalize weight to be between 0 and 1
	// You might want to adjust these constants based on your token economics
	const maxLogWeight = 15.0 // log10 of 1 quadrillion (10^15) + 1
	normalizedWeight := weight / maxLogWeight

	// Ensure weight is between 0 and 1
	if normalizedWeight > 1.0 {
		return 1.0
	}
	return normalizedWeight
}

// Example message parser for social media inputs
type MessageParser struct {
	// nlp           *NLPProcessor
	// intentModel   *IntentClassifier
	prefExtractor *PreferenceExtractor
}

func NewMessageParser() *MessageParser {
	return &MessageParser{
		prefExtractor: &PreferenceExtractor{},
	}
}

func (mp *MessageParser) Parse(msg SocialMessage) (*StakeholderInput, error) {
	prefs, err := mp.prefExtractor.Extract(msg.Content)
	if err != nil {
		return nil, err
	}

	return &StakeholderInput{
		StakeholderID: msg.UserID,
		Platform:      msg.Platform,
		MessageID:     uuid.New().String(),
		Content:       msg.Content,
		Timestamp:     msg.Timestamp,
		Preferences:   prefs,
	}, nil
}

// Example usage with social media integration
// func ExampleStakeholderSystem() {
// 	// Initialize system
// 	sm := NewStakeholderManager(db, tokenManager)

// 	// Process Twitter message
// 	twitterMsg := SocialMessage{
// 		Platform:  "twitter",
// 		Content:   "The trading agent should be more conservative with stop losses set at 2%",
// 		Timestamp: time.Now(),
// 	}

// 	sm.ProcessMessage(context.Background(), twitterMsg)

// 	// Process Discord message
// 	discordMsg := Message{
// 		Platform:  "discord",
// 		AuthorID:  "user123",
// 		Content:   "!setpref risk_tolerance 0.3",
// 		Timestamp: time.Now(),
// 	}

// 	sm.ProcessMessage(context.Background(), discordMsg)

// 	// Get current preferences for agent execution
// 	prefs, _ := sm.GetAggregatedPreferences(context.Background())

// 	// Execute task with stakeholder preferences
// 	agent.ExecuteTask(Task{
// 		Type: "trading",
// 		Params: map[string]interface{}{
// 			"asset": "BTC",
// 			"size":  1000,
// 		},
// 	}, prefs)
// }

// Preference extractor for different message formats
type PreferenceExtractor struct {
	patterns map[string][]string
}

func (pe *PreferenceExtractor) Extract(content string) (map[string]interface{}, error) {
	prefs := make(map[string]interface{})
	// Implement me

	return prefs, nil
}
