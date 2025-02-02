package actions

type Action interface {
	Name() string
	Description() string
	Execute() error
}

type Manager interface {
	Register(action Action) error
	GetAvailableActions() []Action
}
