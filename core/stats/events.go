package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/validators"
	"time"
)

const (
	RemovedMemPoolTxEventID       = eventbus.EventID("removed-mem-pool-tx")
	VoteCountingStepResultEventID = eventbus.EventID("vote-counting-result")
	VoteCountingResultEventID     = eventbus.EventID("vote-counting-step-result")
	ProofProposalEventID          = eventbus.EventID("proof-proposal")
	BlockProposalEventID          = eventbus.EventID("block-proposal")
)

type RemovedMemPoolTxEvent struct {
	Tx *types.Transaction
}

func (e *RemovedMemPoolTxEvent) EventID() eventbus.EventID {
	return RemovedMemPoolTxEventID
}

type VoteCountingStepResultEvent struct {
	Round               uint64
	Step                uint8
	VotesByBlock        map[common.Hash]map[common.Address]*types.Vote
	NecessaryVotesCount int
	CheckedRoundVotes   int
}

func (e *VoteCountingStepResultEvent) EventID() eventbus.EventID {
	return VoteCountingStepResultEventID
}

type VoteCountingResultEvent struct {
	Round      uint64
	Step       uint8
	Validators *validators.StepValidators
	Hash       common.Hash
	Cert       *types.FullBlockCert
	Err        error
}

func (e *VoteCountingResultEvent) EventID() eventbus.EventID {
	return VoteCountingResultEventID
}

type ProofProposalEvent struct {
	Round          uint64
	Hash           common.Hash
	ProposerPubKey []byte
	Modifier       int64
}

func (e *ProofProposalEvent) EventID() eventbus.EventID {
	return ProofProposalEventID
}

type BlockProposalEvent struct {
	Proposal      *types.BlockProposal
	ReceivingTime time.Time
}

func (e *BlockProposalEvent) EventID() eventbus.EventID {
	return BlockProposalEventID
}
