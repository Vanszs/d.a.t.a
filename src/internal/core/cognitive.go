package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

// ReasoningType defines the type of reasoning we want to perform
type ReasoningType string

const (
	RTypeTaskEvaluation   ReasoningType = "task_evaluation"
	RTypeActionGeneration ReasoningType = "action_generation"
	RTypeVerification     ReasoningType = "verification"
	RTypeReflection       ReasoningType = "reflection"
)

type StepPurpose string

const (
	PurposeExploration StepPurpose = "exploration"
	PurposeAnalysis    StepPurpose = "analysis"
	PurposeReconsider  StepPurpose = "reconsider"
	PurposeRefinement  StepPurpose = "refinement"
	PurposeConcrete    StepPurpose = "concrete"
)

type CognitiveEngine struct {
	llm                llm.Client
	basePrompt         string
	maxSteps           int
	minConfidence      float64
	stakeholderManager token.StakeholderManager
	character          *characters.Character
}

type CognitiveConfig struct {
	NumIterations      int
	SamplesPerBatch    int
	MinRewardThreshold float64
	Temperature        float64
	MaxChainLength     int
	StabilityWindow    int
}

// ThoughtChain represents a sequence of reasoning steps
type ThoughtChain struct {
	InitialThought string
	Steps          []*ThoughtStep
	// Confidence      float64
	Reflection      string
	FinalConclusion string
	Timestamp       time.Time
}

// ThoughtStep represents a single step in the reasoning process
type ThoughtStep struct {
	Reasoning  *Reasoning
	Confidence float64
	// Verification         string
	Alternatives         []string
	ContributesToOutcome bool
	Purpose              StepPurpose
}

type RewardModel struct {
	accuracyWeight  float64
	coherenceWeight float64
	lengthWeight    float64
	stakingWeight   float64
}

// Reasoning represents a structured thought process
type Reasoning struct {
	// Core thought components
	Type         string
	Content      string // The actual thought content
	RawLLMOutput string // Original LLM output for analysis
	Evidence     []string
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// Alternative represents an alternative approach considered
type Alternative struct {
	Description string   // Description of the alternative
	Pros        []string // Benefits of this approach
	Cons        []string // Drawbacks of this approach
	Confidence  float64  // Confidence in this alternative
	Selected    bool     // Whether this alternative was chosen
}

func NewCognitiveEngine(llmClient llm.Client, character *characters.Character) *CognitiveEngine {
	return &CognitiveEngine{
		llm:           llmClient,
		maxSteps:      10,
		minConfidence: 0.7,
		basePrompt: `A conversation between User and Assistant. The assistant first thinks about the 
								reasoning process in the mind and then provides the answer. The reasoning process 
								and answer are enclosed within <think> </think> and <answer> </answer> tags.`,
		character: character,
	}
}

// GenerateThoughtChain creates a DeepSeek-style reasoning chain
func (e *CognitiveEngine) GenerateThoughtChain(ctx context.Context, input interface{}, prefs map[string]interface{}) (*ThoughtChain, error) {
	chain := &ThoughtChain{
		InitialThought: "",
		Steps:          make([]*ThoughtStep, 0),
		Timestamp:      time.Now(),
	}

	// Generate initial thought
	prompt := e.buildPrompt("initial_thought", input)
	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: "deepseek",
		Messages: []llm.Message{
			{Role: "system", Content: e.basePrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	chain.InitialThought = extractThinkingContent(response)

	// Generate reasoning steps
	for i := 0; i < e.maxSteps; i++ {
		// Determine appropriate step purpose based on progress
		purpose := e.determineStepPurpose(i)
		reasoning, err := e.generateReasoning(ctx, chain, input, purpose)
		if err != nil {
			return nil, err
		}

		// Generate alternatives
		alternatives, err := e.generateAlternatives(ctx, reasoning)
		if err != nil {
			return nil, err
		}

		step := &ThoughtStep{
			// Core reasoning content
			Reasoning:            reasoning,
			Alternatives:         alternatives,
			Purpose:              purpose,
			ContributesToOutcome: e.doesStepContributeToOutcome(purpose, chain),
		}

		// 3. Check for "aha moment" - potential reconsideration
		if AhaMomentDetection := e.detectAhaMoment(ctx, step, chain.Steps, step.Alternatives, prefs); AhaMomentDetection.Triggered {
			// Generate reconsideration step
			reconsideration, err := e.generateReconsideration(ctx, step, chain)
			if err != nil {
				return nil, err
			}
			step = &ThoughtStep{
				Purpose:              PurposeReconsider,
				Reasoning:            reconsideration,
				Alternatives:         alternatives,
				ContributesToOutcome: true,
			}
		}

		// Verify the reasoning
		// verification, err := e.verifySelfConsistency(ctx, reasoning)
		// if err != nil {
		// 	return nil, err
		// }
		// step.Verification = verification

		chain.Steps = append(chain.Steps, step)

		// Check if we need more steps
		if e.isConclusive(chain) {
			break
		}
	}

	return chain, nil
}

// determineStepPurpose decides appropriate purpose for current step
func (e *CognitiveEngine) determineStepPurpose(stepIndex int) StepPurpose {
	totalSteps := float64(e.maxSteps)
	progress := float64(stepIndex) / totalSteps

	switch {
	case progress < 0.3:
		return PurposeExploration
	case progress < 0.5:
		return PurposeAnalysis
	case progress < 0.7:
		return PurposeRefinement
	default:
		return PurposeConcrete
	}
}

// doesStepContributeToOutcome determines if step contributes to final actions/tasks
func (e *CognitiveEngine) doesStepContributeToOutcome(purpose StepPurpose, chain *ThoughtChain) bool {
	// Concrete steps always contribute
	if purpose == PurposeConcrete {
		return true
	}

	// Reconsideration steps that improve the solution contribute
	if purpose == PurposeReconsider {
		return true
	}

	// Late refinement steps often contribute
	if purpose == PurposeRefinement && len(chain.Steps) > 5 {
		return true
	}

	return false
}

func formatPreviousSteps(steps []*ThoughtStep) string {
	if len(steps) == 0 {
		return "No previous steps"
	}

	var formatted string
	for i, step := range steps {
		formatted += fmt.Sprintf("Step %d (%s):\n%s\n\n",
			i+1, step.Reasoning.Type, step.Reasoning.Content)
	}
	return formatted
}

// GenerateActions uses chain-of-thought for action planning
func (e *CognitiveEngine) GenerateActions(ctx context.Context, task *tasks.Task, prefs map[string]interface{}) (*ActionGeneration, error) {
	// Build action context
	actionContext := map[string]interface{}{
		"task":        task,
		"preferences": prefs,
		"goal":        "generate detailed action plan",
	}

	// Generate thought chain for action planning
	chain, err := e.GenerateThoughtChain(ctx, actionContext, prefs)
	if err != nil {
		return nil, err
	}

	// Convert thought chain to actions
	actions, _ := convertThoughtChainToActions(chain)

	return &ActionGeneration{
		Actions: actions,
		Chain:   chain,
	}, nil
}

// GenerateActions uses chain-of-thought for action planning
func (e *CognitiveEngine) GenerateMessage(ctx context.Context, input interface{}, prefs map[string]interface{}) (string, error) {
	// Build action context
	msgContext := map[string]interface{}{
		"character":   e.character,
		"preferences": prefs,
		"goal":        "generate detailed action plan",
	}

	// Generate initial thought
	prompt := e.buildPrompt("message", msgContext)
	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: "deepseek",
		Messages: []llm.Message{
			{Role: "system", Content: e.basePrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}

	return response, nil
}

// GenerateTasks uses chain-of-thought for tasks planning
func (e *CognitiveEngine) GenerateTasks(ctx context.Context, state *SystemState) (*TaskGeneration, error) {
	// Build action context
	taskContext := map[string]interface{}{
		"state": state,
		"goal":  "generate detailed tasks plan",
	}

	// Generate thought chain for action planning
	chain, err := e.GenerateThoughtChain(ctx, taskContext, state.StakeholderPreferences)
	if err != nil {
		return nil, err
	}

	// Convert thought chain to actions
	tasks, _ := convertThoughtChainToTasks(chain)

	return &TaskGeneration{
		Tasks: tasks,
		Chain: chain,
	}, nil
}

func (e *CognitiveEngine) generateReasoning(ctx context.Context, chain *ThoughtChain, input interface{}, purpose StepPurpose) (*Reasoning, error) {
	prompt := e.buildReasoningPrompt(RTypeActionGeneration, input, chain.Steps)

	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: "deepseek",
		Messages: []llm.Message{
			{Role: "system", Content: e.basePrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	reasoning := &Reasoning{}

	// Extract evidence and confidence
	thought, evidence := e.extractThoughtAndEvidence(response)
	reasoning.Content = thought
	reasoning.Evidence = evidence

	return reasoning, nil
}

func (e *CognitiveEngine) extractThoughtAndEvidence(llmResponse string) (string, []string) {
	// Extract thought content
	thought := extractThinkingContent(llmResponse)

	// Extract evidence
	// evidenceStr := extractEvidence(llmResponse)

	// // Parse evidence into structured format
	// TODO: Implement evidence parsing
	// for _, evidenceItem := range parseEvidenceItems(evidenceStr) {
	// 	evidence = append(evidence, Evidence{
	// 		// Type:        identifyEvidenceType(evidenceItem),
	// 		Description: evidenceItem,
	// 		// Weight:      calculateEvidenceWeight(evidenceItem),
	// 	})
	// }

	return thought, []string{}
}

// buildReasoningPrompt constructs appropriate prompt based on reasoning type
func (e *CognitiveEngine) buildReasoningPrompt(rType ReasoningType, input interface{}, prevSteps []*ThoughtStep) string {
	var promptTemplate string

	switch rType {
	case RTypeTaskEvaluation:
		promptTemplate = `Evaluate this task thoroughly. Consider:
1. Key objectives and requirements
2. Potential challenges and risks
3. Required resources and dependencies
4. Success criteria and metrics
5. Stakeholder implications

Previous steps to consider:
%s

Task details:
%v

Format your response as:
<think>
[Your thorough analysis here]
</think>
<evidence>
[List specific evidence supporting your analysis]
</evidence>`

	case RTypeActionGeneration:
		promptTemplate = `Generate detailed action steps. Consider:
1. Concrete implementable actions
2. Dependencies and ordering
3. Resource requirements
4. Risk mitigation steps
5. Success indicators

Previous reasoning:
%s

Context:
%v

Format your response as:
<think>
[Your action plan here]
</think>
<evidence>
[Supporting evidence for each action]
</evidence>`

	case RTypeVerification:
		promptTemplate = `Verify this reasoning step critically. Consider:
1. Logical consistency
2. Evidence validity
3. Assumption validation
4. Edge cases and failure modes
5. Alternative perspectives

Step to verify:
%s

Format your response as:
<think>
[Your verification analysis]
</think>
<evidence>
[Verification evidence]
</evidence>`

	case RTypeReflection:
		promptTemplate = `Reflect on this reasoning process. Consider:
1. Key insights and learnings
2. Areas for improvement
3. Unexpected discoveries
4. Alternative approaches
5. Future implications

Process to reflect on:
%s

Format your response as:
<think>
[Your reflection]
</think>
<evidence>
[Supporting observations]
</evidence>`
	}

	// Format previous steps if any
	prevStepsStr := formatPreviousSteps(prevSteps)

	return fmt.Sprintf(promptTemplate, prevStepsStr, input)
}

func (e *CognitiveEngine) generateReconsideration(ctx context.Context, step *ThoughtStep, chain *ThoughtChain) (*Reasoning, error) {
	// Build reconsideration prompt
	prompt := fmt.Sprintf(`Let's reconsider our approach. 

Current thinking:
%s

Previous steps:
%s

Let's reevaluate step-by-step:
1. What assumptions might we be making?
2. What alternative approaches could we consider?
3. Are we missing any important perspectives?

Format your response as:
<think>
Reconsideration analysis...
</think>`, step.Reasoning, formatPreviousSteps(chain.Steps))

	// Generate reconsideration
	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: "deepseek",
		Messages: []llm.Message{
			{Role: "system", Content: e.basePrompt},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create reconsideration reasoning
	reasoning := &Reasoning{
		Type:      "reflection",
		Content:   extractThinkingContent(response),
		Timestamp: time.Now(),
	}

	// Add markers for reconsideration
	reasoning.Metadata = map[string]interface{}{
		"is_reconsideration": true,
		"original_step":      step.Reasoning,
	}

	return reasoning, nil
}

func (e *CognitiveEngine) buildPrompt(promptType string, context interface{}) string {
	switch promptType {
	case "initial_thought":
		return fmt.Sprintf(`Given the following context, let's think through this step by step:
					Context: %v
					Think about the key aspects, potential challenges, and possible approaches.
					Format your response as:
					<think>
					Initial analysis...
					Key considerations...
					Potential approaches...
					</think>`, context)
	case "reflection":
		return fmt.Sprintf(`Review the following reasoning chain and provide a critical reflection:
					Chain: %v
					Consider:
					1. Are there gaps in the reasoning?
					2. What assumptions were made?
					3. Are there alternative approaches worth considering?
					4. What are the potential risks?`, context)
	default:
		return fmt.Sprintf("Let's analyze: %v", context)
	}
}

// verifySelfConsistency checks the logical consistency of a reasoning step
// func (e *CognitiveEngine) verifySelfConsistency(ctx context.Context, reasoning *Reasoning) (string, error) {
// 	verificationPrompt := fmt.Sprintf(`Verify the following reasoning step, considering logical consistency and potential flaws:

// Reasoning: %s
// Evidence provided: %s

// Follow these verification steps:
// 1. Check logical consistency
// 2. Validate evidence
// 3. Identify potential flaws or gaps
// 4. Consider edge cases

// Format your response in <think> tags for verification process and <answer> tags for the result.`,
// 		reasoning.Content, strings.Join(reasoning.Evidence, ", "))

// 	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
// 		Model: "deepseek",
// 		Messages: []llm.Message{
// 			{Role: "system", Content: e.basePrompt},
// 			{Role: "user", Content: verificationPrompt},
// 		},
// 	})
// 	if err != nil {
// 		return "", fmt.Errorf("verification completion failed: %w", err)
// 	}

// 	// Extract verification results
// 	thinkingProcess := extractThinkingContent(response)
// 	answer := extractAnswerContent(response)

// 	// Identify potential issues
// 	issues := e.identifyLogicalIssues(thinkingProcess)
// 	if len(issues) > 0 {
// 		step.Confidence *= 0.8 // Reduce confidence if issues found
// 	}

// 	// Generate verification summary
// 	verification := fmt.Sprintf(`Verification Results:
// Consistency Check: %s
// Issues Identified: %s
// Confidence Level: %.2f`,
// 		answer,
// 		strings.Join(issues, ", "),
// 		step.Confidence)

// 	return verification, nil
// }

// generateAlternatives produces alternative approaches for a given reasoning step
func (e *CognitiveEngine) generateAlternatives(ctx context.Context, step *Reasoning) ([]string, error) {
	alternativePrompt := fmt.Sprintf(`Consider alternative approaches to this reasoning step:

Current approach:
%s

Generate 2-3 fundamentally different approaches that could achieve the same goal.
Consider:
1. Different methodologies
2. Alternative perspectives
3. Novel solutions
4. Stakeholder preferences

Format each alternative in <think> tags with its rationale and potential benefits/drawbacks.`,
		step.Content)

	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: "deepseek",
		Messages: []llm.Message{
			{Role: "system", Content: e.basePrompt},
			{Role: "user", Content: alternativePrompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("alternative generation failed: %w", err)
	}

	// Extract and analyze alternatives
	alternatives := parseAlternatives(response)

	// Evaluate each alternative
	var validatedAlternatives []string
	for _, alt := range alternatives {
		score := e.evaluateAlternative(alt)
		if score >= e.minConfidence {
			validatedAlternatives = append(validatedAlternatives, alt)
		}
	}

	return validatedAlternatives, nil
}

// isConclusive determines if the reasoning chain has reached a satisfactory conclusion
func (e *CognitiveEngine) isConclusive(chain *ThoughtChain) bool {
	// Check minimum confidence threshold
	// if chain.Confidence < e.minConfidence {
	// 	return false
	// }

	// Must have at least one step
	if len(chain.Steps) == 0 {
		return false
	}

	// Check if we've addressed all key aspects
	// aspectsCovered := e.checkAspectsCoverage(chain)
	// if !aspectsCovered {
	// 	return false
	// }

	// Verify last step completion
	lastStep := chain.Steps[len(chain.Steps)-1]
	if lastStep.Purpose != PurposeConcrete {
		return false
	}

	// Calculate overall reasoning completeness
	// completeness := e.calculateCompleteness(chain)
	return true
}

// Helper functions

func (e *CognitiveEngine) identifyLogicalIssues(thinking string) []string {
	var issues []string

	// Common logical fallacies and issues to check
	checks := map[string]string{
		"circular_reasoning":   `\b(because.*therefore.*because|therefore.*because.*therefore)\b`,
		"false_assumption":     `\b(must|always|never|everyone|nobody)\b`,
		"causal_fallacy":       `\b(leads to|causes|results in)\b`,
		"hasty_generalization": `\b(all|none|every|no one)\b`,
	}

	for issueType, pattern := range checks {
		if strings.Contains(thinking, pattern) {
			issues = append(issues, issueType)
		}
	}

	return issues
}

func (e *CognitiveEngine) evaluateAlternative(alternative string) float64 {
	var score float64 = 1.0

	// Evaluate completeness
	if !strings.Contains(alternative, "benefits") || !strings.Contains(alternative, "drawbacks") {
		score *= 0.8
	}

	// Check for concrete steps
	if !containsConcreteSteps(alternative) {
		score *= 0.7
	}

	// Assess feasibility
	if !assessFeasibility(alternative) {
		score *= 0.6
	}

	return score
}

// func (e *CognitiveEngine) checkAspectsCoverage(chain *ThoughtChain) bool {
// 	requiredAspects := []string{
// 		"problem_definition",
// 		"methodology",
// 		"validation",
// 		"risks",
// 		"outcomes",
// 	}

// 	coveredAspects := make(map[string]bool)

// 	// Check each step for aspect coverage
// 	for _, step := range chain.Steps {
// 		for _, aspect := range requiredAspects {
// 			if containsAspect(step.Reasoning.Content, aspect) {
// 				coveredAspects[aspect] = true
// 			}
// 		}
// 	}

// 	// All required aspects must be covered
// 	return len(coveredAspects) == len(requiredAspects)
// }

// func (e *CognitiveEngine) calculateCompleteness(chain *ThoughtChain) float64 {
// 	// Factors contributing to completeness
// 	factors := map[string]float64{
// 		"step_coverage":         e.calculateStepCoverage(chain),
// 		"verification":          e.calculateVerificationScore(chain),
// 		"confidence":            chain.Confidence,
// 		"alternatives":          float64(len(chain.Alternatives)) / 3.0, // Normalize to 0-1
// 		"stakeholder_alignment": e.calculateStakeholderAlignment(chain),
// 	}

// 	// Weights for each factor
// 	weights := map[string]float64{
// 		"step_coverage":         0.3,
// 		"verification":          0.2,
// 		"confidence":            0.2,
// 		"alternatives":          0.1,
// 		"stakeholder_alignment": 0.2,
// 	}

// 	var completeness float64
// 	for factor, score := range factors {
// 		completeness += score * weights[factor]
// 	}

// 	return math.Min(1.0, completeness)
// }

// func (e *CognitiveEngine) calculateStepCoverage(chain *ThoughtChain) float64 {
// 	coveredSteps := 0
// 	for _, step := range chain.Steps {
// 		if e.isStepComplete(step) {
// 			coveredSteps++
// 		}
// 	}
// 	return float64(coveredSteps) / float64(len(chain.Steps))
// }

// func (e *CognitiveEngine) isStepComplete(step Reasoning) bool {
// 	return step.Confidence >= e.minConfidence &&
// 		len(step.Evidence) > 0 &&
// 		step.Verification != ""
// }

// func (e *CognitiveEngine) calculateVerificationScore(chain *ThoughtChain) float64 {
// 	if len(chain.Verification) == 0 {
// 		return 0.0
// 	}

// 	var totalScore float64
// 	for _, v := range chain.Verification {
// 		if strings.Contains(v, "Consistency Check: Pass") {
// 			totalScore += 1.0
// 		} else if strings.Contains(v, "Consistency Check: Partial") {
// 			totalScore += 0.5
// 		}
// 	}

// 	return totalScore / float64(len(chain.Verification))
// }

// Utility functions

func parseAlternatives(response string) []string {
	// Extract alternatives between <think> tags
	alternatives := make([]string, 0)

	// Split response by <think> tags
	parts := strings.Split(response, "<think>")
	for _, part := range parts[1:] { // Skip first empty part
		if idx := strings.Index(part, "</think>"); idx != -1 {
			alt := strings.TrimSpace(part[:idx])
			alternatives = append(alternatives, alt)
		}
	}

	return alternatives
}

func containsConcreteSteps(alternative string) bool {
	// Check for numbered steps or action words
	return strings.Contains(alternative, "1.") ||
		strings.Contains(alternative, "First") ||
		strings.Contains(alternative, "Initially") ||
		strings.Contains(alternative, "Start by")
}

func assessFeasibility(alternative string) bool {
	// Check for implementation details and resource considerations
	return strings.Contains(alternative, "implement") ||
		strings.Contains(alternative, "resource") ||
		strings.Contains(alternative, "require") ||
		strings.Contains(alternative, "need")
}

func containsAspect(step string, aspect string) bool {
	aspectPatterns := map[string][]string{
		"problem_definition": {"problem", "challenge", "objective", "goal"},
		"methodology":        {"method", "approach", "strategy", "process"},
		"validation":         {"verify", "validate", "check", "confirm"},
		"risks":              {"risk", "challenge", "issue", "concern"},
		"outcomes":           {"result", "outcome", "impact", "effect"},
	}

	patterns, exists := aspectPatterns[aspect]
	if !exists {
		return false
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(step), pattern) {
			return true
		}
	}
	return false
}

// Helper functions
func extractThinkingContent(response string) string {
	// TODO: implement me
	// Extract content between <think> tags
	// Implementation details...
	return ""
}

func extractEvidence(response string) []string {
	// TODO: implement me
	// Extract evidence from response
	// Implementation details...
	return nil
}

func extractAnwser(response string) []string {
	// TODO: implement me
	// Extract evidence from response
	// Implementation details...
	return nil
}

func calculateConfidence(response string) float64 {
	// TODO: implement me
	// Calculate confidence based on response
	// Implementation details...
	return 0.0
}

func generateAlternativeApproach(chain *ThoughtChain) string {
	// TODO: implement me
	// Generate alternative approach based on current chain
	// Implementation details...
	return ""
}
