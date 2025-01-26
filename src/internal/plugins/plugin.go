package plugins

import "context"

// Plugin defines behavior for extending D.A.T.A
// Hook Event Summary:
// 1. BeforeAnalysis: Invoked before the cognitive analysis starts.
//    - Purpose: Modify or inspect the input data before analysis.
//    - Input: `Input` (raw user input).
// 2. AfterAnalysis: Invoked after cognitive analysis completes.
//    - Purpose: Add metadata or post-process the analysis results.
//    - Input: `*Analysis` (analysis object generated from input).
// 3. FilterAction: Invoked during action evaluation to filter out invalid or unwanted actions.
//    - Purpose: Enable or disable actions based on the current analysis.
//    - Input: `*Action`, `*Analysis` (current action and related analysis).
//    - Output: `bool` (allow or reject action), `error`.
// 4. BeforeActionExecution: Invoked before an action is executed.
//    - Purpose: Prepare or modify the action before execution.
//    - Input: `*Action` (the action about to be executed).
// 5. AfterActionExecution: Invoked after an action is executed.
//    - Purpose: Inspect or log the result of the executed action.
//    - Input: `*Action`, `*ActionResult` (executed action and its result).

// Plugin interface
type Plugin interface {
	Initialize(ctx context.Context) error
	Execute(ctx context.Context, event string, data interface{}) error
	Name() string
}
