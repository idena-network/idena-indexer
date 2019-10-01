package stats

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/stats/collector"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

const (
	proposerReward       = "Proposer"
	finalCommitteeReward = "FinalCommittee"
)

type blockStatsCollector struct {
	collect bool
	stats   *Stats
}

func NewBlockStatsCollector() collector.BlockStatsCollector {
	return &blockStatsCollector{
		collect: false,
	}
}

func (c *blockStatsCollector) EnableCollecting() {
	c.initStats()
	c.collect = true
}

func (c *blockStatsCollector) canCollect() bool {
	return c.collect
}

func (c *blockStatsCollector) initStats() {
	if c.stats != nil {
		return
	}
	c.stats = &Stats{}
}

func (c *blockStatsCollector) initRewardStats() {
	if c.stats.RewardsStats != nil {
		return
	}
	c.stats.RewardsStats = &RewardsStats{}
}

func (c *blockStatsCollector) SetValidation(validation *statsTypes.ValidationStats) {
	if !c.canCollect() {
		return
	}
	c.stats.ValidationStats = validation
}

func (c *blockStatsCollector) SetAuthors(authors *types.ValidationAuthors) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Authors = authors
}

func (c *blockStatsCollector) SetTotalReward(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Total = amount
}

func (c *blockStatsCollector) SetTotalValidationReward(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Validation = amount
}

func (c *blockStatsCollector) SetTotalFlipsReward(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Flips = amount
}

func (c *blockStatsCollector) SetTotalInvitationsReward(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Invitations = amount
}

func (c *blockStatsCollector) SetTotalFoundationPayouts(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.FoundationPayouts = amount
}

func (c *blockStatsCollector) SetTotalZeroWalletFund(amount *big.Int) {
	if !c.canCollect() {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.ZeroWalletFund = amount
}

func (c *blockStatsCollector) AddValidationReward(addr common.Address, age uint16, balance *big.Int, stake *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addReward(addr, balance, stake, Validation)
	if c.stats.RewardsStats.AgesByAddress == nil {
		c.stats.RewardsStats.AgesByAddress = make(map[string]uint16)
	}
	c.stats.RewardsStats.AgesByAddress[conversion.ConvertAddress(addr)] = age + 1
}

func (c *blockStatsCollector) AddFlipsReward(addr common.Address, balance *big.Int, stake *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addReward(addr, balance, stake, Flips)
}

func (c *blockStatsCollector) AddInvitationsReward(addr common.Address, balance *big.Int, stake *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addReward(addr, balance, stake, Invitations)
}

func (c *blockStatsCollector) AddFoundationPayout(addr common.Address, balance *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addReward(addr, balance, nil, FoundationPayouts)
}

func (c *blockStatsCollector) AddZeroWalletFund(addr common.Address, balance *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addReward(addr, balance, nil, ZeroWalletFund)
}

func (c *blockStatsCollector) addReward(addr common.Address, balance *big.Int, stake *big.Int, rewardType RewardType) {
	if (balance == nil || balance.Sign() == 0) && (stake == nil || stake.Sign() == 0) {
		return
	}
	c.initRewardStats()
	c.stats.RewardsStats.Rewards = append(c.stats.RewardsStats.Rewards, &RewardStats{
		Address: addr,
		Balance: balance,
		Stake:   stake,
		Type:    rewardType,
	})
}

func (c *blockStatsCollector) AddProposerReward(addr common.Address, balance *big.Int, stake *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addMiningReward(addr, balance, stake, proposerReward)
}

func (c *blockStatsCollector) AddFinalCommitteeReward(addr common.Address, balance *big.Int, stake *big.Int) {
	if !c.canCollect() {
		return
	}
	c.addMiningReward(addr, balance, stake, finalCommitteeReward)
	c.stats.FinalCommittee = append(c.stats.FinalCommittee, addr)
}

func (c *blockStatsCollector) addMiningReward(addr common.Address, balance *big.Int, stake *big.Int, rType string) {
	c.stats.MiningRewards = append(c.stats.MiningRewards, &db.Reward{
		Address: conversion.ConvertAddress(addr),
		Balance: blockchain.ConvertToFloat(balance),
		Stake:   blockchain.ConvertToFloat(stake),
		Type:    rType,
	})
}

func (c *blockStatsCollector) GetStats() *Stats {
	return c.stats
}

func (c *blockStatsCollector) CompleteCollecting() {
	c.stats = nil
	c.collect = false
}
