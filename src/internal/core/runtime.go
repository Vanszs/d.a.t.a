package core

import (
	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/goal"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
)

type agentRuntime struct {
	agent     *Agent
	userInput data.UserInput
}

func (r *agentRuntime) GetMemory() memory.Manager           { return r.agent.memory }
func (r *agentRuntime) GetData() *data.Manager              { return r.agent.dataManager }
func (r *agentRuntime) GetGoals() *goal.Manager             { return r.agent.goalManager }
func (r *agentRuntime) GetCharacter() *characters.Character { return r.agent.character }
