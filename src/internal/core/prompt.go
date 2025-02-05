package core

import (
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
)

func getGeneralInfo(systemState *SystemState) string {
	return fmt.Sprintf(`
	You are an agent **%s**. Here are your basic information:
	### **Agent Information**
	- **Description**: %s
	- **Primary Goals**: %s
	- **Stakeholder Preferences**: %s

	Here are your available tools:
	### **Available Tools**
	The following tools are available to the AI Agent:
	%s
	Each tool has specific capabilities. When generating tasks, consider how these tools can be leveraged. You shouldn't create tasks that can't be fullfilled by the given tools.`,
		systemState.Character.Name,
		systemState.Character.System,
		convertGoalsToString(systemState.AgentStates.Goals),
		formatMap(systemState.StakeholderPreferences),
		formatTools(systemState.AvailableTools),
	)
}

func generateTasksPromptFunc(systemState *SystemState) promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
		switch stepPurpose {
		case PurposeInitial:
			return fmt.Sprintf(`
%s

I need you to generate some tasks for yourself to help you achieve your goals.
These tasks should be actionable, strategically valuable, and scalable. Consider the available tools and resources when creating tasks.
The tasks should align with your primary goals and stakeholder preferences.
The tasks should be excutable by the tools available to you.

### **Task Generation Guidelines**
1. **Strategic Alignment**: Tasks should contribute directly to the **core objectives** of the agent.
2. **Situational Awareness**: Consider the **current system state**, adapting tasks to evolving conditions.
3. **Stakeholder Relevance**: Ensure tasks **align with preferences** and expectations.
4. **Variety and Coverage**: Generate **different alternatives**, from analytical to operational.
5. **Tools limitation**: Generate tasks that can be executed with the current set of tools available.

### **Result Format**
Structure your response as follows:

<think>
- **Task Name**: [Concise, action-oriented title]
- **Objective**: [What is the purpose of this task?]
- **Expected Impact**: [How does this improve the system or benefit stakeholders?]
- **Key Considerations**: [Challenges, dependencies, or required data]
</think>

<evidence>
[List specific evidence supporting your analysis]
</evidence>

<alternatives>
[List of alternative approaches]
</alternatives>

### **Output Requirements**
- Provide **at least 2-4** distinct tasks.
- Ensure tasks are **actionable** and **strategically valuable**.

Now, generate the most relevant and impactful tasks for **%s**.`,
				getGeneralInfo(systemState),
				systemState.Character.Name,
			)
		case PurposeAnalysis:
			// Purpose Analysis: Evaluate the tasks that have been generated to assess their feasibility, risks, and alignment with goals.
			return fmt.Sprintf(`
We have identified the following potential tasks:

%s

Now, let's evaluate these tasks in detail based on:
1. **Strategic Alignment**: Does each task directly contribute to the agent's core goals? 
2. **Feasibility**: Can these tasks be realistically accomplished with the available resources and data?
3. **Risk and Challenges**: What are the risks associated with each task? Are there any dependencies or obstacles?
4. **Stakeholder Impact**: How do these tasks align with stakeholder preferences and expectations?
5. **Tools limitation**: Are these tasks feasible with the current set of tools available?

%s

### **Task Evaluation Format**
For each task, provide the following evaluation:

**<think>**
- **Task Name**: [Task being analyzed]
- **Strategic Alignment**: [Does this align with the core objectives?]
- **Feasibility**: [Is this achievable within the available resources?]
- **Risk and Challenges**: [Identify potential issues]
- **Stakeholder Impact**: [How will stakeholders be affected?]
**</think>

Evaluate all tasks thoroughly and determine their **suitability** for further refinement.
`,
				formatPreviousSteps(steps),
				getGeneralInfo(systemState),
			)
		case PurposeReconsider:
			return fmt.Sprintf(`
Let's **reconsider** our current approach carefully. We will evaluate the current reasoning and explore whether there might be better alternatives or improvements.

### **Previous Steps:**
%s

### **Reevaluation Questions:**
1. **Assumptions Check**: What **assumptions** are we making, and how can we validate them? Could these assumptions be limiting our options?
2. **Alternative Approaches**: Are there **other approaches** that might be more effective or efficient? What are they?
3. **Stakeholder Considerations**: Have we considered the **stakeholder needs** and **preferences** in the current approach? What feedback might we have missed?
4. **New Insights**: Is there any **new information** that could change our perspective or approach?
5. **Risk Assessment**: Are there any **risks** we've overlooked, or should we consider more **robust mitigation** strategies?
5. **Tools limitation**: Are these tasks feasible with the current set of tools available?

%s

### **Thought Process:**
Format your response as follows:
**<think>**
- **Reconsideration Analysis**: [A critical reflection on the current approach]
- **Identified Assumptions**: [List the assumptions and why they might need validation or change]
- **Alternative Approaches**: [Describe any alternative solutions]
- **Stakeholder Impact**: [How would each alternative affect stakeholders?]
- **Risks and Mitigation**: [What risks do we face with the current approach?]
</think>

Please provide a **comprehensive reconsideration** of the current approach and suggest **new strategies** that might be more aligned with the goal.`,
				formatPreviousSteps(steps),
				getGeneralInfo(systemState),
			)
		case PurposeRefinement:
			// Purpose Refinement: Improve and polish the tasks based on analysis and feedback.
			return fmt.Sprintf(`
Let's refine the tasks based on the analysis.

### **Previous Steps:**
%s

### **Refinement Questions:**
1. **Clarity and Focus**: Are the tasks clearly defined with a specific, actionable goal?
2. **Prioritization**: Which tasks should be prioritized based on their potential impact and feasibility?
3. **Efficiency**: Can the tasks be broken down into smaller, more manageable steps?
4. **Stakeholder Consideration**: Are there any further adjustments needed to meet stakeholder preferences?
5. **Tools limitation**: Are these tasks feasible with the current set of tools available?

%s

### **Refined Task Format**
For each task, provide a detailed refinement:

**<think>**
- **Task Name**: [Refined task title]
- **Objective**: [What is the clear and actionable goal?]
- **Execution Plan**: [Break down the task into actionable steps]
- **Priority**: [High / Medium / Low]
- **Stakeholder Alignment**: [How does this meet stakeholder needs?]
- **Tools limitation**: [Can this task be executed with the current tools?]
**</think>

Refine the tasks, making them **clearer, actionable, and aligned with the overall goals**.
`,
				formatPreviousSteps(steps),
				getGeneralInfo(systemState),
			)
		case PurposeConcrete:
			// Purpose Concrete: Finalize the tasks into fully executable plans with precise actions.
			return fmt.Sprintf(`
The tasks are now ready for execution. Let's select the most promising tasks and create details.

### **Finalization Steps:**
1. **Actionability**: Ensure each task can be executed with a clear step-by-step plan.
2. **Responsibility Assignment**: Assign tasks to specific agents or systems responsible for execution.
3. **Resources**: Ensure all necessary resources (e.g., data, tools) are available to carry out the tasks.
4. **Timeline**: Define clear **deadlines** or **milestones** for each task.
5. **Tools and Dependencies**: Identify any existing tools or dependencies required for task execution.

Previous Steps:
%s

%s

### **Task Format**
Create a final version of task. Please generate a json format result for the task in the below Task strucuture:

type Task struct {
	ID                       string
	Name                     string
	Description              string
	Priority                 float64
	ExecutionSteps           []string
	Status                   TaskStatus
	Deadline                 *time.Time
	RequiresApproval         bool
	Tools 									 []string
	RequiresStakeholderInput bool
	CreatedBy                string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

Please wrap the JSON format of the final task in the tag <json> and </json>.
**<think>**
- **JSON format of the final task**: [The final task for execution]
**</think>**

Finalize the task into **Task structure**.
`,
				formatPreviousSteps(steps),
				getGeneralInfo(systemState),
			)
		}
		return ""
	}
}

func generateActionsPromptFunc(systemState *SystemState, task *tasks.Task, actions []actions.Action) promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
		switch stepPurpose {
		case PurposeInitial:
			// Initial Action Generation
			actionDescriptions := ""
			for _, action := range actions {
				actionDescriptions += fmt.Sprintf("\n- **%s**: %s", action.Name(), action.Description())
			}

			return fmt.Sprintf(`
		You are an AI action planning assistant for the agent **%s**. Your goal is to generate a set of **high-level actions** to achieve the tasks defined for the agent.
		
		### **Agent Information**
		- **Description**: %s
		- **Primary Goals**: %s
		- **Stakeholder Preferences**: %s
		
		### **Available Tools**
		The following tools are available to the AI Agent:
		%s
		
		### **Action Generation Guidelines**
		1. **Strategic Alignment**: Actions should directly contribute to **achieving the high-level tasks**.
		2. **Situational Awareness**: Consider the **current system state**, adapting actions to evolving conditions.
		3. **Stakeholder Relevance**: Ensure actions are **aligned with preferences** and expectations.
		4. **Feasibility**: Consider the **capabilities of the available tools** in the action design.
		5. **Variety and Coverage**: Generate a **wide range of alternative actions** for each task.
		
		### **Result Format**
		For each action, structure your response as follows:
		
		<think>
		- **Action Name**: [Concise, action-oriented title]
		- **Objective**: [What is the purpose of this action?]
		- **Expected Outcome**: [What result will this action achieve?]
		- **Tools Required**: [What tools will be needed for this action?]
		- **Dependencies**: [What data or actions must precede this one?]
		**</think>**
		
		<evidence>
		[List specific evidence or reasoning supporting the action]
		</evidence>
		
		<alternatives>
		[List alternative approaches for achieving the same goal]
		</alternatives>
		
		### **Output Requirements**
		- Provide **at least 5-7** distinct, high-level actions.
		- Ensure actions are **feasible**, **strategically valuable**, and **scalable**.
		
		Now, generate the most relevant and impactful actions for **%s**.`,
				systemState.Character.Name,
				systemState.Character.System,
				convertGoalsToString(systemState.AgentStates.Goals),
				formatMap(systemState.StakeholderPreferences),
				actionDescriptions,
				systemState.Character.Name,
			)

		case PurposeExploration:
			// Exploration: Generate many potential action ideas, even unconventional ones
			return fmt.Sprintf(`
		Let's explore different possible **actions** for the task:
		
		### **Previous Steps:**
		%s
		
		Think creatively and list as many possible **high-level actions** as possible, even if they seem unconventional. Consider:
		1. **Actions that align with the core task objectives**.
		2. **Actions that make use of the available tools**.
		3. **Actions that could address stakeholder needs**.
		4. **Unconventional approaches** that might be effective.
		
		### **Action Format**
		For each action, structure your response as follows:
		
		<think>
		- **Action Name**: [Concise, action-oriented title]
		- **Objective**: [What is the purpose of this action?]
		- **Expected Outcome**: [How will this contribute to the task goal?]
		- **Tools Involved**: [Which tools will be used to execute this action?]
		</think>
		
		<evidence>
		[List specific evidence or reasoning supporting the action]
		</evidence>
		
		<alternatives>
		[List other possible approaches]
		</alternatives>
		
		Provide **at least 7-10** action ideas that could be explored further.
		`,
				formatPreviousSteps(steps),
			)

		case PurposeAnalysis:
			// Analysis: Evaluate the feasibility and impact of the actions generated
			return fmt.Sprintf(`
		We have identified the following potential actions:
		
		%s
		
		Let's analyze each action for **feasibility**, **alignment with goals**, and **impact**. Consider the following:
		1. **Strategic Alignment**: Does this action contribute to the task's overall goal?
		2. **Feasibility**: Is this action achievable given the available tools and resources?
		3. **Risk and Challenges**: What are the potential risks associated with this action?
		4. **Stakeholder Impact**: How does this action align with stakeholder preferences and priorities?
		
		### **Action Evaluation Format**
		For each action, provide the following analysis:
		
		**<think>**
		- **Action Name**: [Action being analyzed]
		- **Strategic Alignment**: [Does this align with the task’s core objectives?]
		- **Feasibility**: [Is this achievable with current tools and resources?]
		- **Risk and Challenges**: [What risks should be mitigated?]
		- **Stakeholder Impact**: [How will stakeholders be affected?]
		**</think>
		
		Evaluate each action based on its **alignment**, **feasibility**, and **impact**.
		`,
				formatPreviousSteps(steps),
			)

		case PurposeReconsider:
			// Reconsider: Reflect on the actions and suggest improvements
			return fmt.Sprintf(`
		Let's **reconsider** the actions we have generated.
		
		### **Previous Actions:**
		%s
		
		Evaluate each action to determine if:
		1. **Alternative approaches** could be more efficient.
		2. **Improvement opportunities** exist (e.g., breaking down complex actions into smaller steps).
		3. The actions should be **reprioritized** based on updated insights.
		
		### **Reconsideration Questions:**
		1. **What assumptions are we making** about the task or resources?
		2. **Are there better alternatives** that would achieve the same goal with fewer resources?
		3. **How can we improve** the efficiency of these actions?
		
		### **Reconsidered Action Format**
		For each reconsidered action, structure your response as follows:
		
		**<think>**
		- **Action Name**: [Action being reconsidered]
		- **Improvement Opportunity**: [What can be improved?]
		- **Alternative Approach**: [Describe a better alternative]
		- **Stakeholder Alignment**: [How does this alternative align with stakeholder needs?]
		</think>
		
		<evidence>
		[List any supporting evidence for the reconsidered action]
		</evidence>
		
		<alternatives>
		[List alternative approaches]
		</alternatives>
		
		Reconsider each action and suggest **improvements** or **new alternatives**.
		`,
				formatPreviousSteps(steps),
			)

		case PurposeRefinement:
			// Refinement: Polish the actions, making them clear and actionable
			return fmt.Sprintf(`
		Let's refine the actions for **clarity and effectiveness**.
		
		### **Actions to Refine:**
		%s
		
		Ensure each action is:
		1. **Clear** and **actionable** with specific steps.
		2. **Efficient**, minimizing unnecessary complexity.
		3. **Aligned with the core goals** and **stakeholder preferences**.
		
		### **Refined Action Format**
		For each action, structure your response as follows:
		
		**<think>**
		- **Action Name**: [Refined action name]
		- **Execution Plan**: [Detailed steps for execution]
		- **Resources Required**: [What tools, data, or support is needed?]
		- **Timeline**: [Define deadlines or milestones]
		</think>
		
		Refine the actions, ensuring they are **clear**, **efficient**, and **actionable**.
		`,
				formatPreviousSteps(steps),
			)

		case PurposeConcrete:
			// Concrete: Finalize the actions into **clear, executable plans**
			return fmt.Sprintf(`
		We are now finalizing the actions for **execution**.
		
		### **Final Action Format**
		For each action, generate a **finalized execution plan** with clear, actionable steps:
		
		**<think>**
		- **Action Name**: [Finalized action name]
		- **Execution Steps**: [Step-by-step breakdown of the action]
		- **Assigned To**: [Which agent or tool is responsible?]
		- **Resources Needed**: [Any required tools or data]
		- **Deadline**: [Timeline for completion]
		</think>
		
		Finalize each action, ensuring it is **ready for execution** with a detailed **step-by-step plan**.
		`,
				formatPreviousSteps(steps),
			)
		}
		return ""
	}
}

func convertGoalsToString(goals []Goal) string {
	var sb strings.Builder
	for _, goal := range goals {
		sb.WriteString(fmt.Sprintf("Name: %s, Description: %s, Weight: %f\n", goal.Name, goal.Description, goal.Weight))
	}
	return sb.String()
}

func formatMap(data map[string]interface{}) string {
	var result string
	for key, value := range data {
		result += fmt.Sprintf("%s: %v\n", key, value)
	}
	return result
}

func formatTools(tools []tools.Tool) string {
	var result string
	for _, tool := range tools {
		result += fmt.Sprintf("- **%s**: %s\n", tool.Name(), tool.Description())
	}
	return result
}

func buildMessagePrompt(state *SystemState, msg *SocialMessage, historicalMsgs []string, stakeholderType token.StakeholderType) string {
	var priorityAccountInfo string
	if stakeholderType == token.StakeholderTypePriority {
		priorityAccountInfo = "IMPORTANT! This is a priority account. The input from this account should be more important and require immediate attention."
	}
	// Create a prompt that explains all the possible types and asks for structured analysis
	return fmt.Sprintf(`
%s
You received this user message from %s. You should analysis the message and return a JSON object with specific fields.
Available Intent Types: question, feedback, complaint, suggestion, greeting, inquiry, request, acknowledge
Available Entity Types: person, product, company, location, datetime, crypto, wallet, contract
Available Emotion Types: positive, negative, neutral

For the message input from the user: "%s"

Historical messages and context from this user: %s

%s

Please analyze the message and provide the following information:

Return a JSON object with these fields:
{
	"intent": "one of the intent types",
	"entity": "one of the entity types",
	"emotion": "one of the emotion types",
	"confidence": confidence score between 0 and 1,
	"should_reply": boolean indicating if a reply is needed,
	"response_msg": "appropriate response message if should_reply is true",
	"should_generate_task": boolean indicating if this requires task creation,
	"should_generate_action": boolean indicating if this requires action generation
}

If you want to generate the reply, you should mainly focus on the message input from the user and only use the historical messages for context.
If you need more information, feel free to ask the user for clarification.
`, getGeneralInfo(state), msg.Platform, strings.Join(historicalMsgs, ";"), priorityAccountInfo, msg.Content)
}
