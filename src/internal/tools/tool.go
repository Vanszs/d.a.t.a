package tools

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

// Tool interface
type Tool interface {
	Initialize(ctx context.Context) error
	Name() string
	Description() string
	AvailableActions() []actions.Action
}
