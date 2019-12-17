package stats

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/stats/collector"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"github.com/shopspring/decimal"
	"math/big"
)

const (
	proposerReward       = "Proposer"
	finalCommitteeReward = "FinalCommittee"
)

type statsCollector struct {
	stats                          *Stats
	invitationRewardsByAddrAndType map[common.Address]map[RewardType]*RewardStats
}

func NewStatsCollector() collector.StatsCollector {
	return &statsCollector{}
}

func (c *statsCollector) EnableCollecting() {
	c.stats = &Stats{}
}

func (c *statsCollector) initRewardStats() {
	if c.stats.RewardsStats != nil {
		return
	}
	c.stats.RewardsStats = &RewardsStats{}
}

func (c *statsCollector) initInvitationRewardsByAddrAndType() {
	if c.invitationRewardsByAddrAndType != nil {
		return
	}
	c.invitationRewardsByAddrAndType = make(map[common.Address]map[RewardType]*RewardStats)
}

func (c *statsCollector) SetValidation(validation *statsTypes.ValidationStats) {
	c.stats.ValidationStats = validation
}

func (c *statsCollector) SetAuthors(authors *types.ValidationAuthors) {
	c.initRewardStats()
	c.stats.RewardsStats.Authors = authors
}

func (c *statsCollector) SetTotalReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Total = amount
}

func (c *statsCollector) SetTotalValidationReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Validation = amount
}

func (c *statsCollector) SetTotalFlipsReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Flips = amount
}

func (c *statsCollector) SetTotalInvitationsReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Invitations = amount
}

func (c *statsCollector) SetTotalFoundationPayouts(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.FoundationPayouts = amount
}

func (c *statsCollector) SetTotalZeroWalletFund(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.ZeroWalletFund = amount
}

func (c *statsCollector) AddValidationReward(addr common.Address, age uint16, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Validation)
	if c.stats.RewardsStats.AgesByAddress == nil {
		c.stats.RewardsStats.AgesByAddress = make(map[string]uint16)
	}
	c.stats.RewardsStats.AgesByAddress[conversion.ConvertAddress(addr)] = age + 1
}

func (c *statsCollector) AddFlipsReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Flips)
}

func (c *statsCollector) AddInvitationsReward(addr common.Address, balance *big.Int, stake *big.Int, age uint16) {
	var rewardType RewardType
	switch age {
	case 1:
		rewardType = Invitations
	case 2:
		rewardType = Invitations2
	case 3:
		rewardType = Invitations3
	default:
		return
	}
	c.addReward(addr, balance, stake, rewardType)
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
	if c.increaseInvitationRewardIfExists(rewardsStats) {
		return
	}
	c.stats.RewardsStats.Rewards = append(c.stats.RewardsStats.Rewards, rewardsStats)
}

func (c *statsCollector) increaseInvitationRewardIfExists(rewardsStats *RewardStats) bool {
	if rewardsStats.Type != Invitations && rewardsStats.Type != Invitations2 && rewardsStats.Type != Invitations3 {
		return false
	}
	c.initInvitationRewardsByAddrAndType()
	addrInvitationRewardsByType, ok := c.invitationRewardsByAddrAndType[rewardsStats.Address]
	if ok {
		if ir, ok := addrInvitationRewardsByType[rewardsStats.Type]; ok {
			ir.Balance.Add(ir.Balance, rewardsStats.Balance)
			ir.Stake.Add(ir.Stake, rewardsStats.Stake)
			return true
		}
	} else {
		addrInvitationRewardsByType = make(map[RewardType]*RewardStats)
	}
	addrInvitationRewardsByType[rewardsStats.Type] = rewardsStats
	c.invitationRewardsByAddrAndType[rewardsStats.Address] = addrInvitationRewardsByType
	return false
}

func (c *statsCollector) AddProposerReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addMiningReward(addr, balance, stake, proposerReward)
}

func (c *statsCollector) AddFinalCommitteeReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addMiningReward(addr, balance, stake, finalCommitteeReward)
	c.stats.FinalCommittee = append(c.stats.FinalCommittee, addr)
}

func (c *statsCollector) addMiningReward(addr common.Address, balance *big.Int, stake *big.Int, rType string) {
	c.stats.MiningRewards = append(c.stats.MiningRewards, &db.Reward{
		Address: conversion.ConvertAddress(addr),
		Balance: blockchain.ConvertToFloat(balance),
		Stake:   blockchain.ConvertToFloat(stake),
		Type:    rType,
	})
}

func (c *statsCollector) AfterSubPenalty(addr common.Address, amount *big.Int, appState *appstate.AppState) {
	if amount == nil || amount.Sign() != 1 {
		return
	}
	c.detectAndCollectCompletedPenalty(addr, appState)
}

func (c *statsCollector) detectAndCollectCompletedPenalty(addr common.Address, appState *appstate.AppState) {
	updatedPenalty := appState.State.GetPenalty(addr)
	if updatedPenalty != nil && updatedPenalty.Sign() == 1 {
		return
	}
	c.initBurntPenaltiesByAddr()
	c.stats.BurntPenaltiesByAddr[addr] = updatedPenalty
}

func (c *statsCollector) BeforeClearPenalty(addr common.Address, appState *appstate.AppState) {
	c.detectAndCollectBurntPenalty(addr, appState)
}

func (c *statsCollector) BeforeSetPenalty(addr common.Address, appState *appstate.AppState) {
	c.detectAndCollectBurntPenalty(addr, appState)
}

func (c *statsCollector) detectAndCollectBurntPenalty(addr common.Address, appState *appstate.AppState) {
	curPenalty := appState.State.GetPenalty(addr)
	if curPenalty == nil || curPenalty.Sign() != 1 {
		return
	}
	c.initBurntPenaltiesByAddr()
	c.stats.BurntPenaltiesByAddr[addr] = curPenalty
}

func (c *statsCollector) initBurntPenaltiesByAddr() {
	if c.stats.BurntPenaltiesByAddr != nil {
		return
	}
	c.stats.BurntPenaltiesByAddr = make(map[common.Address]*big.Int)
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
	if c.stats.BurntCoinsByAddr == nil {
		c.stats.BurntCoinsByAddr = make(map[common.Address][]*db.BurntCoins)
	}
	c.stats.BurntCoins.Add(c.stats.BurntCoins, amount)
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
	c.addBurntCoins(addr, amount, db.KilledBurntCoins, nil)
}

func (c *statsCollector) AddBurnTxBurntCoins(addr common.Address, tx *types.Transaction) {
	c.addBurntCoins(addr, tx.AmountOrZero(), db.BurnTxBurntCoins, tx)
}

func (c *statsCollector) AfterBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.initBalanceUpdatesByAddr()
	c.stats.BalanceUpdateAddrs.Add(addr)
}

func (c *statsCollector) initBalanceUpdatesByAddr() {
	if c.stats.BalanceUpdateAddrs != nil {
		return
	}
	c.stats.BalanceUpdateAddrs = mapset.NewSet()
}

func (c *statsCollector) GetStats() *Stats {
	return c.stats
}

func (c *statsCollector) CompleteCollecting() {
	c.stats = nil
	c.invitationRewardsByAddrAndType = nil
}

func (c *statsCollector) AfterKillIdentity(addr common.Address, appState *appstate.AppState) {
	c.initKilledAddrs()
	c.stats.KilledAddrs.Add(addr)
}

func (c *statsCollector) initKilledAddrs() {
	if c.stats.KilledAddrs != nil {
		return
	}
	c.stats.KilledAddrs = mapset.NewSet()
}

func (c *statsCollector) AfterAddStake(addr common.Address, amount *big.Int) {
	if c.stats.KilledAddrs == nil || !c.stats.KilledAddrs.Contains(addr) {
		return
	}
	c.addBurntCoins(addr, amount, db.KilledBurntCoins, nil)
}
