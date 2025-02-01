package core

import (
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/characters"
)

func generateTasksPromptFunc(character characters.Character, systemState *SystemState) promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
		switch stepPurpose {
		case PurposeInitial:
			return fmt.Sprintf(`
You are an AI planning assistant for the agent **%s**. Your goal is to generate a diverse set of **high-level tasks** that will guide the AI agent toward fulfilling its mission.

### **Agent Information**
- **Description**: %s
- **Primary Goals**: %s
- **Stakeholder Preferences**: %s

### **Task Generation Guidelines**
1. **Strategic Alignment**: Tasks should contribute directly to the **core objectives** of the agent.
2. **Situational Awareness**: Consider the **current system state**, adapting tasks to evolving conditions.
3. **Stakeholder Relevance**: Ensure tasks **align with preferences** and expectations.
4. **Variety and Coverage**: Generate **different alternatives**, from analytical to operational.

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
- Provide **at least 5-7** distinct, high-level tasks.
- Ensure tasks are **actionable**, **strategically valuable**, and **scalable**.

Now, generate the most relevant and impactful tasks for **%s**.`,
				character.Name,
				character.System,
				convertGoalsToString(systemState.AgentStates.Goals),
				formatMap(systemState.StakeholderPreferences),
				character.Name,
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
			)
		case PurposeRefinement:
			// Purpose Refinement: Improve and polish the tasks based on analysis and feedback.
			return fmt.Sprintf(`
Let's refine the tasks based on the analysis.

### **Refinement Questions:**
1. **Clarity and Focus**: Are the tasks clearly defined with a specific, actionable goal?
2. **Prioritization**: Which tasks should be prioritized based on their potential impact and feasibility?
3. **Efficiency**: Can the tasks be broken down into smaller, more manageable steps?
4. **Stakeholder Consideration**: Are there any further adjustments needed to meet stakeholder preferences?

### **Refined Task Format**
For each task, provide a detailed refinement:

**<think>**
- **Task Name**: [Refined task title]
- **Objective**: [What is the clear and actionable goal?]
- **Execution Plan**: [Break down the task into actionable steps]
- **Priority**: [High / Medium / Low]
- **Stakeholder Alignment**: [How does this meet stakeholder needs?]
**</think>

Refine the tasks, making them **clearer, actionable, and aligned with the overall goals**.
`,
				formatPreviousSteps(steps),
			)
		case PurposeConcrete:
			// Purpose Concrete: Finalize the tasks into fully executable plans with precise actions.
			return fmt.Sprintf(`
The tasks are now ready for execution.

### **Finalization Steps:**
1. **Actionability**: Ensure each task can be executed with a clear step-by-step plan.
2. **Responsibility Assignment**: Assign tasks to specific agents or systems responsible for execution.
3. **Resources**: Ensure all necessary resources (e.g., data, tools) are available to carry out the tasks.
4. **Timeline**: Define clear **deadlines** or **milestones** for each task.

### **Execution Plan Format**
For each task, create a final **execution plan**:

**<think>**
- **Task Name**: [Task name for execution]
- **Execution Steps**: [List clear steps for completing the task]
- **Assigned To**: [Who will execute the task?]
- **Resources Required**: [Data, tools, or dependencies needed]
- **Timeline**: [Specify milestones or deadlines]
**</think>

Finalize the tasks into **actionable execution plans**, ensuring all dependencies and resources are accounted for.
`,
				formatPreviousSteps(steps),
			)
		}
		return ""
	}
}

func generateActionsPromptFunc() promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
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

func buildMessagePrompt(msgContext map[string]interface{}) string {
	return ""
}
