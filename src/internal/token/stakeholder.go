package token

import (
	"context"
	"math/big"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
)

// StakeholderManager manages stakeholder interactions and influences
type StakeholderManager struct {
	tokenManager  *TokenManager
	memoryManager memory.Manager
	messageParser *MessageParser
}

// StakeholderInput represents processed input from social media
type StakeholderInput struct {
	StakeholderID string
	Platform      string // "twitter", "discord"
	MessageID     string
	Content       string
	Timestamp     time.Time
	TokenBalance  *big.Int
	Intent        InputIntent
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

// StakeholderState maintains current stakeholder status
type StakeholderState struct {
	ID           string
	TokenBalance *big.Int
	Reputation   float64
	ActiveVotes  []Vote
	Preferences  map[string]Preference
	LastUpdated  time.Time
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
func (sm *StakeholderManager) ProcessMessage(ctx context.Context, msg Message) error {
	// 1. Parse message to understand intent
	input, err := sm.messageParser.Parse(msg)
	if err != nil {
		return err
	}

	// 2. Validate stakeholder token holdings
	holdings, err := sm.tokenManager.GetHoldings(input.StakeholderID)
	if err != nil {
		return err
	}
	input.TokenBalance = holdings

	// 3. Update stakeholder state
	if err := sm.updateStakeholderState(ctx, input); err != nil {
		return err
	}

	// 4. Notify affected agents
	if err := sm.notifyAgents(ctx, input); err != nil {
		return err
	}

	return nil
}

// updateStakeholderState persists changes from new input
func (sm *StakeholderManager) updateStakeholderState(ctx context.Context, input StakeholderInput) error {
	// Get current state
	state, err := sm.getState(ctx, input.StakeholderID)
	if err != nil {
		state = NewStakeholderState(input.StakeholderID)
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
	return sm.saveState(ctx, state)
}

// GetAggregatedPreferences gets current preferences weighted by stake
func (sm *StakeholderManager) GetAggregatedPreferences(ctx context.Context) (map[string]interface{}, error) {
	// Get all stakeholder states
	states, err := sm.getAllStates(ctx)
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

// Example message parser for social media inputs
type MessageParser struct {
	// nlp           *NLPProcessor
	// intentModel   *IntentClassifier
	prefExtractor *PreferenceExtractor
}

func NewMessageParser() *MessageParser {
}

func (mp *MessageParser) Parse(msg Message) (StakeholderInput, error) {
	// Use NLP to understand message
	entities, err := mp.nlp.ExtractEntities(msg.Content)
	if err != nil {
		return StakeholderInput{}, err
	}

	// Classify intent
	intent, err := mp.intentModel.Classify(msg.Content)
	if err != nil {
		return StakeholderInput{}, err
	}

	// Extract preferences
	prefs, err := mp.prefExtractor.Extract(msg.Content, intent)
	if err != nil {
		return StakeholderInput{}, err
	}

	return StakeholderInput{
		StakeholderID: msg.AuthorID,
		Platform:      msg.Platform,
		MessageID:     msg.ID,
		Content:       msg.Content,
		Timestamp:     msg.Timestamp,
		Intent:        intent,
		Preferences:   prefs,
	}, nil
}

// Database schema for stakeholder persistence
const createStakeholderTableSQL = `
CREATE TABLE IF NOT EXISTS stakeholders (
	id TEXT PRIMARY KEY,
	token_balance NUMERIC,
	reputation FLOAT,
	preferences JSONB,
	last_updated TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stakeholder_inputs (
	id TEXT PRIMARY KEY,
	stakeholder_id TEXT,
	platform TEXT,
	message_id TEXT,
	content TEXT,
	intent JSONB,
	preferences JSONB,
	timestamp TIMESTAMP,
	FOREIGN KEY (stakeholder_id) REFERENCES stakeholders(id)
);
`

// Example usage with social media integration
func ExampleStakeholderSystem() {
	// Initialize system
	sm := NewStakeholderManager(db, tokenManager)

	// Process Twitter message
	twitterMsg := Message{
		Platform:  "twitter",
		AuthorID:  "user123",
		Content:   "The trading agent should be more conservative with stop losses set at 2%",
		Timestamp: time.Now(),
	}

	sm.ProcessMessage(context.Background(), twitterMsg)

	// Process Discord message
	discordMsg := Message{
		Platform:  "discord",
		AuthorID:  "user123",
		Content:   "!setpref risk_tolerance 0.3",
		Timestamp: time.Now(),
	}

	sm.ProcessMessage(context.Background(), discordMsg)

	// Get current preferences for agent execution
	prefs, _ := sm.GetAggregatedPreferences(context.Background())

	// Execute task with stakeholder preferences
	agent.ExecuteTask(Task{
		Type: "trading",
		Params: map[string]interface{}{
			"asset": "BTC",
			"size":  1000,
		},
	}, prefs)
}

// Preference extractor for different message formats
type PreferenceExtractor struct {
	patterns map[string][]string
}

func (pe *PreferenceExtractor) Extract(content string, intent InputIntent) (map[string]interface{}, error) {
	prefs := make(map[string]interface{})

	switch intent.Type {
	case "command":
		// Parse command-style messages (e.g., !setpref key value)
		prefs = pe.parseCommand(content)

	case "natural":
		// Use NLP to extract preferences from natural language
		prefs = pe.extractFromNaturalLanguage(content)

	case "feedback":
		// Convert feedback to preference adjustments
		prefs = pe.convertFeedbackToPreferences(content)
	}

	return prefs, nil
}
