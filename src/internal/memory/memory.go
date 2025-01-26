package memory

import "context"

type Manager struct {
	workingMemory  WorkingMemory
	longTermMemory LongTermMemory
}

func NewManager(workingMemory WorkingMemory, longTermMemory LongTermMemory) *Manager {
	return &Manager{
		workingMemory:  workingMemory,
		longTermMemory: longTermMemory,
	}
}

func (m *Manager) Add(ctx context.Context, entry Entry) error {
	return m.workingMemory.Add(ctx, entry)
}

func (m *Manager) StoreFailure(ctx context.Context, entry Entry) error {
	return m.longTermMemory.Store(ctx, entry)
}
