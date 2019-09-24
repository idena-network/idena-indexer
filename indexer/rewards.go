package indexer

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
)

var rewardTypes = map[stats.RewardType]string{
	stats.Validation:        "Validation",
	stats.Flips:             "Flips",
	stats.Invitations:       "Invitations",
	stats.FoundationPayouts: "FoundationPayouts",
	stats.ZeroWalletFund:    "ZeroWalletFund",
}

func (indexer *Indexer) detectEpochRewards(block *types.Block) *db.EpochRewards {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil
	}

	rewardsStats := indexer.blockStatsHolder().GetStats().RewardsStats
	if rewardsStats == nil {
		return nil
	}

	epochRewards := &db.EpochRewards{}
	epochRewards.BadAuthors = convertBadAuthors(rewardsStats.Authors.BadAuthors)
	epochRewards.GoodAuthors = convertGoodAuthors(rewardsStats.Authors.GoodAuthors)
	epochRewards.ValidationRewards, epochRewards.FundRewards = convertRewards(rewardsStats.Rewards)
	epochRewards.Total = convertTotalRewards(rewardsStats)

	return epochRewards
}

func convertBadAuthors(badAuthors map[common.Address]struct{}) []string {
	var res []string
	for address := range badAuthors {
		res = append(res, conversion.ConvertAddress(address))
	}
	return res
}

func convertGoodAuthors(goodAuthors map[common.Address]*types.ValidationResult) []*db.ValidationResult {
	var res []*db.ValidationResult
	for address, vr := range goodAuthors {
		res = append(res, &db.ValidationResult{
			Address:           conversion.ConvertAddress(address),
			StrongFlips:       vr.StrongFlips,
			WeakFlips:         vr.WeakFlips,
			SuccessfulInvites: vr.SuccessfulInvites,
		})
	}
	return res
}

func convertTotalRewards(rewardsStats *stats.RewardsStats) *db.TotalRewards {
	if rewardsStats == nil {
		return nil
	}
	return &db.TotalRewards{
		Total:             blockchain.ConvertToFloat(rewardsStats.Total),
		Validation:        blockchain.ConvertToFloat(rewardsStats.Validation),
		Flips:             blockchain.ConvertToFloat(rewardsStats.Flips),
		Invitations:       blockchain.ConvertToFloat(rewardsStats.Invitations),
		FoundationPayouts: blockchain.ConvertToFloat(rewardsStats.FoundationPayouts),
		ZeroWalletFund:    blockchain.ConvertToFloat(rewardsStats.ZeroWalletFund),
	}
}

func convertRewards(rewards []*stats.RewardStats) (validationRewards, fundRewards []*db.Reward) {
	for _, reward := range rewards {
		convertedReward := convertReward(reward)
		if reward.Type == stats.FoundationPayouts || reward.Type == stats.ZeroWalletFund {
			fundRewards = append(fundRewards, convertedReward)
		} else {
			validationRewards = append(validationRewards, convertedReward)
		}
	}
	return
}

func convertReward(reward *stats.RewardStats) *db.Reward {
	return &db.Reward{
		Address: conversion.ConvertAddress(reward.Address),
		Balance: blockchain.ConvertToFloat(reward.Balance),
		Stake:   blockchain.ConvertToFloat(reward.Stake),
		Type:    convertRewardType(reward.Type),
	}
}

func convertRewardType(rewardType stats.RewardType) string {
	if res, ok := rewardTypes[rewardType]; ok {
		return res
	}
	return fmt.Sprintf("Unknown reward type %d", rewardType)
}
