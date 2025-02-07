package tools

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type WalletTool struct {
}

func (t *WalletTool) Initialize(ctx context.Context) error {
	return nil
}

func (t *WalletTool) Name() string {
	return "wallet tool"
}

func (t *WalletTool) Description() string {
	return `The wallet tool allows you to manage your wallet. You can control your own wallet to send tokens, receive tokens, etc.
	It also allows you to create your own token, manage your own token, list your token on decentralized exchanges, etc.`
}

func (t *WalletTool) AvailableActions() []core.Action {
	return nil
}
