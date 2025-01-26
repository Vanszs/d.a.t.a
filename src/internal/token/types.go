package token

import (
	"math/big"
	"time"
)

type TokenState struct {
	Supply     *big.Int
	Holders    map[string]*big.Int
	Price      float64
	LastUpdate time.Time
}

type ProposalState string

const (
	ProposalStatePending  ProposalState = "PENDING"
	ProposalStateActive   ProposalState = "ACTIVE"
	ProposalStateApproved ProposalState = "APPROVED"
	ProposalStateRejected ProposalState = "REJECTED"
	ProposalStateExecuted ProposalState = "EXECUTED"
)

type Proposal struct {
	ID          string
	Title       string
	Description string
	State       ProposalState
	Creator     string
	StartTime   time.Time
	EndTime     time.Time
	Votes       map[string]Vote
	Actions     []ProposalAction
}

type Vote struct {
	Voter     string
	Power     *big.Int
	Support   bool
	Timestamp time.Time
}
