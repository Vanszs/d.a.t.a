package tools

import (
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type Manager struct {
	tools map[string]core.Tool
}

// TODO: keep one of tool or plugin
func NewManager() *Manager {
	return &Manager{
		tools: make(map[string]core.Tool),
	}
}

// Register a plugin
func (m *Manager) Register(tool core.Tool) error {
	if _, exists := m.tools[tool.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", tool.Name())
	}

	m.tools[tool.Name()] = tool
	return nil
}

func (m *Manager) AvailableTools() []core.Tool {
	var tools []core.Tool
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}
