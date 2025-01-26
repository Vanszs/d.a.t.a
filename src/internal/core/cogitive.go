// internal/core/cognitive.go

package core

import (
	"context"
	"data-agent/internal/data"
	"data-agent/pkg/llm"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
)

// CognitiveEngine handles the core decision-making process
type CognitiveEngine struct {
	llm       llm.Client
	evaluator *Evaluator
	context   *CognitiveContext
}

// CognitiveContext maintains the current state for decision making
type CognitiveContext struct {
	Goals      []Goal
	DataList   []data.Data
	TokenState TokenState
	Memory     memory.WorkingMemory
	Data       []data.Data
	Timestamp  time.Time
}

type Analysis struct {
	Understanding string
	Intent        Intent
	Relevance     map[string]float64
	Actions       []Action
	Confidences   []float64
}

type Intent struct {
	Type        string // e.g., "question", "proposal", "feedback"
	Priority    float64
	Stakeholder data.StakeholderInfo
	GoalImpact  map[string]float64
}

type Suggestion struct {
	Type         string
	Action       string
	Rationale    string
	Priority     float64
	Requirements map[string]interface{}
}

// Analyze processes input and generates comprehensive analysis
func (c *CognitiveEngine) Analyze(ctx context.Context, input Input, identity data.IdentityInfo) (Analysis, error) {
	// 1. Build analysis context
	analysisCtx, err := c.buildAnalysisContext(input, identity)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to build analysis context: %w", err)
	}

	// 2. Understand input intent
	understanding, err := c.understandInput(ctx, input, analysisCtx)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to understand input: %w", err)
	}

	// 3. Evaluate stakeholder influence
	stakeholderInfo, err := c.evaluateStakeholder(ctx, identity, understanding)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to evaluate stakeholder: %w", err)
	}

	// 4. Assess goal relevance
	relevance, err := c.assessGoalRelevance(ctx, understanding, stakeholderInfo)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to assess goal relevance: %w", err)
	}

	// 5. Generate suggestions
	suggestions, err := c.generateSuggestions(ctx, understanding, stakeholderInfo, relevance)
	if err != nil {
		return Analysis{}, fmt.Errorf("failed to generate suggestions: %w", err)
	}

	// 6. Calculate confidence
	confidence := c.calculateConfidence(understanding, stakeholderInfo, suggestions)

	return Analysis{
		Understanding: understanding.Summary,
		Intent: Intent{
			Type:        understanding.Type,
			Priority:    c.calculatePriority(stakeholderInfo, understanding),
			Stakeholder: stakeholderInfo,
			GoalImpact:  relevance,
		},
		Relevance:   relevance,
		Suggestions: suggestions,
		Confidence:  confidence,
	}, nil
}

// buildAnalysisContext gathers all relevant information for analysis
func (c *CognitiveEngine) buildAnalysisContext(input Input, identity data.IdentityInfo) (llm.Context, error) {
	// Combine multiple sources of context
	context := llm.Context{
		CurrentGoals:    c.context.Goals,
		TokenState:      c.context.TokenState,
		RecentMemory:    c.context.Memory.Recent(),
		StakeholderInfo: identity,
		MarketContext:   c.context.MarketData,
	}

	return context, nil
}

// understandInput uses LLM to comprehend input meaning and intent
func (c *CognitiveEngine) understandInput(ctx context.Context, input Input, analysisCtx llm.Context) (Understanding, error) {
	prompt := c.generateUnderstandingPrompt(input, analysisCtx)

	response, err := c.llm.GenerateText(ctx, llm.State, prompt)
	if err != nil {
		return Understanding{}, err
	}

	return c.parseUnderstanding(response)
}

// evaluateStakeholder assesses stakeholder influence and history
func (c *CognitiveEngine) evaluateStakeholder(ctx context.Context, identity data.IdentityInfo, understanding Understanding) (data.StakeholderInfo, error) {
	// Get token holdings
	holdings, err := c.context.TokenState.GetHoldings(identity.Address)
	if err != nil {
		return data.StakeholderInfo{}, err
	}

	// Get interaction history
	history, err := c.context.Memory.GetStakeholderHistory(identity.ID)
	if err != nil {
		return data.StakeholderInfo{}, err
	}

	return data.StakeholderInfo{
		Identity:    identity,
		Holdings:    holdings,
		History:     history,
		Influence:   c.calculateInfluence(holdings, history),
		Preferences: c.inferPreferences(history),
	}, nil
}

// assessGoalRelevance evaluates how input relates to current goals
func (c *CognitiveEngine) assessGoalRelevance(ctx context.Context, understanding Understanding, stakeholder data.StakeholderInfo) (map[string]float64, error) {
	relevance := make(map[string]float64)

	prompt := c.generateRelevancePrompt(understanding, stakeholder, c.context.Goals)

	response, err := c.llm.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return c.parseRelevanceScores(response)
}

// generateSuggestions creates action suggestions based on analysis
func (c *CognitiveEngine) generateSuggestions(ctx context.Context, understanding Understanding, stakeholder data.StakeholderInfo, relevance map[string]float64) ([]Suggestion, error) {
	prompt := c.generateSuggestionsPrompt(understanding, stakeholder, relevance)

	response, err := c.llm.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	suggestions := c.parseSuggestions(response)

	// Filter suggestions based on governance rules
	return c.filterSuggestions(suggestions, stakeholder), nil
}

// calculateConfidence determines confidence level in analysis
func (c *CognitiveEngine) calculateConfidence(understanding Understanding, stakeholder data.StakeholderInfo, suggestions []Suggestion) float64 {
	// Consider multiple factors for confidence
	factors := map[string]float64{
		"understanding_clarity": understanding.Confidence,
		"stakeholder_history":   c.getStakeholderConfidence(stakeholder),
		"suggestion_consensus":  c.getSuggestionConsensus(suggestions),
		"data_quality":          c.getDataQuality(),
	}

	// Weight and combine factors
	return c.calculateWeightedScore(factors)
}
