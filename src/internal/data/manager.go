package data

import (
	"context"
	"math/big"

	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

type UserInput struct {
	Content string
}

type StakeholderInfo struct {
	TokenBalance map[string]TokenInfo
}

type TokenInfo struct {
	Network      string
	ContractAddr string
	Balance      *big.Int
}

type Manager interface {
	// Register(ctx context.Context, source DataSource) error
	GetTokenBalance(
		ctx context.Context,
		id string,
		platform string,
		network string,
		ticker string,
	) (*TokenInfo, error)
}

type DataSource interface {
	Name() string
	Initialize(ctx context.Context) error
	Fetch(ctx context.Context, dataType string, input interface{}) (*DataOutput, error)
}

type managerImpl struct {
	carvClient *carv.Client
	// carvDataSource
	llmClient llm.Client
}

type DataOutput struct {
	Blob []byte
}

func NewManager(llmClient llm.Client, carvClient *carv.Client) *managerImpl {
	return &managerImpl{
		carvClient: carvClient,
		llmClient:  llmClient,
	}
}

func (m *managerImpl) GetStakeholderInfo(
	ctx context.Context,
	id string,
	platform string,
	network string,
	ticker string,
) (*StakeholderInfo, error) {
	return nil, nil
}
