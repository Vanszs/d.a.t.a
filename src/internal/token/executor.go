package token

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Executor struct {
	client       *ethclient.Client
	tokenAddress common.Address
	log          *zap.SugaredLogger
}

func (e *Executor) ExecuteProposal(ctx context.Context, proposal *Proposal) error {
	if proposal.State != ProposalStateApproved {
		return ErrInvalidProposalState
	}

	e.log.Infow("Executing proposal",
		"id", proposal.ID,
		"actions", len(proposal.Actions))

	for _, action := range proposal.Actions {
		if err := e.executeAction(ctx, action); err != nil {
			return err
		}
	}

	proposal.State = ProposalStateExecuted
	return nil
}

func (e *Executor) executeAction(ctx context.Context, action ProposalAction) error {
	switch action.Type {
	case ActionTypeMint:
		return e.executeMint(ctx, action)
	case ActionTypeBurn:
		return e.executeBurn(ctx, action)
	case ActionTypeTransfer:
		return e.executeTransfer(ctx, action)
	default:
		return ErrUnsupportedAction
	}
}
