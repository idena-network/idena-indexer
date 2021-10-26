package mempool

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"time"
)

const retryInterval = time.Second * 10

type VoteCountingTracker interface {
	SubmitVoteCountingStepResultEvent(e *stats.VoteCountingStepResultEvent)
	SubmitVoteCountingResultEvent(e *stats.VoteCountingResultEvent)
	SubmitProofProposalEvent(e *stats.ProofProposalEvent)
	SubmitBlockProposalEvent(e *stats.BlockProposalEvent)
}

func NewVoteCountingTracker(dbAccessor db.Accessor, logger log.Logger) VoteCountingTracker {
	res := &voteCountingTracker{
		dbAccessor:        dbAccessor,
		countingQueue:     make(chan *voteCountingResultEventWrapper, 100),
		countingStepQueue: make(chan *voteCountingStepResultEventWrapper, 1000),
		proofQueue:        make(chan *proofProposalEventWrapper, 1000),
		blockQueue:        make(chan *stats.BlockProposalEvent, 1000),
		logger:            logger,
	}
	go res.track()
	return res
}

type voteCountingTracker struct {
	dbAccessor        db.Accessor
	countingQueue     chan *voteCountingResultEventWrapper
	countingStepQueue chan *voteCountingStepResultEventWrapper
	proofQueue        chan *proofProposalEventWrapper
	blockQueue        chan *stats.BlockProposalEvent
	logger            log.Logger
}

type voteCountingStepResultEventWrapper struct {
	timestamp time.Time
	event     *stats.VoteCountingStepResultEvent
}

type voteCountingResultEventWrapper struct {
	timestamp time.Time
	event     *stats.VoteCountingResultEvent
}

type proofProposalEventWrapper struct {
	timestamp time.Time
	event     *stats.ProofProposalEvent
}

func (t *voteCountingTracker) SubmitVoteCountingStepResultEvent(e *stats.VoteCountingStepResultEvent) {
	timestamp := time.Now()
	eventCopy := &stats.VoteCountingStepResultEvent{}
	eventCopy.Round = e.Round
	eventCopy.Step = e.Step
	eventCopy.CheckedRoundVotes = e.CheckedRoundVotes
	eventCopy.NecessaryVotesCount = e.NecessaryVotesCount
	if len(e.VotesByBlock) > 0 {
		eventCopy.VotesByBlock = make(map[common.Hash]map[common.Address]*types.Vote, len(e.VotesByBlock))
		for hash, votes := range e.VotesByBlock {
			if len(votes) == 0 {
				continue
			}
			votesByAddr := make(map[common.Address]*types.Vote, len(votes))
			for addr, vote := range votes {
				votesByAddr[addr] = vote
			}
			eventCopy.VotesByBlock[hash] = votesByAddr
		}
	}
	eventWrapper := &voteCountingStepResultEventWrapper{
		timestamp: timestamp,
		event:     eventCopy,
	}
	select {
	case t.countingStepQueue <- eventWrapper:
	default:
		t.logger.Warn("Vote counting step result skipped")
	}
}

func (t *voteCountingTracker) SubmitVoteCountingResultEvent(e *stats.VoteCountingResultEvent) {
	timestamp := time.Now()
	eventCopy := &stats.VoteCountingResultEvent{}
	eventCopy.Round = e.Round
	eventCopy.Step = e.Step
	eventCopy.Hash = e.Hash
	eventCopy.Err = e.Err
	if e.Validators != nil {
		var originalCopy, addressesCopy mapset.Set
		if e.Validators.Original != nil {
			originalCopy = e.Validators.Original.Clone()
		}
		if e.Validators.Addresses != nil {
			addressesCopy = e.Validators.Addresses.Clone()
		}
		eventCopy.Validators = &validators.StepValidators{
			Original:  originalCopy,
			Addresses: addressesCopy,
			Size:      e.Validators.Size,
		}
	}
	if e.Cert != nil {
		eventCopy.Cert = &types.FullBlockCert{}
		if len(e.Cert.Votes) > 0 {
			eventCopy.Cert.Votes = make([]*types.Vote, len(e.Cert.Votes))
			copy(eventCopy.Cert.Votes, e.Cert.Votes)
		}
	}
	eventWrapper := &voteCountingResultEventWrapper{
		timestamp: timestamp,
		event:     eventCopy,
	}
	select {
	case t.countingQueue <- eventWrapper:
	default:
		t.logger.Warn("Vote counting result skipped")
	}
}

func (t *voteCountingTracker) SubmitProofProposalEvent(e *stats.ProofProposalEvent) {
	timestamp := time.Now()
	eventWrapper := &proofProposalEventWrapper{
		timestamp: timestamp,
		event:     e,
	}
	select {
	case t.proofQueue <- eventWrapper:
	default:
		t.logger.Warn("Vote counting result skipped")
	}
}

func (t *voteCountingTracker) SubmitBlockProposalEvent(e *stats.BlockProposalEvent) {
	select {
	case t.blockQueue <- e:
	default:
		t.logger.Warn("Block proposal skipped")
	}
}

func (t *voteCountingTracker) track() {
	deleteOldDataTicker := time.NewTicker(time.Minute)
	for {
		select {
		case countingStep := <-t.countingStepQueue:
			t.handleCountingStep(countingStep)
		case counting := <-t.countingQueue:
			t.handleCounting(counting)
		case proofProposal := <-t.proofQueue:
			t.handleProofProposal(proofProposal)
		case blockProposal := <-t.blockQueue:
			t.handleBlockProposal(blockProposal)
		case <-deleteOldDataTicker.C:
			t.deleteOldData()
		}
	}
}

func (t *voteCountingTracker) handleCountingStep(countingStep *voteCountingStepResultEventWrapper) {
	err := t.dbAccessor.SaveVoteCountingStepResult(convertVoteCountingStepResultEventWrapper(countingStep))
	for err != nil {
		t.logger.Error("Failed to save vote counting step details", "err", err)
		time.Sleep(retryInterval)
		err = t.dbAccessor.SaveVoteCountingStepResult(convertVoteCountingStepResultEventWrapper(countingStep))
	}
	t.logger.Trace("Vote counting step details saved")
}

func (t *voteCountingTracker) handleCounting(counting *voteCountingResultEventWrapper) {
	err := t.dbAccessor.SaveVoteCountingResult(convertVoteCountingResultEvent(counting))
	for err != nil {
		t.logger.Error("Failed to save vote counting details", "err", err)
		time.Sleep(retryInterval)
		err = t.dbAccessor.SaveVoteCountingResult(convertVoteCountingResultEvent(counting))
	}
	t.logger.Trace("Vote counting details saved")
}

func (t *voteCountingTracker) handleProofProposal(proofProposal *proofProposalEventWrapper) {
	value, err := convertProofProposalEvent(proofProposal)
	if err != nil {
		t.logger.Warn("Failed to convert proof proposal event", "err", err)
		return
	}
	err = t.dbAccessor.SaveProofProposal(value)
	for err != nil {
		t.logger.Error("Failed to save proof proposal", "err", err)
		time.Sleep(retryInterval)
		err = t.dbAccessor.SaveProofProposal(value)
	}
	t.logger.Trace("Proof proposal saved")
}

func (t *voteCountingTracker) handleBlockProposal(blockProposal *stats.BlockProposalEvent) {
	err := t.dbAccessor.SaveBlockProposal(convertBlockProposalEvent(blockProposal))
	for err != nil {
		t.logger.Error("Failed to save block proposal", "err", err)
		time.Sleep(retryInterval)
		err = t.dbAccessor.SaveBlockProposal(convertBlockProposalEvent(blockProposal))
	}
	t.logger.Trace("Block proposal saved")
}

func convertVoteCountingStepResultEventWrapper(value *voteCountingStepResultEventWrapper) *db.VoteCountingStepResult {
	res := &db.VoteCountingStepResult{
		Timestamp:           value.timestamp.UTC().UnixNano(),
		Round:               value.event.Round,
		Step:                value.event.Step,
		NecessaryVotesCount: value.event.NecessaryVotesCount,
		CheckedRoundVotes:   value.event.CheckedRoundVotes,
	}
	for _, votesByAddress := range value.event.VotesByBlock {
		for _, vote := range votesByAddress {
			res.Votes = append(res.Votes, convertVote(vote))
		}
	}
	return res
}

func convertVote(value *types.Vote) *db.Vote {
	return &db.Vote{
		Voter:       value.VoterAddr(),
		ParentHash:  value.Header.ParentHash,
		VotedHash:   value.Header.VotedHash,
		TurnOffline: value.Header.TurnOffline,
		Upgrade:     value.Header.Upgrade,
	}
}

func convertVoteCountingResultEvent(value *voteCountingResultEventWrapper) *db.VoteCountingResult {
	res := &db.VoteCountingResult{
		Timestamp: value.timestamp.UTC().UnixNano(),
		Round:     value.event.Round,
		Step:      value.event.Step,
		Hash:      value.event.Hash,
	}
	if value.event.Err != nil {
		err := value.event.Err.Error()
		res.Err = &err
	}
	if value.event.Validators != nil {
		stepValidators := &db.StepValidators{
			Size: value.event.Validators.Size,
		}
		if value.event.Validators.Original != nil {
			stepValidators.Original = make([]common.Address, 0, value.event.Validators.Original.Cardinality())
			value.event.Validators.Original.Each(func(address interface{}) bool {
				stepValidators.Original = append(stepValidators.Original, address.(common.Address))
				return false
			})
		}
		if value.event.Validators.Addresses != nil {
			stepValidators.Addresses = make([]common.Address, 0, value.event.Validators.Addresses.Cardinality())
			value.event.Validators.Addresses.Each(func(address interface{}) bool {
				stepValidators.Addresses = append(stepValidators.Addresses, address.(common.Address))
				return false
			})
		}
		res.Validators = stepValidators
	}
	if value.event.Cert != nil {
		cert := &db.FullBlockCert{
			Votes: make([]*db.Vote, 0, len(value.event.Cert.Votes)),
		}
		for _, vote := range value.event.Cert.Votes {
			cert.Votes = append(cert.Votes, convertVote(vote))
		}
		res.Cert = cert
	}
	return res
}

func convertProofProposalEvent(value *proofProposalEventWrapper) (*db.ProofProposal, error) {
	res := &db.ProofProposal{
		Timestamp: value.timestamp.UTC().UnixNano(),
		Round:     value.event.Round,
		Hash:      value.event.Hash,
		Modifier:  value.event.Modifier,
	}
	proposer, err := crypto.PubKeyBytesToAddress(value.event.ProposerPubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert pub key to address")
	}
	res.Proposer = proposer
	q := common.HashToFloat(value.event.Hash, value.event.Modifier)
	res.VrfScore, _ = q.Float64()
	return res, nil
}

func convertBlockProposalEvent(value *stats.BlockProposalEvent) *db.BlockProposal {
	res := &db.BlockProposal{
		ReceivingTime: value.ReceivingTime.UTC().UnixNano(),
		Height:        value.Proposal.Height(),
		Proposer:      value.Proposal.Header.Coinbase(),
		Hash:          value.Proposal.Hash(),
	}
	return res
}

func (t *voteCountingTracker) deleteOldData() {
	if err := t.dbAccessor.DeleteVoteCountingOldData(); err != nil {
		t.logger.Error("Failed to delete vote counting old data", "err", err)
	}
	t.logger.Trace("Vote counting old data deleted")
}
