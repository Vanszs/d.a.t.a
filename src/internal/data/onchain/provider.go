package onchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Provider struct {
	client *ethclient.Client
	log    *zap.SugaredLogger
}

func (p *Provider) GetTokenBalance(ctx context.Context, token, address string) (*big.Int, error) {
	p.log.Debugw("Fetching token balance",
		"token", token,
		"address", address)

	// Implementation for fetching token balance
	return nil, nil
}

func (p *Provider) GetTokenMetrics(ctx context.Context, token string) (TokenMetrics, error) {
	p.log.Debugw("Fetching token metrics", "token", token)

	// Implementation for fetching token metrics
	return TokenMetrics{}, nil
}
