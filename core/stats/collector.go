package stats

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/collector"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/vm/helpers"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

var (
	maxOracleVotingHash *big.Float
)

func init() {
	var max [32]byte
	for i := range max {
		max[i] = 0xFF
	}
	i := new(big.Int)
	i.SetBytes(max[:])
	maxOracleVotingHash = new(big.Float).SetInt(i)
}

type statsCollector struct {
	stats        *Stats
	statsEnabled bool
	pending      *pending
	bus          eventbus.Bus
}

type pending struct {
	epochRewardsByTypeAndAddr     map[RewardType]map[common.Address]*RewardStats
	balanceUpdates                []*db.BalanceUpdate
	balanceUpdatesByReasonAndAddr map[db.BalanceUpdateReason]map[common.Address]*db.BalanceUpdate
	identityStates                []state.IdentityState
	tx                            *pendingTx
	identitiesByAddr              map[common.Address]*identityInfo
	finalCommitteeRewardsByAddr   map[common.Address]*MiningReward
	linkedBalanceUpdatesByPrev    map[*db.BalanceUpdate]*db.BalanceUpdate
	linkedBalanceUpdatesByNext    map[*db.BalanceUpdate]*db.BalanceUpdate
}

type identityInfo struct {
	pubKey []byte
	state  state.IdentityState
}

type pendingTx struct {
	tx                                      *types.Transaction
	contractBalanceUpdatesByAddr            map[common.Address]*db.BalanceUpdate
	contractBurntCoins                      []*pendingBurntCoins
	oracleVotingContractDeploy              *db.OracleVotingContract
	oracleVotingContractCallStart           *db.OracleVotingContractCallStart
	oracleVotingCommitteeStartCtx           *oracleVotingCommitteeStartCtx
	oracleVotingContractCallVoteProof       *db.OracleVotingContractCallVoteProof
	oracleVotingContractCallVote            *db.OracleVotingContractCallVote
	oracleVotingContractCallFinish          *db.OracleVotingContractCallFinish
	oracleVotingContractCallProlongation    *db.OracleVotingContractCallProlongation
	oracleVotingContractCallAddStake        *db.OracleVotingContractCallAddStake
	oracleVotingContractTermination         *db.OracleVotingContractTermination
	oracleLockContract                      *db.OracleLockContract
	oracleLockContractCallCheckOracleVoting *db.OracleLockContractCallCheckOracleVoting
	oracleLockContractCallPush              *db.OracleLockContractCallPush
	oracleLockContractTermination           *db.OracleLockContractTermination
	refundableOracleLockContract            *db.RefundableOracleLockContract
	refundableOracleLockContractCallDeposit *db.RefundableOracleLockContractCallDeposit
	refundableOracleLockContractCallPush    *db.RefundableOracleLockContractCallPush
	refundableOracleLockContractCallRefund  *db.RefundableOracleLockContractCallRefund
	refundableOracleLockContractTermination *db.RefundableOracleLockContractTermination
	multisigContract                        *db.MultisigContract
	multisigContractCallAdd                 *db.MultisigContractCallAdd
	multisigContractCallSend                *db.MultisigContractCallSend
	multisigContractCallPush                *db.MultisigContractCallPush
	multisigContractTermination             *db.MultisigContractTermination
	timeLockContract                        *db.TimeLockContract
	timeLockContractCallTransfer            *db.TimeLockContractCallTransfer
	timeLockContractTermination             *db.TimeLockContractTermination
	inviteTxHash                            *common.Hash
}

type pendingBurntCoins struct {
	address common.Address
	amount  *big.Int
	reason  db.BalanceUpdateReason
}

type oracleVotingCommitteeStartCtx struct {
	committeeSize uint64
	networkSize   int
}

func NewStatsCollector(bus eventbus.Bus) collector.StatsCollector {
	return &statsCollector{
		bus: bus,
	}
}

func (c *statsCollector) RemoveMemPoolTx(tx *types.Transaction) {
	if tx == nil {
		return
	}
	c.bus.Publish(&RemovedMemPoolTxEvent{
		Tx: tx,
	})
}

func (c *statsCollector) SubmitVoteCountingStepResult(round uint64, step uint8, votesByBlock map[common.Hash]map[common.Address]*types.Vote, necessaryVotesCount, checkedRoundVotes int) {
	c.bus.Publish(&VoteCountingStepResultEvent{
		Round:               round,
		Step:                step,
		VotesByBlock:        votesByBlock,
		NecessaryVotesCount: necessaryVotesCount,
		CheckedRoundVotes:   checkedRoundVotes,
	})
}

func (c *statsCollector) SubmitVoteCountingResult(round uint64, step uint8, validators *validators.StepValidators, hash common.Hash, cert *types.FullBlockCert, err error) {
	c.bus.Publish(&VoteCountingResultEvent{
		Round:      round,
		Step:       step,
		Validators: validators,
		Hash:       hash,
		Cert:       cert,
		Err:        err,
	})
}

func (c *statsCollector) SubmitProofProposal(round uint64, hash common.Hash, proposerPubKey []byte, modifier int) {
	c.bus.Publish(&ProofProposalEvent{
		Round:          round,
		Hash:           hash,
		ProposerPubKey: proposerPubKey,
		Modifier:       int64(modifier),
	})
}

func (c *statsCollector) SubmitBlockProposal(proposal *types.BlockProposal, receivingTime time.Time) {
	c.bus.Publish(&BlockProposalEvent{
		Proposal:      proposal,
		ReceivingTime: receivingTime,
	})
}

func (c *statsCollector) EnableCollecting() {
	c.stats = &Stats{}
	c.pending = &pending{}
}

func (c *statsCollector) initRewardStats() {
	if c.stats.RewardsStats != nil {
		return
	}
	c.stats.RewardsStats = &RewardsStats{}
}

func (c *statsCollector) SetValidation(validation *statsTypes.ValidationStats) {
	c.stats.ValidationStats = validation
}

func (c *statsCollector) SetMinScoreForInvite(score float32) {
	c.stats.MinScoreForInvite = &score
}

func (c *statsCollector) SetValidationResults(validationResults map[common.ShardId]*types.ValidationResults) {
	c.initRewardStats()
	c.stats.RewardsStats.ValidationResults = validationResults
}

func (c *statsCollector) SetTotalReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Total = amount
}

func (c *statsCollector) SetTotalValidationReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Validation = amount
	c.stats.RewardsStats.ValidationShare = share
}

func (c *statsCollector) SetTotalStakingReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Staking = amount
	c.stats.RewardsStats.StakingShare = share
}

func (c *statsCollector) SetTotalCandidateReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Candidate = amount
	c.stats.RewardsStats.CandidateShare = share
}

func (c *statsCollector) SetTotalFlipsReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Flips = amount
	c.stats.RewardsStats.FlipsShare = share
}

func (c *statsCollector) SetTotalReportsReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Reports = amount
	c.stats.RewardsStats.ReportsShare = share
}

func (c *statsCollector) SetTotalInvitationsReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Invitations = amount
	c.stats.RewardsStats.InvitationsShare = share
}

func (c *statsCollector) SetTotalFoundationPayouts(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.FoundationPayouts = amount
}

func (c *statsCollector) SetTotalZeroWalletFund(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.ZeroWalletFund = amount
}

func (c *statsCollector) AddValidationReward(balanceDest, stakeDest common.Address, age uint16, balance, stake *big.Int) {
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, Validation)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, Validation)
	}

	c.initRewardStats()
	if c.stats.RewardsStats.AgesByAddress == nil {
		c.stats.RewardsStats.AgesByAddress = make(map[string]uint16)
	}
	c.stats.RewardsStats.AgesByAddress[conversion.ConvertAddress(baseRewardRecipient)] = age + 1

	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func (c *statsCollector) AddCandidateReward(balanceDest, stakeDest common.Address, balance, stake *big.Int) {
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, Candidate)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, Candidate)
	}
	c.initRewardStats()
	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func (c *statsCollector) AddStakingReward(balanceDest, stakeDest common.Address, stakedAmount *big.Int, balance, stake *big.Int) {
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, Staking)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, Staking)
	}

	c.initRewardStats()
	if stakedAmount != nil {
		if c.stats.RewardsStats.StakedAmountsByAddress == nil {
			c.stats.RewardsStats.StakedAmountsByAddress = make(map[string]*big.Int)
		}
		c.stats.RewardsStats.StakedAmountsByAddress[conversion.ConvertAddress(baseRewardRecipient)] = new(big.Int).Set(stakedAmount)
	}

	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func (c *statsCollector) AddFlipsReward(balanceDest, stakeDest common.Address, balance, stake *big.Int, flipsToReward []*types.FlipToReward) {
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, Flips)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, Flips)
	}
	c.addRewardedFlips(flipsToReward)

	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func (c *statsCollector) addRewardedFlips(flipsToReward []*types.FlipToReward) {
	if len(flipsToReward) == 0 {
		return
	}
	c.initRewardStats()
	for _, rewardedFlip := range flipsToReward {
		flipCid, _ := cid.Parse(rewardedFlip.Cid)
		c.stats.RewardsStats.RewardedFlipCids = append(c.stats.RewardsStats.RewardedFlipCids, flipCid.String())
	}
}

func (c *statsCollector) AddReportedFlipsReward(balanceDest, stakeDest common.Address, shardId common.ShardId, flipIdx int, balance *big.Int, stake *big.Int) {
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, ReportedFlips)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, ReportedFlips)
	}
	c.addReportedFlipReward(baseRewardRecipient, shardId, flipIdx, balance, stake)

	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func (c *statsCollector) addReportedFlipReward(addr common.Address, shardId common.ShardId, flipIdx int, balance *big.Int, stake *big.Int) {
	cidBytes, ok := c.getFlipCid(shardId, flipIdx)
	if !ok {
		log.Warn(fmt.Sprintf("Cid for flip %d not found", flipIdx))
		return
	}
	flipCid, _ := cid.Parse(cidBytes)
	c.initRewardStats()
	c.stats.RewardsStats.ReportedFlipRewards = append(c.stats.RewardsStats.ReportedFlipRewards, &db.ReportedFlipReward{
		Address: conversion.ConvertAddress(addr),
		Balance: blockchain.ConvertToFloat(balance),
		Stake:   blockchain.ConvertToFloat(stake),
		Cid:     flipCid.String(),
	})
}

func (c *statsCollector) getFlipCid(shardId common.ShardId, flipIdx int) ([]byte, bool) {
	if flipIdx < 0 || c.stats.ValidationStats.Shards == nil {
		return nil, false
	}
	validationShardStats, ok := c.stats.ValidationStats.Shards[shardId]
	if !ok {
		return nil, false
	}
	if flipIdx >= len(validationShardStats.FlipCids) {
		return nil, false
	}
	return validationShardStats.FlipCids[flipIdx], true
}

func (c *statsCollector) AddInvitationsReward(balanceDest, stakeDest common.Address, balance *big.Int, stake *big.Int, age uint16,
	txHash *common.Hash, epochHeight uint32, isSavedInviteWinner bool) {
	rewardType, err := determineInvitationsRewardType(age, isSavedInviteWinner)
	if err != nil {
		log.Warn(err.Error())
		return
	}
	baseRewardRecipient := stakeDest
	c.addReward(baseRewardRecipient, balance, stake, rewardType)
	if balanceDest != stakeDest {
		c.addDelegateeReward(stakeDest, balanceDest, balance, nil, rewardType)
	}
	c.addRewardedInvite(baseRewardRecipient, txHash, rewardType, epochHeight)

	c.addAddrTotalReward(baseRewardRecipient, balance, stake)
}

func determineInvitationsRewardType(age uint16, isSavedInviteWinner bool) (RewardType, error) {
	switch age {
	case 0:
		if isSavedInviteWinner {
			return SavedInviteWin, nil
		}
		return SavedInvite, nil
	case 1:
		return Invitations, nil
	case 2:
		return Invitations2, nil
	case 3:
		return Invitations3, nil
	default:
		return 0, errors.Errorf("no invitations reward type for age: %v, isSavedInviteWinner: %v", age, isSavedInviteWinner)
	}
}

func (c *statsCollector) addRewardedInvite(addr common.Address, txHash *common.Hash, rewardType RewardType, epochHeight uint32) {
	if rewardType == SavedInviteWin || rewardType == SavedInvite {
		c.initRewardStats()
		if c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType == nil {
			c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType = make(map[common.Address]map[RewardType]uint8)
		}
		if _, ok := c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr]; !ok {
			c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr] = make(map[RewardType]uint8)
		}
		c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr][rewardType]++
		return
	}
	if txHash == nil {
		log.Warn(fmt.Sprintf("wrong value txHash=nil for rewardType=%v", rewardType))
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.RewardedInvites = append(c.stats.RewardsStats.RewardedInvites, &db.RewardedInvite{
		TxHash:      conversion.ConvertHash(*txHash),
		Type:        byte(rewardType),
		EpochHeight: epochHeight,
	})
}

func (c *statsCollector) AddFoundationPayout(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, FoundationPayouts)
}

func (c *statsCollector) AddZeroWalletFund(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, ZeroWalletFund)
}

func (c *statsCollector) addReward(addr common.Address, balance *big.Int, stake *big.Int, rewardType RewardType) {
	if (balance == nil || balance.Sign() == 0) && (stake == nil || stake.Sign() == 0) {
		return
	}
	c.initRewardStats()
	rewardsStats := &RewardStats{
		Address: addr,
		Balance: balance,
		Stake:   stake,
		Type:    rewardType,
	}
	if c.increaseEpochRewardIfExists(rewardsStats) {
		return
	}

	c.stats.RewardsStats.Rewards = append(c.stats.RewardsStats.Rewards, rewardsStats)
}

func (c *statsCollector) initTotalRewardsByAddr() {
	c.initRewardStats()
	if c.stats.RewardsStats.TotalRewardsByAddr != nil {
		return
	}
	c.stats.RewardsStats.TotalRewardsByAddr = make(map[common.Address]*big.Int)
}

func (c *statsCollector) addAddrTotalReward(addr common.Address, balance, stake *big.Int) {
	if (balance == nil || balance.Sign() == 0) && (stake == nil || stake.Sign() == 0) {
		return
	}
	if balance == nil {
		balance = new(big.Int)
	}
	if stake == nil {
		stake = new(big.Int)
	}
	c.initTotalRewardsByAddr()
	reward := new(big.Int).Add(balance, stake)
	if amount, ok := c.stats.RewardsStats.TotalRewardsByAddr[addr]; ok {
		c.stats.RewardsStats.TotalRewardsByAddr[addr] = new(big.Int).Add(amount, reward)
	} else {
		c.stats.RewardsStats.TotalRewardsByAddr[addr] = reward
	}
}

func (c *statsCollector) increaseEpochRewardIfExists(rewardsStats *RewardStats) bool {
	if rewardsStats.Type != Invitations && rewardsStats.Type != Invitations2 && rewardsStats.Type != Invitations3 &&
		rewardsStats.Type != SavedInvite && rewardsStats.Type != SavedInviteWin && rewardsStats.Type != ReportedFlips &&
		rewardsStats.Type != Flips && rewardsStats.Type != Validation {
		return false
	}
	if c.pending.epochRewardsByTypeAndAddr == nil {
		c.pending.epochRewardsByTypeAndAddr = make(map[RewardType]map[common.Address]*RewardStats)
	}
	rewardsByAddr, ok := c.pending.epochRewardsByTypeAndAddr[rewardsStats.Type]
	if ok {
		if reward, ok := rewardsByAddr[rewardsStats.Address]; ok {
			reward.Balance.Add(reward.Balance, rewardsStats.Balance)
			reward.Stake.Add(reward.Stake, rewardsStats.Stake)
			return true
		}
	} else {
		rewardsByAddr = make(map[common.Address]*RewardStats)
	}
	rewardsByAddr[rewardsStats.Address] = rewardsStats
	c.pending.epochRewardsByTypeAndAddr[rewardsStats.Type] = rewardsByAddr
	return false
}

func (c *statsCollector) addDelegateeReward(delegator, delegatee common.Address, balance, stake *big.Int, rewardType RewardType) {
	if (balance == nil || balance.Sign() == 0) && (stake == nil || stake.Sign() == 0) {
		return
	}
	if balance == nil {
		balance = new(big.Int)
	}
	if stake == nil {
		stake = new(big.Int)
	}
	c.initRewardStats()
	if c.stats.RewardsStats.DelegateesEpochRewards == nil {
		c.stats.RewardsStats.DelegateesEpochRewards = make(map[common.Address]*DelegateeEpochRewards)
	}
	delegateeEpochRewards, ok := c.stats.RewardsStats.DelegateesEpochRewards[delegatee]
	if !ok {
		delegateeEpochRewards = &DelegateeEpochRewards{
			TotalRewards:           make(map[RewardType]*EpochReward),
			DelegatorsEpochRewards: make(map[common.Address]*DelegatorEpochRewards),
		}
		c.stats.RewardsStats.DelegateesEpochRewards[delegatee] = delegateeEpochRewards
	}
	totalReward, ok := delegateeEpochRewards.TotalRewards[rewardType]
	if !ok {
		totalReward = &EpochReward{
			Balance: new(big.Int),
			Stake:   new(big.Int),
		}
		delegateeEpochRewards.TotalRewards[rewardType] = totalReward
	}
	totalReward.Balance.Add(totalReward.Balance, balance)
	totalReward.Stake.Add(totalReward.Stake, stake)
	delegatorEpochRewards, ok := delegateeEpochRewards.DelegatorsEpochRewards[delegator]
	if !ok {
		delegatorEpochRewards = &DelegatorEpochRewards{
			EpochRewards: make(map[RewardType]*EpochReward),
		}
		delegateeEpochRewards.DelegatorsEpochRewards[delegator] = delegatorEpochRewards
	}
	delegatorEpochReward, ok := delegatorEpochRewards.EpochRewards[rewardType]
	if !ok {
		delegatorEpochReward = &EpochReward{
			Balance: new(big.Int),
			Stake:   new(big.Int),
		}
		delegatorEpochRewards.EpochRewards[rewardType] = delegatorEpochReward
	}
	delegatorEpochReward.Balance.Add(delegatorEpochReward.Balance, balance)
	delegatorEpochReward.Stake.Add(delegatorEpochReward.Stake, stake)
}

func (c *statsCollector) AddProposerReward(balanceDest, stakeDest common.Address, balance, stake *big.Int, stakeWeight *big.Float) {
	if balanceDest == stakeDest {
		c.addMiningReward(balanceDest, balance, stake, stakeWeight, true)
		return
	}
	c.addMiningReward(balanceDest, balance, new(big.Int), stakeWeight, true)
	c.addMiningReward(stakeDest, new(big.Int), stake, stakeWeight, true)
}

func (c *statsCollector) AddFinalCommitteeReward(balanceDest, stakeDest common.Address, balance *big.Int, stake *big.Int, stakeWeight *big.Float) {
	if balance == nil {
		balance = new(big.Int)
	}
	if stake == nil {
		stake = new(big.Int)
	}
	if balance.Sign() != 0 || stake.Sign() != 0 {
		if stakeWeight == nil {
			stakeWeight = math.Zero()
		}
		if balanceDest == stakeDest {
			c.addFinalCommitteeReward(balanceDest, balance, stake, stakeWeight)
		} else {
			c.addFinalCommitteeReward(balanceDest, balance, new(big.Int), stakeWeight)
			c.addFinalCommitteeReward(stakeDest, new(big.Int), stake, stakeWeight)
		}
	}
	if c.stats.OriginalFinalCommittee == nil {
		c.stats.OriginalFinalCommittee = make(map[common.Address]struct{})
	}
	c.stats.OriginalFinalCommittee[stakeDest] = struct{}{}
	if c.stats.PoolFinalCommittee == nil {
		c.stats.PoolFinalCommittee = make(map[common.Address]struct{})
	}
	c.stats.PoolFinalCommittee[balanceDest] = struct{}{}
}

func (c *statsCollector) initFinalCommitteeRewardsByAddr() {
	if c.pending.finalCommitteeRewardsByAddr != nil {
		return
	}
	c.pending.finalCommitteeRewardsByAddr = make(map[common.Address]*MiningReward)
}

func (c *statsCollector) addFinalCommitteeReward(addr common.Address, balance *big.Int, stake *big.Int, stakeWeight *big.Float) {
	c.initFinalCommitteeRewardsByAddr()
	reward, ok := c.pending.finalCommitteeRewardsByAddr[addr]
	if ok {
		reward.Balance.Add(reward.Balance, balance)
		reward.Stake.Add(reward.Stake, stake)
		reward.StakeWeight.Add(reward.StakeWeight, stakeWeight)
		return
	}
	if reward = c.addMiningReward(addr, balance, stake, stakeWeight, false); reward != nil {
		c.pending.finalCommitteeRewardsByAddr[addr] = reward
	}
}

func (c *statsCollector) addMiningReward(addr common.Address, balance *big.Int, stake *big.Int, stakeWeight *big.Float, isProposerReward bool) *MiningReward {
	if balance == nil {
		balance = new(big.Int)
	}
	if stake == nil {
		stake = new(big.Int)
	}
	if balance.Sign() == 0 && stake.Sign() == 0 {
		return nil
	}
	if stakeWeight == nil {
		stakeWeight = math.Zero()
	}
	res := &MiningReward{
		Address:     addr,
		Balance:     new(big.Int).Set(balance),
		Stake:       new(big.Int).Set(stake),
		StakeWeight: math.Zero().Set(stakeWeight),
		Proposer:    isProposerReward,
	}
	c.stats.MiningRewards = append(c.stats.MiningRewards, res)
	return res
}

func (c *statsCollector) BeforeSetPenalty(addr common.Address, amount *big.Int, seconds uint16, appState *appstate.AppState) {
	c.addChargedPenalty(addr, amount)
	c.addChargedPenaltySeconds(addr, seconds)
}

func (c *statsCollector) addChargedPenalty(addr common.Address, amount *big.Int) {
	if amount == nil || amount.Sign() != 1 {
		return
	}
	c.initChargedPenaltiesByAddr()
	c.stats.ChargedPenaltiesByAddr[addr] = new(big.Int).Set(amount)
}

func (c *statsCollector) initChargedPenaltiesByAddr() {
	if c.stats.ChargedPenaltiesByAddr != nil {
		return
	}
	c.stats.ChargedPenaltiesByAddr = make(map[common.Address]*big.Int)
}

func (c *statsCollector) addChargedPenaltySeconds(addr common.Address, value uint16) {
	if value == 0 {
		return
	}
	c.initChargedPenaltySecondsByAddr()
	c.stats.ChargedPenaltySecondsByAddr[addr] = value
}

func (c *statsCollector) initChargedPenaltySecondsByAddr() {
	if c.stats.ChargedPenaltySecondsByAddr != nil {
		return
	}
	c.stats.ChargedPenaltySecondsByAddr = make(map[common.Address]uint16)
}

func (c *statsCollector) AddMintedCoins(amount *big.Int) {
	if amount == nil {
		return
	}
	if c.stats.MintedCoins == nil {
		c.stats.MintedCoins = big.NewInt(0)
	}
	c.stats.MintedCoins.Add(c.stats.MintedCoins, amount)
}

func (c *statsCollector) addBurntCoins(addr common.Address, amount *big.Int, reason db.BurntCoinsReason, tx *types.Transaction) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	if c.stats.BurntCoins == nil {
		c.stats.BurntCoins = big.NewInt(0)
	}
	c.stats.BurntCoins.Add(c.stats.BurntCoins, amount)
	if c.stats.BurntCoinsByAddr == nil {
		c.stats.BurntCoinsByAddr = make(map[common.Address][]*db.BurntCoins)
	}
	var txHash string
	if tx != nil {
		txHash = tx.Hash().Hex()
	}
	c.stats.BurntCoinsByAddr[addr] = append(c.stats.BurntCoinsByAddr[addr], &db.BurntCoins{
		Amount: blockchain.ConvertToFloat(amount),
		Reason: reason,
		TxHash: txHash,
	})
}

func (c *statsCollector) AddPenaltyBurntCoins(addr common.Address, amount *big.Int) {
	c.addBurntCoins(addr, amount, db.PenaltyBurntCoins, nil)
}

func (c *statsCollector) AddInviteBurntCoins(addr common.Address, amount *big.Int, tx *types.Transaction) {
	c.addBurntCoins(addr, amount, db.InviteBurntCoins, tx)
}

func (c *statsCollector) AddFeeBurntCoins(addr common.Address, feeAmount *big.Int, burntRate float32, tx *types.Transaction) {
	if feeAmount == nil || feeAmount.Sign() == 0 {
		return
	}
	burntFee := decimal.NewFromBigInt(feeAmount, 0)
	burntFee = burntFee.Mul(decimal.NewFromFloat32(burntRate))
	c.addBurntCoins(addr, math.ToInt(burntFee), db.FeeBurntCoins, tx)
}

func (c *statsCollector) AddKilledBurntCoins(addr common.Address, amount *big.Int) {
	var tx *types.Transaction
	if c.pending != nil && c.pending.tx != nil {
		tx = c.pending.tx.tx
	}
	c.addBurntCoins(addr, amount, db.KilledBurntCoins, tx)
}

func (c *statsCollector) AddBurnTxBurntCoins(addr common.Address, tx *types.Transaction) {
	c.addBurntCoins(addr, tx.AmountOrZero(), db.BurnTxBurntCoins, tx)
}

func (c *statsCollector) afterBalanceUpdate(addr common.Address) {
	c.initBalanceUpdatesByAddr()
	c.stats.BalanceUpdateAddrs.Add(addr)
}

func (c *statsCollector) initBalanceUpdatesByAddr() {
	if c.stats.BalanceUpdateAddrs != nil {
		return
	}
	c.stats.BalanceUpdateAddrs = mapset.NewSet()
}

func (c *statsCollector) CompleteCollecting() {
	c.stats = nil
	c.pending = nil
}

func (c *statsCollector) AfterAddStake(addr common.Address, amount *big.Int, appState *appstate.AppState) {
	if appState.State.GetIdentityState(addr) == state.Killed {
		c.addBurntCoins(addr, amount, db.KilledBurntCoins, nil)
	}
}

func (c *statsCollector) AddActivationTxBalanceTransfer(tx *types.Transaction, amount *big.Int) {
	sender, _ := types.Sender(tx)
	if sender == *tx.To {
		return
	}
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.stats.ActivationTxTransfers = append(c.stats.ActivationTxTransfers, db.ActivationTxTransfer{
		TxHash:          conversion.ConvertHash(tx.Hash()),
		BalanceTransfer: blockchain.ConvertToFloat(amount),
	})
}

func (c *statsCollector) AddKillTxStakeTransfer(tx *types.Transaction, amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.stats.KillTxTransfers = append(c.stats.KillTxTransfers, db.KillTxTransfer{
		TxHash:        conversion.ConvertHash(tx.Hash()),
		StakeTransfer: blockchain.ConvertToFloat(amount),
	})
}

func (c *statsCollector) AddKillInviteeTxStakeTransfer(tx *types.Transaction, stake, stakeToTransfer *big.Int) {

	if !common.ZeroOrNil(stakeToTransfer) {
		c.stats.KillInviteeTxTransfers = append(c.stats.KillInviteeTxTransfers, db.KillInviteeTxTransfer{
			TxHash:        conversion.ConvertHash(tx.Hash()),
			StakeTransfer: blockchain.ConvertToFloat(stakeToTransfer),
		})
	}

	if !common.ZeroOrNil(stake) && !common.ZeroOrNil(stakeToTransfer) && stake.Cmp(stakeToTransfer) > 0 {
		burntStake := new(big.Int).Sub(stake, stakeToTransfer)
		c.AddKilledBurntCoins(*tx.To, burntStake)
	}

}

func (c *statsCollector) BeginVerifiedStakeTransferBalanceUpdate(addrFrom, addrTo common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addrFrom, appState, db.VerifiedStakeTransferReason, nil, nil)
	if addrFrom != addrTo {
		c.addPendingBalanceUpdate(addrTo, appState, db.VerifiedStakeTransferReason, nil, nil)
	}
}

func (c *statsCollector) BeginTxBalanceUpdate(tx *types.Transaction, appState *appstate.AppState) {
	sender, _ := types.Sender(tx)
	txHash := tx.Hash()
	c.addPendingBalanceUpdate(sender, appState, db.TxReason, &txHash, nil)
	if tx.To != nil && *tx.To != sender {
		c.addPendingBalanceUpdate(*tx.To, appState, db.TxReason, &txHash, nil)
	}
}

func (c *statsCollector) BeginProposerRewardBalanceUpdate(balanceDest, stakeDest common.Address, potentialPenaltyPayment *big.Int, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(balanceDest, appState, db.ProposerRewardReason, nil, potentialPenaltyPayment)
	if balanceDest != stakeDest {
		c.addPendingBalanceUpdate(stakeDest, appState, db.ProposerRewardReason, nil, nil)
	}
}

func (c *statsCollector) BeginCommitteeRewardBalanceUpdate(balanceDest, stakeDest common.Address, potentialPenaltyPayment *big.Int, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(balanceDest, appState, db.CommitteeRewardReason, nil, potentialPenaltyPayment)
	if balanceDest != stakeDest {
		c.addPendingBalanceUpdate(stakeDest, appState, db.CommitteeRewardReason, nil, nil)
	}
}

func (c *statsCollector) BeginEpochRewardBalanceUpdate(balanceDest, stakeDest common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(balanceDest, appState, db.EpochRewardReason, nil, nil)
	if balanceDest != stakeDest {
		c.addPendingBalanceUpdate(stakeDest, appState, db.EpochRewardReason, nil, nil)
	}
}

func (c *statsCollector) BeginFailedValidationBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.FailedValidationReason, nil, nil)
}

func (c *statsCollector) BeginPenaltyBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.PenaltyReason, nil, nil)
}

func (c *statsCollector) BeginEpochPenaltyResetBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.EpochPenaltyResetReason, nil, nil)
}

func (c *statsCollector) BeginDustClearingBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.DustClearingReason, nil, nil)
}

func (c *statsCollector) CompleteBalanceUpdate(appState *appstate.AppState) {
	balanceUpdates := c.completeBalanceUpdates(appState)

	if len(balanceUpdates) == 2 && balanceUpdates[0].Reason == balanceUpdates[1].Reason &&
		balanceUpdates[0].Reason == db.EpochRewardReason && balanceUpdates[0].Address != balanceUpdates[1].Address {
		var delegatorBalanceUpdate, delegateeBalanceUpdate *db.BalanceUpdate
		if !isBalanceChanged(balanceUpdates[0]) {
			delegatorBalanceUpdate = balanceUpdates[0]
			delegateeBalanceUpdate = balanceUpdates[1]
		} else {
			delegatorBalanceUpdate = balanceUpdates[1]
			delegateeBalanceUpdate = balanceUpdates[0]
		}
		rewardBalanceUpdate := &db.BalanceUpdate{
			Address:           delegatorBalanceUpdate.Address,
			BalanceOld:        delegatorBalanceUpdate.BalanceOld,
			StakeOld:          delegatorBalanceUpdate.StakeOld,
			PenaltyOld:        delegatorBalanceUpdate.PenaltyOld,
			PenaltySecondsOld: delegatorBalanceUpdate.PenaltySecondsOld,
			BalanceNew: new(big.Int).Add(delegatorBalanceUpdate.BalanceOld,
				new(big.Int).Sub(delegateeBalanceUpdate.BalanceNew, delegateeBalanceUpdate.BalanceOld)),
			StakeNew:          delegatorBalanceUpdate.StakeNew,
			PenaltyNew:        delegatorBalanceUpdate.PenaltyNew,
			PenaltySecondsNew: delegatorBalanceUpdate.PenaltySecondsNew,
			PenaltyPayment:    nil,
			TxHash:            nil,
			Reason:            db.EpochRewardReason,
		}
		delegatorBalanceUpdate = &db.BalanceUpdate{
			Address:           rewardBalanceUpdate.Address,
			BalanceOld:        rewardBalanceUpdate.BalanceNew,
			StakeOld:          rewardBalanceUpdate.StakeNew,
			PenaltyOld:        rewardBalanceUpdate.PenaltyNew,
			PenaltySecondsOld: rewardBalanceUpdate.PenaltySecondsNew,
			BalanceNew:        delegatorBalanceUpdate.BalanceNew,
			StakeNew:          delegatorBalanceUpdate.StakeNew,
			PenaltyNew:        delegatorBalanceUpdate.PenaltyNew,
			PenaltySecondsNew: delegatorBalanceUpdate.PenaltySecondsNew,
			PenaltyPayment:    nil,
			TxHash:            nil,
			Reason:            db.DelegatorEpochRewardReason,
		}
		delegateeBalanceUpdate.Reason = db.DelegateeEpochRewardReason
		if !c.handleMultipleBalanceUpdate(rewardBalanceUpdate) {
			c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, rewardBalanceUpdate)
			c.afterBalanceUpdate(rewardBalanceUpdate.Address)
		}
		if !c.handleMultipleBalanceUpdate(delegatorBalanceUpdate) {
			c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, delegatorBalanceUpdate)
			c.afterBalanceUpdate(delegatorBalanceUpdate.Address)
		}
		if !c.handleMultipleBalanceUpdate(delegateeBalanceUpdate) {
			c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, delegateeBalanceUpdate)
			c.afterBalanceUpdate(delegateeBalanceUpdate.Address)
		}
		return
	}

	for _, balanceUpdate := range balanceUpdates {
		if !isBalanceOrStakeOrPenaltyChanged(balanceUpdate) {
			continue
		}
		if balanceUpdate.Reason == db.DustClearingReason {
			c.addBurntCoins(balanceUpdate.Address, balanceUpdate.BalanceOld, db.DustClearingBurntCoins, nil)
		}
		if balanceUpdate.Reason == db.EpochRewardReason || balanceUpdate.Reason == db.CommitteeRewardReason || balanceUpdate.Reason == db.VerifiedStakeTransferReason {
			if c.handleMultipleBalanceUpdate(balanceUpdate) {
				continue
			}
		}
		c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, balanceUpdate)
		c.afterBalanceUpdate(balanceUpdate.Address)
	}
}

func (c *statsCollector) getPendingBalanceUpdate(reason db.BalanceUpdateReason, addr common.Address) (*db.BalanceUpdate, bool) {
	if c.pending.balanceUpdatesByReasonAndAddr == nil {
		return nil, false
	}
	reasonBalanceUpdatesByAddr, ok := c.pending.balanceUpdatesByReasonAndAddr[reason]
	if !ok {
		return nil, false
	}
	balanceUpdate, ok := reasonBalanceUpdatesByAddr[addr]
	return balanceUpdate, ok
}

func (c *statsCollector) handleMultipleBalanceUpdate(balanceUpdate *db.BalanceUpdate) bool {
	if c.pending.balanceUpdatesByReasonAndAddr == nil {
		c.pending.balanceUpdatesByReasonAndAddr = make(map[db.BalanceUpdateReason]map[common.Address]*db.BalanceUpdate)
	}
	balanceUpdatesByAddr, ok := c.pending.balanceUpdatesByReasonAndAddr[balanceUpdate.Reason]
	if !ok {
		balanceUpdatesByAddr = make(map[common.Address]*db.BalanceUpdate)
		c.pending.balanceUpdatesByReasonAndAddr[balanceUpdate.Reason] = balanceUpdatesByAddr
	}
	if bu, ok := balanceUpdatesByAddr[balanceUpdate.Address]; ok {
		bu.BalanceNew = balanceUpdate.BalanceNew
		bu.StakeNew = balanceUpdate.StakeNew
		bu.PenaltyNew = balanceUpdate.PenaltyNew
		bu.PenaltySecondsNew = balanceUpdate.PenaltySecondsNew
		if balanceUpdate.PenaltyPayment != nil {
			if bu.PenaltyPayment == nil {
				bu.PenaltyPayment = new(big.Int)
			}
			bu.PenaltyPayment = new(big.Int).Add(bu.PenaltyPayment, balanceUpdate.PenaltyPayment)
		}
		if balanceUpdate.Reason == db.DelegatorEpochRewardReason {
			bu.BalanceOld = balanceUpdate.BalanceOld
			bu.StakeOld = balanceUpdate.StakeOld
			bu.PenaltyOld = balanceUpdate.PenaltyOld
			bu.PenaltySecondsOld = balanceUpdate.PenaltySecondsOld
			bu.PenaltyPayment = balanceUpdate.PenaltyPayment
		}
		c.updateLinkedBalanceUpdates(bu)
		return true
	}
	balanceUpdatesByAddr[balanceUpdate.Address] = balanceUpdate

	if balanceUpdate.Reason == db.EpochRewardReason {
		if delegatorBalanceUpdate, delegatorBalanceUpdatePresent := c.getPendingBalanceUpdate(db.DelegatorEpochRewardReason, balanceUpdate.Address); delegatorBalanceUpdatePresent {
			if c.pending.linkedBalanceUpdatesByNext == nil {
				c.pending.linkedBalanceUpdatesByNext = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
			}
			c.pending.linkedBalanceUpdatesByNext[balanceUpdate] = delegatorBalanceUpdate
			if c.pending.linkedBalanceUpdatesByPrev == nil {
				c.pending.linkedBalanceUpdatesByPrev = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
			}
			c.pending.linkedBalanceUpdatesByPrev[delegatorBalanceUpdate] = balanceUpdate
		} else {
			delegateeBalanceUpdate, delegateeBalanceUpdatePresent := c.getPendingBalanceUpdate(db.DelegateeEpochRewardReason, balanceUpdate.Address)
			if delegateeBalanceUpdatePresent {
				if c.pending.linkedBalanceUpdatesByNext == nil {
					c.pending.linkedBalanceUpdatesByNext = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
				}
				c.pending.linkedBalanceUpdatesByNext[balanceUpdate] = delegateeBalanceUpdate
				if c.pending.linkedBalanceUpdatesByPrev == nil {
					c.pending.linkedBalanceUpdatesByPrev = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
				}
				c.pending.linkedBalanceUpdatesByPrev[delegateeBalanceUpdate] = balanceUpdate
			}
		}
	}
	if balanceUpdate.Reason == db.DelegatorEpochRewardReason || balanceUpdate.Reason == db.DelegateeEpochRewardReason {
		if epochRewardBalanceUpdate, epochRewardBalanceUpdatePresent := c.getPendingBalanceUpdate(db.EpochRewardReason, balanceUpdate.Address); epochRewardBalanceUpdatePresent {
			if c.pending.linkedBalanceUpdatesByNext == nil {
				c.pending.linkedBalanceUpdatesByNext = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
			}
			c.pending.linkedBalanceUpdatesByNext[balanceUpdate] = epochRewardBalanceUpdate
			if c.pending.linkedBalanceUpdatesByPrev == nil {
				c.pending.linkedBalanceUpdatesByPrev = make(map[*db.BalanceUpdate]*db.BalanceUpdate)
			}
			c.pending.linkedBalanceUpdatesByPrev[epochRewardBalanceUpdate] = balanceUpdate
		}
	}
	c.updateLinkedBalanceUpdates(balanceUpdate)
	return false
}

func (c *statsCollector) updateLinkedBalanceUpdates(balanceUpdate *db.BalanceUpdate) {
	if balanceUpdate.Reason != db.DelegateeEpochRewardReason && balanceUpdate.Reason != db.EpochRewardReason &&
		balanceUpdate.Reason != db.DelegatorEpochRewardReason {
		return
	}
	if prevBalanceUpdate, prevBalanceUpdateOk := c.pending.linkedBalanceUpdatesByNext[balanceUpdate]; prevBalanceUpdateOk {
		balanceUpdate.BalanceOld = prevBalanceUpdate.BalanceNew
		balanceUpdate.StakeOld = prevBalanceUpdate.StakeNew
	} else if nextBalanceUpdate, nextBalanceUpdateOk := c.pending.linkedBalanceUpdatesByPrev[balanceUpdate]; nextBalanceUpdateOk {
		balanceUpdate.BalanceNew = new(big.Int).Sub(balanceUpdate.BalanceNew, new(big.Int).Sub(nextBalanceUpdate.BalanceNew, nextBalanceUpdate.BalanceOld))
		balanceUpdate.StakeNew = new(big.Int).Sub(balanceUpdate.StakeNew, new(big.Int).Sub(nextBalanceUpdate.StakeNew, nextBalanceUpdate.StakeOld))
		nextBalanceUpdate.BalanceNew = new(big.Int).Add(nextBalanceUpdate.BalanceNew, new(big.Int).Sub(balanceUpdate.BalanceNew, nextBalanceUpdate.BalanceOld))
		nextBalanceUpdate.BalanceOld = balanceUpdate.BalanceNew
		nextBalanceUpdate.StakeNew = new(big.Int).Add(nextBalanceUpdate.StakeNew, new(big.Int).Sub(balanceUpdate.StakeNew, nextBalanceUpdate.StakeOld))
		nextBalanceUpdate.StakeOld = balanceUpdate.StakeNew
	}
}

func isBalanceOrStakeOrPenaltyChanged(balanceUpdate *db.BalanceUpdate) bool {
	return balanceUpdate.BalanceOld.Cmp(balanceUpdate.BalanceNew) != 0 ||
		balanceUpdate.StakeOld.Cmp(balanceUpdate.StakeNew) != 0 ||
		valueOrZero(balanceUpdate.PenaltyOld).Cmp(valueOrZero(balanceUpdate.PenaltyNew)) != 0 ||
		balanceUpdate.PenaltySecondsOld != balanceUpdate.PenaltySecondsNew ||
		balanceUpdate.PenaltyPayment != nil && balanceUpdate.PenaltyPayment.Sign() > 0
}

func isBalanceChanged(balanceUpdate *db.BalanceUpdate) bool {
	return balanceUpdate.BalanceOld.Cmp(balanceUpdate.BalanceNew) != 0
}

func valueOrZero(v *big.Int) *big.Int {
	if v == nil {
		return common.Big0
	}
	return v
}

func (c *statsCollector) addPendingBalanceUpdate(
	addr common.Address,
	appState *appstate.AppState,
	reason db.BalanceUpdateReason,
	txHash *common.Hash,
	potentialPenaltyPayment *big.Int,
) {
	penaltySecondsOld := c.getPenaltySecondsIfNotKilled(addr, appState)
	var penaltyPayment *big.Int
	if potentialPenaltyPayment != nil && penaltySecondsOld > 0 {
		penaltyPayment = new(big.Int).Set(potentialPenaltyPayment)
	}
	c.pending.balanceUpdates = append(c.pending.balanceUpdates, &db.BalanceUpdate{
		Address:           addr,
		BalanceOld:        appState.State.GetBalance(addr),
		StakeOld:          c.getStakeIfNotKilled(addr, appState),
		PenaltyOld:        c.getPenaltyIfNotKilled(addr, appState),
		PenaltySecondsOld: penaltySecondsOld,
		PenaltyPayment:    penaltyPayment,
		Reason:            reason,
		TxHash:            txHash,
	})
}

func (c *statsCollector) completeBalanceUpdates(appState *appstate.AppState) []*db.BalanceUpdate {
	for _, balanceUpdate := range c.pending.balanceUpdates {
		balanceUpdate.BalanceNew = appState.State.GetBalance(balanceUpdate.Address)
		balanceUpdate.StakeNew = c.getStakeIfNotKilled(balanceUpdate.Address, appState)
		balanceUpdate.PenaltyNew = c.getPenaltyIfNotKilled(balanceUpdate.Address, appState)
		balanceUpdate.PenaltySecondsNew = c.getPenaltySecondsIfNotKilled(balanceUpdate.Address, appState)
	}
	balanceUpdates := c.pending.balanceUpdates
	c.pending.balanceUpdates = nil
	return balanceUpdates
}

func (c *statsCollector) getStakeIfNotKilled(addr common.Address, appState *appstate.AppState) *big.Int {
	if appState.State.GetIdentityState(addr) == state.Killed {
		return common.Big0
	}
	return appState.State.GetStakeBalance(addr)
}

func (c *statsCollector) getPenaltyIfNotKilled(addr common.Address, appState *appstate.AppState) *big.Int {
	if appState.State.GetIdentityState(addr) == state.Killed {
		return common.Big0
	}
	return appState.State.GetIdentity(addr).Penalty // State.GetPenalty is not used since it may add new identity to state
}

func (c *statsCollector) getPenaltySecondsIfNotKilled(addr common.Address, appState *appstate.AppState) uint16 {
	if appState.State.GetIdentityState(addr) == state.Killed {
		return 0
	}
	identity := appState.State.GetIdentity(addr)
	return identity.PenaltySeconds()
}

func (c *statsCollector) SetCommitteeRewardShare(amount *big.Int) {
	c.stats.CommitteeRewardShare = amount
}

func (c *statsCollector) BeginApplyingTx(tx *types.Transaction, appState *appstate.AppState) {
	c.pending.tx = &pendingTx{
		tx: tx,
	}
	sender, _ := types.Sender(tx)
	senderState := appState.State.GetIdentityState(sender)
	c.pending.identityStates = []state.IdentityState{senderState}
	if tx.To != nil && *tx.To != sender {
		recipientState := appState.State.GetIdentityState(*tx.To)
		c.pending.identityStates = append(c.pending.identityStates, recipientState)
	}
	if tx.Type == types.ActivationTx {
		senderIdentity := appState.State.GetIdentity(sender)
		inviter := senderIdentity.Inviter
		if inviter != nil {
			c.pending.tx.inviteTxHash = &inviter.TxHash
		}
	}
}

func (c *statsCollector) CompleteApplyingTx(appState *appstate.AppState) {
	tx := c.pending.tx.tx
	var changesByAddress map[common.Address]*IdentityStateChange
	initChangesByAddress := func() {
		if changesByAddress != nil {
			return
		}
		changesByAddress = make(map[common.Address]*IdentityStateChange)
	}
	sender, _ := types.Sender(tx)
	senderState := appState.State.GetIdentityState(sender)
	if c.pending.identityStates[0] != senderState {
		initChangesByAddress()
		changesByAddress[sender] = &IdentityStateChange{
			PrevState: c.pending.identityStates[0],
			NewState:  senderState,
		}
	}
	if tx.To != nil && *tx.To != sender {
		recipientState := appState.State.GetIdentityState(*tx.To)
		if c.pending.identityStates[1] != recipientState {
			initChangesByAddress()
			changesByAddress[*tx.To] = &IdentityStateChange{
				PrevState: c.pending.identityStates[1],
				NewState:  recipientState,
			}
		}
	}
	if len(changesByAddress) > 0 {
		if c.stats.IdentityStateChangesByTxHashAndAddress == nil {
			c.stats.IdentityStateChangesByTxHashAndAddress = make(map[common.Hash]map[common.Address]*IdentityStateChange)
		}
		c.stats.IdentityStateChangesByTxHashAndAddress[tx.Hash()] = changesByAddress
		for _, stateChange := range changesByAddress {
			if stateChange.PrevState == state.Candidate && stateChange.NewState != state.Candidate {
				c.stats.EpochSummaryUpdate.CandidateCountDiff--
			}
			if stateChange.NewState == state.Candidate && stateChange.PrevState != state.Candidate {
				c.stats.EpochSummaryUpdate.CandidateCountDiff++
			}
		}
	}

	c.collectTxSentToContract(appState)
	c.collectActivationTx(appState)

	c.pending.tx = nil
}

func (c *statsCollector) collectActivationTx(appState *appstate.AppState) {
	tx := c.pending.tx.tx
	if tx.Type != types.ActivationTx {
		return
	}
	recipient := *tx.To
	c.stats.ActivationTxs = append(c.stats.ActivationTxs, db.ActivationTx{
		TxHash:       conversion.ConvertHash(tx.Hash()),
		InviteTxHash: conversion.ConvertHash(*c.pending.tx.inviteTxHash),
		ShardId:      appState.State.ShardId(recipient),
	})
}

func (c *statsCollector) collectTxSentToContract(appState *appstate.AppState) {
	tx := c.pending.tx.tx
	if isContractTx(tx) || tx.To == nil {
		return
	}
	to := *tx.To
	if !isContractAddress(to, appState) {
		return
	}
	sender, _ := types.Sender(tx)
	senderContractTxBalanceUpdate := &db.ContractTxBalanceUpdate{
		Address:    sender,
		BalanceOld: nil,
		BalanceNew: nil,
	}
	c.stats.ContractTxsBalanceUpdates = append(c.stats.ContractTxsBalanceUpdates, &db.ContractTxBalanceUpdates{
		TxHash:          tx.Hash(),
		ContractAddress: to,
		Updates:         []*db.ContractTxBalanceUpdate{senderContractTxBalanceUpdate},
	})
}

func isContractAddress(address common.Address, appState *appstate.AppState) bool {
	return appState.State.GetCodeHash(address) != nil
}

func isContractTx(tx *types.Transaction) bool {
	return tx.Type == types.DeployContractTx || tx.Type == types.CallContractTx || tx.Type == types.TerminateContractTx
}

func (c *statsCollector) AddTxFee(feeAmount *big.Int) {
	tx := c.pending.tx.tx
	if c.stats.FeesByTxHash == nil {
		c.stats.FeesByTxHash = make(map[common.Hash]*big.Int)
	}
	c.stats.FeesByTxHash[tx.Hash()] = feeAmount
}

func (c *statsCollector) AddContractStake(amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	if c.pending.tx.oracleVotingContractDeploy != nil {
		c.pending.tx.oracleVotingContractDeploy.Stake = amount
	}
	if c.pending.tx.oracleLockContract != nil {
		c.pending.tx.oracleLockContract.Stake = amount
	}
	if c.pending.tx.refundableOracleLockContract != nil {
		c.pending.tx.refundableOracleLockContract.Stake = amount
	}
	if c.pending.tx.timeLockContract != nil {
		c.pending.tx.timeLockContract.Stake = amount
	}
	if c.pending.tx.multisigContract != nil {
		c.pending.tx.multisigContract.Stake = amount
	}
}

func (c *statsCollector) AddContractBalanceUpdate(address common.Address, getCurrentBalance collector.GetBalanceFunc, newBalance *big.Int, appState *appstate.AppState) {
	if c.pending.tx.contractBalanceUpdatesByAddr == nil {
		c.pending.tx.contractBalanceUpdatesByAddr = make(map[common.Address]*db.BalanceUpdate)
	}
	balanceUpdate, ok := c.pending.tx.contractBalanceUpdatesByAddr[address]
	if !ok {
		txHash := c.pending.tx.tx.Hash()
		var balanceOld *big.Int
		if b := getCurrentBalance(address); b != nil {
			balanceOld = new(big.Int).Set(b)
		}
		balanceUpdate = &db.BalanceUpdate{
			Address:           address,
			BalanceOld:        balanceOld,
			StakeOld:          c.getStakeIfNotKilled(address, appState),
			PenaltyOld:        c.getPenaltyIfNotKilled(address, appState),
			PenaltySecondsOld: c.getPenaltySecondsIfNotKilled(address, appState),
			PenaltyPayment:    nil,
			Reason:            db.EmbeddedContractReason,
			TxHash:            &txHash,
		}
		c.pending.tx.contractBalanceUpdatesByAddr[address] = balanceUpdate
	}
	balanceUpdate.BalanceNew = newBalance
	balanceUpdate.StakeNew = balanceUpdate.StakeOld
	balanceUpdate.PenaltyNew = balanceUpdate.PenaltyOld
	balanceUpdate.PenaltySecondsNew = balanceUpdate.PenaltySecondsOld
}

func (c *statsCollector) AddContractBurntCoins(address common.Address, getAmount collector.GetBalanceFunc) {
	amount := getAmount(address)
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.pending.tx.contractBurntCoins = append(c.pending.tx.contractBurntCoins, &pendingBurntCoins{
		address: address,
		amount:  amount,
		reason:  db.EmbeddedContractReason,
	})
}

func (c *statsCollector) AddContractTerminationBurntCoins(address common.Address, stake, refund *big.Int) {
	if stake == nil || refund == nil || stake.Cmp(refund) == 0 {
		return
	}
	amount := big.NewInt(0).Sub(stake, refund)
	c.pending.tx.contractBurntCoins = append(c.pending.tx.contractBurntCoins, &pendingBurntCoins{
		address: address,
		amount:  amount,
		reason:  db.EmbeddedContractTerminationReason,
	})
}

func (c *statsCollector) AddOracleVotingDeploy(contractAddress common.Address, startTime uint64, votingMinPayment *big.Int,
	fact []byte, state byte, votingDuration, publicVotingDuration uint64, winnerThreshold, quorum byte, committeeSize uint64, ownerFee byte) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractDeploy = &db.OracleVotingContract{
		TxHash:               tx.Hash(),
		ContractAddress:      contractAddress,
		Stake:                nil,
		StartTime:            startTime,
		VotingDuration:       votingDuration,
		VotingMinPayment:     votingMinPayment,
		Fact:                 fact,
		State:                state,
		PublicVotingDuration: publicVotingDuration,
		WinnerThreshold:      winnerThreshold,
		Quorum:               quorum,
		CommitteeSize:        committeeSize,
		OwnerFee:             ownerFee,
	}
}

func (c *statsCollector) AddOracleVotingCallStart(state byte, startBlock uint64, epoch uint16, votingMinPayment *big.Int, vrfSeed []byte,
	committeeSize uint64, networkSize int) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractCallStart = &db.OracleVotingContractCallStart{
		TxHash:           tx.Hash(),
		State:            state,
		StartHeight:      startBlock,
		Epoch:            epoch,
		VotingMinPayment: votingMinPayment,
		VrfSeed:          vrfSeed,
	}
	c.pending.tx.oracleVotingCommitteeStartCtx = &oracleVotingCommitteeStartCtx{
		committeeSize: committeeSize,
		networkSize:   networkSize,
	}
}

func (c *statsCollector) AddOracleVotingCallVoteProof(voteHash []byte, newSecretVotesCount *uint64, discriminated bool) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractCallVoteProof = &db.OracleVotingContractCallVoteProof{
		TxHash:              tx.Hash(),
		VoteHash:            voteHash,
		NewSecretVotesCount: newSecretVotesCount,
		Discriminated:       discriminated,
	}
}

func (c *statsCollector) AddOracleVotingCallVote(vote byte, salt []byte, newOptionVotes *uint64, newOptionAllVotes uint64,
	newSecretVotesCount *uint64, delegatee *common.Address, prevPoolVote []byte, newPrevOptionVotes *uint64, discriminated bool) {
	tx := c.pending.tx.tx
	var prevPoolVoteByte *byte
	if prevPoolVote != nil {
		v, _ := helpers.ExtractByte(0, prevPoolVote)
		prevPoolVoteByte = &v
	}
	c.pending.tx.oracleVotingContractCallVote = &db.OracleVotingContractCallVote{
		TxHash:           tx.Hash(),
		Vote:             vote,
		Salt:             salt,
		OptionVotes:      newOptionVotes,
		OptionAllVotes:   &newOptionAllVotes,
		SecretVotesCount: newSecretVotesCount,
		Delegatee:        delegatee,
		Discriminated:    discriminated,
		PrevPoolVote:     prevPoolVoteByte,
		PrevOptionVotes:  newPrevOptionVotes,
	}
}

func (c *statsCollector) AddOracleVotingCallFinish(state byte, result *byte, fund, oracleReward, ownerReward *big.Int) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractCallFinish = &db.OracleVotingContractCallFinish{
		TxHash:       tx.Hash(),
		State:        state,
		Result:       result,
		Fund:         fund,
		OracleReward: oracleReward,
		OwnerReward:  ownerReward,
	}
}

func (c *statsCollector) AddOracleVotingCallProlongation(startBlock *uint64, epoch uint16, vrfSeed []byte, committeeSize, networkSize uint64,
	newEpochWithoutGrowth *byte, newProlongVoteCount *uint64) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractCallProlongation = &db.OracleVotingContractCallProlongation{
		TxHash:             tx.Hash(),
		Epoch:              epoch,
		StartBlock:         startBlock,
		EpochWithoutGrowth: newEpochWithoutGrowth,
		ProlongVoteCount:   newProlongVoteCount,
		VrfSeed:            vrfSeed,
	}
	c.pending.tx.oracleVotingCommitteeStartCtx = &oracleVotingCommitteeStartCtx{
		committeeSize: committeeSize,
		networkSize:   int(networkSize),
	}
}

func (c *statsCollector) AddOracleVotingCallAddStake() {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractCallAddStake = &db.OracleVotingContractCallAddStake{
		TxHash: tx.Hash(),
	}
}

func (c *statsCollector) AddOracleVotingTermination(fund, oracleReward, ownerReward *big.Int) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleVotingContractTermination = &db.OracleVotingContractTermination{
		TxHash:       tx.Hash(),
		Fund:         fund,
		OracleReward: oracleReward,
		OwnerReward:  ownerReward,
	}
}

func (c *statsCollector) AddOracleLockDeploy(contractAddress common.Address, oracleVotingAddress common.Address, value byte, successAddress common.Address,
	failAddress common.Address) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleLockContract = &db.OracleLockContract{
		TxHash:              tx.Hash(),
		ContractAddress:     contractAddress,
		Stake:               nil,
		OracleVotingAddress: oracleVotingAddress,
		ExpectedValue:       value,
		SuccessAddress:      successAddress,
		FailAddress:         failAddress,
	}
}

func (c *statsCollector) AddOracleLockCallPush(success bool, oracleVotingResult byte, transfer *big.Int) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleLockContractCallPush = &db.OracleLockContractCallPush{
		TxHash:             tx.Hash(),
		Success:            success,
		OracleVotingResult: oracleVotingResult,
		Transfer:           transfer,
	}
}

func (c *statsCollector) AddOracleLockCallCheckOracleVoting(votedValue byte, err error) {
	tx := c.pending.tx.tx
	var result *byte
	if err == nil {
		result = &votedValue
	}
	c.pending.tx.oracleLockContractCallCheckOracleVoting = &db.OracleLockContractCallCheckOracleVoting{
		TxHash:             tx.Hash(),
		OracleVotingResult: result,
	}
}

func (c *statsCollector) AddOracleLockTermination(dest common.Address) {
	tx := c.pending.tx.tx
	c.pending.tx.oracleLockContractTermination = &db.OracleLockContractTermination{
		TxHash: tx.Hash(),
		Dest:   dest,
	}
}

func (c *statsCollector) AddRefundableOracleLockDeploy(contractAddress common.Address, oracleVotingAddress common.Address,
	value byte, successAddress common.Address, successAddressErr error, failAddress common.Address, failAddressErr error,
	refundDelay, depositDeadline uint64, oracleVotingFee byte, state byte, sum *big.Int) {
	tx := c.pending.tx.tx
	var successAddressP *common.Address
	if successAddressErr == nil {
		successAddressP = &successAddress
	}
	var failAddressP *common.Address
	if failAddressErr == nil {
		failAddressP = &failAddress
	}
	c.pending.tx.refundableOracleLockContract = &db.RefundableOracleLockContract{
		TxHash:              tx.Hash(),
		ContractAddress:     contractAddress,
		Stake:               nil,
		OracleVotingAddress: oracleVotingAddress,
		ExpectedValue:       value,
		SuccessAddress:      successAddressP,
		FailAddress:         failAddressP,
		RefundDelay:         refundDelay,
		DepositDeadline:     depositDeadline,
		OracleVotingFee:     oracleVotingFee,
	}
}

func (c *statsCollector) AddRefundableOracleLockCallDeposit(ownSum, sum, fee *big.Int) {
	tx := c.pending.tx.tx
	c.pending.tx.refundableOracleLockContractCallDeposit = &db.RefundableOracleLockContractCallDeposit{
		TxHash: tx.Hash(),
		OwnSum: ownSum,
		Sum:    sum,
		Fee:    fee,
	}
}

func (c *statsCollector) AddRefundableOracleLockCallPush(state byte, oracleVotingExists bool, oracleVotingResult byte, oracleVotingResultErr error, transfer *big.Int, refundBlock uint64) {
	tx := c.pending.tx.tx
	var result *byte
	if oracleVotingResultErr == nil {
		result = &oracleVotingResult
	}
	c.pending.tx.refundableOracleLockContractCallPush = &db.RefundableOracleLockContractCallPush{
		TxHash:             tx.Hash(),
		State:              state,
		OracleVotingExists: oracleVotingExists,
		OracleVotingResult: result,
		Transfer:           transfer,
		RefundBlock:        refundBlock,
	}
}

func (c *statsCollector) AddRefundableOracleLockCallRefund(balance *big.Int, coef decimal.Decimal) {
	tx := c.pending.tx.tx
	coefF, _ := coef.Float64()
	c.pending.tx.refundableOracleLockContractCallRefund = &db.RefundableOracleLockContractCallRefund{
		TxHash:  tx.Hash(),
		Balance: balance,
		Coef:    coefF,
	}
}

func (c *statsCollector) AddRefundableOracleLockTermination(dest common.Address) {
	tx := c.pending.tx.tx
	c.pending.tx.refundableOracleLockContractTermination = &db.RefundableOracleLockContractTermination{
		TxHash: tx.Hash(),
		Dest:   dest,
	}
}

func (c *statsCollector) AddMultisigDeploy(contractAddress common.Address, minVotes, maxVotes, state byte) {
	tx := c.pending.tx.tx
	c.pending.tx.multisigContract = &db.MultisigContract{
		TxHash:          tx.Hash(),
		ContractAddress: contractAddress,
		Stake:           nil,
		MinVotes:        minVotes,
		MaxVotes:        maxVotes,
		State:           state,
	}
}

func (c *statsCollector) AddMultisigCallAdd(address common.Address, newState *byte) {
	tx := c.pending.tx.tx
	c.pending.tx.multisigContractCallAdd = &db.MultisigContractCallAdd{
		TxHash:   tx.Hash(),
		Address:  address,
		NewState: newState,
	}
}

func (c *statsCollector) AddMultisigCallSend(dest common.Address, amount []byte) {
	tx := c.pending.tx.tx
	c.pending.tx.multisigContractCallSend = &db.MultisigContractCallSend{
		TxHash: tx.Hash(),
		Dest:   dest,
		Amount: big.NewInt(0).SetBytes(amount),
	}
}

func (c *statsCollector) AddMultisigCallPush(dest common.Address, amount []byte, voteAddressCnt, voteAmountCnt int) {
	tx := c.pending.tx.tx
	c.pending.tx.multisigContractCallPush = &db.MultisigContractCallPush{
		TxHash:         tx.Hash(),
		Dest:           dest,
		Amount:         big.NewInt(0).SetBytes(amount),
		VoteAddressCnt: byte(voteAddressCnt),
		VoteAmountCnt:  byte(voteAmountCnt),
	}
}

func (c *statsCollector) AddMultisigTermination(dest common.Address) {
	tx := c.pending.tx.tx
	c.pending.tx.multisigContractTermination = &db.MultisigContractTermination{
		TxHash: tx.Hash(),
		Dest:   dest,
	}
}

func (c *statsCollector) AddTimeLockDeploy(contractAddress common.Address, timestamp uint64) {
	tx := c.pending.tx.tx
	c.pending.tx.timeLockContract = &db.TimeLockContract{
		TxHash:          tx.Hash(),
		ContractAddress: contractAddress,
		Stake:           nil,
		Timestamp:       timestamp,
	}
}

func (c *statsCollector) AddTimeLockCallTransfer(dest common.Address, amount *big.Int) {
	tx := c.pending.tx.tx
	c.pending.tx.timeLockContractCallTransfer = &db.TimeLockContractCallTransfer{
		TxHash: tx.Hash(),
		Dest:   dest,
		Amount: amount,
	}
}

func (c *statsCollector) AddTimeLockTermination(dest common.Address) {
	tx := c.pending.tx.tx
	c.pending.tx.timeLockContractTermination = &db.TimeLockContractTermination{
		TxHash: tx.Hash(),
		Dest:   dest,
	}
}

func (c *statsCollector) AddTxReceipt(txReceipt *types.TxReceipt, appState *appstate.AppState) {
	var errorMsg string
	if txReceipt.Error != nil {
		errorMsg = txReceipt.Error.Error()
	}
	c.stats.TxReceipts = append(c.stats.TxReceipts, &db.TxReceipt{
		TxHash:  txReceipt.TxHash,
		Success: txReceipt.Success,
		GasUsed: txReceipt.GasUsed,
		GasCost: txReceipt.GasCost,
		Method:  txReceipt.Method,
		Error:   errorMsg,
	})

	sender, _ := types.Sender(c.pending.tx.tx)
	senderContractTxBalanceUpdate := &db.ContractTxBalanceUpdate{
		Address:    sender,
		BalanceOld: nil,
		BalanceNew: nil,
	}
	updates := []*db.ContractTxBalanceUpdate{senderContractTxBalanceUpdate}
	var contractCallMethod *db.ContractCallMethod
	if txReceipt.Success {
		if c.pending.tx.oracleVotingContractDeploy != nil {
			c.stats.OracleVotingContracts = append(c.stats.OracleVotingContracts, c.pending.tx.oracleVotingContractDeploy)
		}
		if c.pending.tx.oracleVotingContractCallStart != nil {
			callMethod := db.OracleVotingCallStart
			contractCallMethod = &callMethod
			oracleVotingContractCallStart := c.pending.tx.oracleVotingContractCallStart
			ctx := c.pending.tx.oracleVotingCommitteeStartCtx
			oracleVotingContractCallStart.Committee = c.getOracleVotingCommittee(
				ctx.committeeSize,
				ctx.networkSize,
				oracleVotingContractCallStart.VrfSeed,
				appState,
			)
			c.stats.OracleVotingContractCallStarts = append(c.stats.OracleVotingContractCallStarts, oracleVotingContractCallStart)
			contractAddress := c.pending.tx.tx.To
			c.stats.NewActualOracleVotingContracts = append(c.stats.NewActualOracleVotingContracts, *contractAddress)
		}
		if c.pending.tx.oracleVotingContractCallVoteProof != nil {
			callMethod := db.OracleVotingCallVoteProof
			contractCallMethod = &callMethod
			c.stats.OracleVotingContractCallVoteProofs = append(c.stats.OracleVotingContractCallVoteProofs, c.pending.tx.oracleVotingContractCallVoteProof)
		}
		if c.pending.tx.oracleVotingContractCallVote != nil {
			callMethod := db.OracleVotingCallVote
			contractCallMethod = &callMethod
			c.stats.OracleVotingContractCallVotes = append(c.stats.OracleVotingContractCallVotes, c.pending.tx.oracleVotingContractCallVote)
		}
		if c.pending.tx.oracleVotingContractCallFinish != nil {
			callMethod := db.OracleVotingCallFinish
			contractCallMethod = &callMethod
			c.stats.OracleVotingContractCallFinishes = append(c.stats.OracleVotingContractCallFinishes, c.pending.tx.oracleVotingContractCallFinish)
			contractAddress := c.pending.tx.tx.To
			c.stats.NewNotActualOracleVotingContracts = append(c.stats.NewNotActualOracleVotingContracts, *contractAddress)
		}
		if c.pending.tx.oracleVotingContractCallProlongation != nil {
			callMethod := db.OracleVotingCallProlong
			contractCallMethod = &callMethod
			oracleVotingContractCallProlongation := c.pending.tx.oracleVotingContractCallProlongation
			ctx := c.pending.tx.oracleVotingCommitteeStartCtx
			oracleVotingContractCallProlongation.Committee = c.getOracleVotingCommittee(
				ctx.committeeSize,
				ctx.networkSize,
				oracleVotingContractCallProlongation.VrfSeed,
				appState,
			)
			c.stats.OracleVotingContractCallProlongations = append(c.stats.OracleVotingContractCallProlongations, oracleVotingContractCallProlongation)
			contractAddress := c.pending.tx.tx.To
			c.stats.NewActualOracleVotingContracts = append(c.stats.NewActualOracleVotingContracts, *contractAddress)
		}
		if c.pending.tx.oracleVotingContractCallAddStake != nil {
			callMethod := db.OracleVotingCallAddStake
			contractCallMethod = &callMethod
			c.stats.OracleVotingContractCallAddStakes = append(c.stats.OracleVotingContractCallAddStakes, c.pending.tx.oracleVotingContractCallAddStake)
		}
		if c.pending.tx.oracleVotingContractTermination != nil {
			c.stats.OracleVotingContractTerminations = append(c.stats.OracleVotingContractTerminations, c.pending.tx.oracleVotingContractTermination)
			contractAddress := c.pending.tx.tx.To
			c.stats.NewNotActualOracleVotingContracts = append(c.stats.NewNotActualOracleVotingContracts, *contractAddress)
		}
		if c.pending.tx.oracleLockContract != nil {
			c.stats.OracleLockContracts = append(c.stats.OracleLockContracts, c.pending.tx.oracleLockContract)
		}
		if c.pending.tx.oracleLockContractCallCheckOracleVoting != nil {
			callMethod := db.OracleLockCallCheckOracleVoting
			contractCallMethod = &callMethod
			c.stats.OracleLockContractCallCheckOracleVotings = append(c.stats.OracleLockContractCallCheckOracleVotings, c.pending.tx.oracleLockContractCallCheckOracleVoting)
		}
		if c.pending.tx.oracleLockContractCallPush != nil {
			callMethod := db.OracleLockCallPush
			contractCallMethod = &callMethod
			c.stats.OracleLockContractCallPushes = append(c.stats.OracleLockContractCallPushes, c.pending.tx.oracleLockContractCallPush)
		}
		if c.pending.tx.oracleLockContractTermination != nil {
			c.stats.OracleLockContractTerminations = append(c.stats.OracleLockContractTerminations, c.pending.tx.oracleLockContractTermination)
		}
		if c.pending.tx.refundableOracleLockContract != nil {
			c.stats.RefundableOracleLockContracts = append(c.stats.RefundableOracleLockContracts, c.pending.tx.refundableOracleLockContract)
		}
		if c.pending.tx.refundableOracleLockContractCallDeposit != nil {
			callMethod := db.RefundableOracleLockCallDeposit
			contractCallMethod = &callMethod
			c.stats.RefundableOracleLockContractCallDeposits = append(c.stats.RefundableOracleLockContractCallDeposits, c.pending.tx.refundableOracleLockContractCallDeposit)
		}
		if c.pending.tx.refundableOracleLockContractCallPush != nil {
			callMethod := db.RefundableOracleLockCallPush
			contractCallMethod = &callMethod
			c.stats.RefundableOracleLockContractCallPushes = append(c.stats.RefundableOracleLockContractCallPushes, c.pending.tx.refundableOracleLockContractCallPush)
		}
		if c.pending.tx.refundableOracleLockContractCallRefund != nil {
			callMethod := db.RefundableOracleLockCallRefund
			contractCallMethod = &callMethod
			c.stats.RefundableOracleLockContractCallRefunds = append(c.stats.RefundableOracleLockContractCallRefunds, c.pending.tx.refundableOracleLockContractCallRefund)
		}
		if c.pending.tx.refundableOracleLockContractTermination != nil {
			c.stats.RefundableOracleLockContractTerminations = append(c.stats.RefundableOracleLockContractTerminations, c.pending.tx.refundableOracleLockContractTermination)
		}
		if c.pending.tx.multisigContract != nil {
			c.stats.MultisigContracts = append(c.stats.MultisigContracts, c.pending.tx.multisigContract)
		}
		if c.pending.tx.multisigContractCallAdd != nil {
			callMethod := db.MultisigCallAdd
			contractCallMethod = &callMethod
			c.stats.MultisigContractCallAdds = append(c.stats.MultisigContractCallAdds, c.pending.tx.multisigContractCallAdd)
		}
		if c.pending.tx.multisigContractCallSend != nil {
			callMethod := db.MultisigCallSend
			contractCallMethod = &callMethod
			c.stats.MultisigContractCallSends = append(c.stats.MultisigContractCallSends, c.pending.tx.multisigContractCallSend)
		}
		if c.pending.tx.multisigContractCallPush != nil {
			callMethod := db.MultisigCallPush
			contractCallMethod = &callMethod
			c.stats.MultisigContractCallPushes = append(c.stats.MultisigContractCallPushes, c.pending.tx.multisigContractCallPush)
		}
		if c.pending.tx.multisigContractTermination != nil {
			c.stats.MultisigContractTerminations = append(c.stats.MultisigContractTerminations, c.pending.tx.multisigContractTermination)
		}
		if c.pending.tx.timeLockContract != nil {
			c.stats.TimeLockContracts = append(c.stats.TimeLockContracts, c.pending.tx.timeLockContract)
		}
		if c.pending.tx.timeLockContractCallTransfer != nil {
			callMethod := db.TimeLockCallTransfer
			contractCallMethod = &callMethod
			c.stats.TimeLockContractCallTransfers = append(c.stats.TimeLockContractCallTransfers, c.pending.tx.timeLockContractCallTransfer)
		}
		if c.pending.tx.timeLockContractTermination != nil {
			c.stats.TimeLockContractTerminations = append(c.stats.TimeLockContractTerminations, c.pending.tx.timeLockContractTermination)
		}
		for _, balanceUpdate := range c.pending.tx.contractBalanceUpdatesByAddr {
			if !isBalanceOrStakeOrPenaltyChanged(balanceUpdate) {
				continue
			}
			c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, balanceUpdate)
			c.afterBalanceUpdate(balanceUpdate.Address)
			if balanceUpdate.Address == sender {
				senderContractTxBalanceUpdate.BalanceOld = balanceUpdate.BalanceOld
				senderContractTxBalanceUpdate.BalanceNew = balanceUpdate.BalanceNew
			} else {
				updates = append(updates, &db.ContractTxBalanceUpdate{
					Address:    balanceUpdate.Address,
					BalanceOld: balanceUpdate.BalanceOld,
					BalanceNew: balanceUpdate.BalanceNew,
				})
			}
		}
		for _, addressBurntCoins := range c.pending.tx.contractBurntCoins {
			c.addBurntCoins(addressBurntCoins.address, addressBurntCoins.amount, addressBurntCoins.reason, c.pending.tx.tx)
		}
	}

	isFailedDeploy := c.pending.tx.tx.Type == types.DeployContractTx && !txReceipt.Success
	if txReceipt.ContractAddress != (common.Address{}) && !isFailedDeploy {
		c.stats.ContractTxsBalanceUpdates = append(c.stats.ContractTxsBalanceUpdates, &db.ContractTxBalanceUpdates{
			TxHash:             c.pending.tx.tx.Hash(),
			ContractAddress:    txReceipt.ContractAddress,
			ContractCallMethod: contractCallMethod,
			Updates:            updates,
		})
	}

}

func (c *statsCollector) getOracleVotingCommittee(committeeSize uint64, networkSize int, vrfSeed []byte, appState *appstate.AppState) []common.Address {
	var res []common.Address
	checkAndAdd := func(addr common.Address, pubKey []byte, state state.IdentityState) {
		if !state.NewbieOrBetter() {
			return
		}
		selectionHash := crypto.Hash(append(pubKey, vrfSeed...))
		v := new(big.Float).SetInt(new(big.Int).SetBytes(selectionHash[:]))
		q := new(big.Float).Quo(v, maxOracleVotingHash)
		networkSizeF := float64(networkSize)
		if networkSize == 0 {
			networkSizeF = 1
		}
		if q.Cmp(big.NewFloat(1-float64(committeeSize)/networkSizeF)) < 0 {
			return
		}
		res = append(res, addr)
	}
	if len(c.pending.identitiesByAddr) > 0 {
		for addr, identity := range c.pending.identitiesByAddr {
			checkAndAdd(addr, identity.pubKey, identity.state)
		}
		return res
	}
	c.pending.identitiesByAddr = make(map[common.Address]*identityInfo)
	appState.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		c.pending.identitiesByAddr[addr] = &identityInfo{
			pubKey: identity.PubKey,
			state:  identity.State,
		}
		checkAndAdd(addr, identity.PubKey, identity.State)
	})
	return res
}

func (c *statsCollector) AddRemovedTransitiveDelegation(delegator, delegatee common.Address) {
	c.stats.RemovedTransitiveDelegations = append(c.stats.RemovedTransitiveDelegations, db.RemovedTransitiveDelegation{
		Delegator: delegator,
		Delegatee: delegatee,
	})
}

func (c *statsCollector) BeginSavedStakeBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	// do nothing
}

func (c *statsCollector) Disable() {
	c.statsEnabled = false
}

func (c *statsCollector) Enable() {
	c.statsEnabled = true
}

func (c *statsCollector) GetStats() *Stats {
	if !c.statsEnabled {
		return &Stats{}
	}
	return c.stats
}
