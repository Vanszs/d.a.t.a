package token

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Governance struct {
	proposals map[string]*Proposal
	state     *TokenState
	log       *zap.SugaredLogger
}

func (g *Governance) CreateProposal(ctx context.Context, creator string, proposal Proposal) error {
	// Validate creator has enough tokens
	balance := g.state.Holders[creator]
	if balance == nil || balance.Cmp(g.minProposalThreshold) < 0 {
		return ErrInsufficientBalance
	}

	g.log.Infow("Creating new proposal",
		"creator", creator,
		"title", proposal.Title)

	proposal.State = ProposalStatePending
	proposal.Creator = creator
	proposal.StartTime = time.Now()
	proposal.EndTime = proposal.StartTime.Add(g.votingPeriod)

	g.proposals[proposal.ID] = &proposal
	return nil
}

func (g *Governance) CastVote(ctx context.Context, proposalID string, voter string, support bool) error {
	proposal, exists := g.proposals[proposalID]
	if !exists {
		return ErrProposalNotFound
	}

	if proposal.State != ProposalStateActive {
		return ErrInvalidProposalState
	}

	balance := g.state.Holders[voter]
	if balance == nil || balance.Sign() <= 0 {
		return ErrInsufficientBalance
	}

	g.log.Infow("Vote cast",
		"proposal", proposalID,
		"voter", voter,
		"support", support,
		"power", balance)

	proposal.Votes[voter] = Vote{
		Voter:     voter,
		Power:     balance,
		Support:   support,
		Timestamp: time.Now(),
	}

	return nil
}
