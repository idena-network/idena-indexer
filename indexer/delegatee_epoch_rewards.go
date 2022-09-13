package indexer

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
)

func convertDelegateesEpochRewards(delegateesEpochRewards map[common.Address]*stats.DelegateeEpochRewards) *delegateeEpochRewardsWrapper {
	res := newDelegateeEpochRewardsWrapper(len(delegateesEpochRewards))
	if len(delegateesEpochRewards) == 0 {
		return res
	}
	for delegatee, incomingDelegateeEpochRewards := range delegateesEpochRewards {
		delegateeEpochReward := db.DelegateeEpochRewards{
			Address: delegatee,
		}
		if len(incomingDelegateeEpochRewards.TotalRewards) > 0 {
			delegateeEpochReward.TotalRewards = make([]db.DelegationEpochReward, 0, len(incomingDelegateeEpochRewards.TotalRewards))
			for rewardType, epochReward := range incomingDelegateeEpochRewards.TotalRewards {
				delegateeEpochReward.TotalRewards = append(delegateeEpochReward.TotalRewards, db.DelegationEpochReward{
					Balance: epochReward.Balance,
					Type:    byte(rewardType),
				})
			}
		}
		if len(incomingDelegateeEpochRewards.DelegatorsEpochRewards) > 0 {
			delegateeEpochReward.DelegatorRewards = make([]db.DelegatorEpochReward, 0, len(incomingDelegateeEpochRewards.DelegatorsEpochRewards))
			for delegator, incomingDelegatorEpochRewards := range incomingDelegateeEpochRewards.DelegatorsEpochRewards {
				delegatorEpochRewards := db.DelegatorEpochReward{
					Address: delegator,
				}
				if len(incomingDelegatorEpochRewards.EpochRewards) > 0 {
					delegatorEpochRewards.TotalRewards = make([]db.DelegationEpochReward, 0, len(incomingDelegatorEpochRewards.EpochRewards))
					for rewardType, epochReward := range incomingDelegatorEpochRewards.EpochRewards {
						delegatorEpochRewards.TotalRewards = append(delegatorEpochRewards.TotalRewards, db.DelegationEpochReward{
							Balance: epochReward.Balance,
							Type:    byte(rewardType),
						})
					}
				}
				delegateeEpochReward.DelegatorRewards = append(delegateeEpochReward.DelegatorRewards, delegatorEpochRewards)
			}
		}
		res.append(delegateeEpochReward)
	}
	return res
}
