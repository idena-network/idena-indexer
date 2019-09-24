package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/stats/collector"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"math/big"
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
	c.initStats()
	if c.stats.RewardsStats != nil {
		return
	}
	c.stats.RewardsStats = &RewardsStats{}
}

func (c *blockStatsCollector) SetValidation(validation *statsTypes.ValidationStats) {
	if !c.canCollect() {
		return
	}
	c.initStats()
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

func (c *blockStatsCollector) AddValidationReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Validation)
}

func (c *blockStatsCollector) AddFlipsReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Flips)
}

func (c *blockStatsCollector) AddInvitationsReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Invitations)
}

func (c *blockStatsCollector) AddFoundationPayout(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, FoundationPayouts)
}

func (c *blockStatsCollector) AddZeroWalletFund(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, ZeroWalletFund)
}

func (c *blockStatsCollector) addReward(addr common.Address, balance *big.Int, stake *big.Int, rewardType RewardType) {
	if !c.canCollect() {
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

func (c *blockStatsCollector) GetStats() *Stats {
	return c.stats
}

func (c *blockStatsCollector) CompleteCollecting() {
	c.stats = nil
	c.collect = false
}
