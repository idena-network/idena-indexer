package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

func (indexer *Indexer) detectEpochRewards(block *types.Block) (*db.EpochRewards, map[common.Address]struct{}, *delegateeEpochRewardsWrapper) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil, newDelegateeEpochRewardsWrapper(0)
	}

	rewardsStats := indexer.statsHolder().GetStats().RewardsStats
	if rewardsStats == nil {
		return nil, nil, newDelegateeEpochRewardsWrapper(0)
	}

	epochRewards := &db.EpochRewards{}
	if rewardsStats.ValidationResults != nil {
		epochRewards.BadAuthors = convertBadAuthors(rewardsStats.ValidationResults)
	}
	var validationRewardsAddresses map[common.Address]struct{}
	epochRewards.ValidationRewards, epochRewards.FundRewards, validationRewardsAddresses = convertRewards(rewardsStats.Rewards)
	epochRewards.Total = convertTotalRewards(rewardsStats)
	epochRewards.AgesByAddress = rewardsStats.AgesByAddress
	epochRewards.StakedAmountsByAddress = rewardsStats.StakedAmountsByAddress
	epochRewards.FailedStakedAmountsByAddress = rewardsStats.FailedStakedAmountsByAddress
	epochRewards.RewardedFlipCids = rewardsStats.RewardedFlipCids
	epochRewards.RewardedExtraFlipCids = rewardsStats.RewardedExtraFlipCids
	epochRewards.RewardedInvitations = rewardsStats.RewardedInvites
	epochRewards.RewardedInvitees = rewardsStats.RewardedInvitees
	epochRewards.SavedInviteRewards = convertSavedInviteRewards(rewardsStats.SavedInviteRewardsCountByAddrAndType)
	epochRewards.ReportedFlipRewards = rewardsStats.ReportedFlipRewards

	return epochRewards, validationRewardsAddresses, convertDelegateesEpochRewards(rewardsStats.DelegateesEpochRewards)
}

func convertBadAuthors(validationResults map[common.ShardId]*types.ValidationResults) []*db.BadAuthor {
	var res []*db.BadAuthor
	for _, vr := range validationResults {
		for address, reason := range vr.BadAuthors {
			res = append(res, &db.BadAuthor{
				Address: conversion.ConvertAddress(address),
				Reason:  reason,
			})
		}
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
		Staking:           blockchain.ConvertToFloat(rewardsStats.Staking),
		Candidate:         blockchain.ConvertToFloat(rewardsStats.Candidate),
		Flips:             blockchain.ConvertToFloat(rewardsStats.Flips),
		FlipsExtra:        blockchain.ConvertToFloat(rewardsStats.FlipsExtra),
		Reports:           blockchain.ConvertToFloat(rewardsStats.Reports),
		Invitations:       blockchain.ConvertToFloat(rewardsStats.Invitations),
		FoundationPayouts: blockchain.ConvertToFloat(rewardsStats.FoundationPayouts),
		ZeroWalletFund:    blockchain.ConvertToFloat(rewardsStats.ZeroWalletFund),
		ValidationShare:   blockchain.ConvertToFloat(rewardsStats.ValidationShare),
		StakingShare:      blockchain.ConvertToFloat(rewardsStats.StakingShare),
		CandidateShare:    blockchain.ConvertToFloat(rewardsStats.CandidateShare),
		FlipsShare:        blockchain.ConvertToFloat(rewardsStats.FlipsShare),
		FlipsExtraShare:   blockchain.ConvertToFloat(rewardsStats.FlipsExtraShare),
		ReportsShare:      blockchain.ConvertToFloat(rewardsStats.ReportsShare),
		InvitationsShare:  blockchain.ConvertToFloat(rewardsStats.InvitationsShare),
	}
}

func convertRewards(rewards []*stats.RewardStats) (validationRewards, fundRewards []*db.Reward, validationRewardsAddresses map[common.Address]struct{}) {
	for _, reward := range rewards {
		convertedReward := convertReward(reward)
		if reward.Type == stats.FoundationPayouts || reward.Type == stats.ZeroWalletFund {
			fundRewards = append(fundRewards, convertedReward)
		} else {
			validationRewards = append(validationRewards, convertedReward)
			if validationRewardsAddresses == nil {
				validationRewardsAddresses = make(map[common.Address]struct{})
			}
			validationRewardsAddresses[reward.Address] = struct{}{}
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

func convertRewardType(rewardType stats.RewardType) byte {
	return byte(rewardType)
}

func convertSavedInviteRewards(savedInviteRewardsCountByAddrAndType map[common.Address]map[stats.RewardType]uint8) []*db.SavedInviteRewards {
	var res []*db.SavedInviteRewards
	for addr, addrCountByType := range savedInviteRewardsCountByAddrAndType {
		for rewardType, count := range addrCountByType {
			res = append(res, &db.SavedInviteRewards{
				Address: conversion.ConvertAddress(addr),
				Type:    byte(rewardType),
				Count:   count,
			})
		}
	}
	return res
}

func convertMiningRewards(rewards []*stats.MiningReward) []*db.MiningReward {
	if len(rewards) == 0 {
		return nil
	}
	res := make([]*db.MiningReward, len(rewards))
	for idx, reward := range rewards {
		res[idx] = &db.MiningReward{
			Address:     conversion.ConvertAddress(reward.Address),
			Balance:     blockchain.ConvertToFloat(reward.Balance),
			Stake:       blockchain.ConvertToFloat(reward.Stake),
			Proposer:    reward.Proposer,
			StakeWeight: math.Zero().Set(reward.StakeWeight),
		}
	}
	return res
}

type rewardsBounds struct {
	rewardBoundsByType map[byte]*db.RewardBounds
}

func (rb *rewardsBounds) getResult() []*db.RewardBounds {
	if len(rb.rewardBoundsByType) == 0 {
		return nil
	}
	res := make([]*db.RewardBounds, 0, len(rb.rewardBoundsByType))
	for _, rewardBounds := range rb.rewardBoundsByType {
		res = append(res, rewardBounds)
	}
	return res
}

func (rb *rewardsBounds) addIfBound(address common.Address, age uint64, amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	if rb.rewardBoundsByType == nil {
		rb.rewardBoundsByType = make(map[byte]*db.RewardBounds)
	}
	boundType := determineRewardBoundType(age)
	typeRewardBounds, ok := rb.rewardBoundsByType[boundType]
	if !ok {
		rb.rewardBoundsByType[boundType] = &db.RewardBounds{
			Type: boundType,
			Min: &db.RewardBound{
				Address: address,
				Amount:  new(big.Int).Set(amount),
			},
			Max: &db.RewardBound{
				Address: address,
				Amount:  new(big.Int).Set(amount),
			},
		}
		return
	}

	if typeRewardBounds.Min.Amount.Cmp(amount) == 1 {
		typeRewardBounds.Min = &db.RewardBound{
			Address: address,
			Amount:  new(big.Int).Set(amount),
		}
	}
	if typeRewardBounds.Max.Amount.Cmp(amount) == -1 {
		typeRewardBounds.Max = &db.RewardBound{
			Address: address,
			Amount:  new(big.Int).Set(amount),
		}
	}
}

func determineRewardBoundType(age uint64) byte {
	if age <= 5 {
		return byte(age)
	}
	return 6
}
