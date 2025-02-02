package tools

import (
	"fmt"
)

type Manager struct {
	tools map[string]Tool
}

// TODO: keep one of tool or plugin
func NewManager() *Manager {
	return &Manager{
		tools: make(map[string]Tool),
	}
}

// Register a plugin
func (m *Manager) Register(tool Tool) error {
	if _, exists := m.tools[tool.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", tool.Name())
	}

	m.tools[tool.Name()] = tool
	return nil
}

func (m *Manager) AvailableTools() []Tool {
	var tools []Tool
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}
