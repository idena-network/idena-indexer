package indexer

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/db"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_delegateeEpochRewardsWrapper(t *testing.T) {
	value := newDelegateeEpochRewardsWrapper(10)
	value.append(db.DelegateeEpochRewards{
		Address:      common.Address{0x1},
		TotalRewards: make([]db.DelegationEpochReward, 11),
	})
	value.append(db.DelegateeEpochRewards{
		Address:      common.Address{0x2},
		TotalRewards: make([]db.DelegationEpochReward, 12),
	})
	value.append(db.DelegateeEpochRewards{
		Address:      common.Address{0x3},
		TotalRewards: make([]db.DelegationEpochReward, 13),
	})
	value.incPenalizedDelegators(common.Address{0x1})
	value.incPenalizedDelegators(common.Address{0x3})
	value.incPenalizedDelegators(common.Address{0x1})
	value.incPenalizedDelegators(common.Address{0x4})

	require.Len(t, value.indexesByDelegatee, 4)
	require.Len(t, value.rewards, 4)

	require.Equal(t, uint32(2), value.rewards[0].PenalizedDelegators)
	require.Equal(t, common.Address{0x1}, value.rewards[0].Address)
	require.Len(t, value.rewards[0].TotalRewards, 11)

	require.Equal(t, uint32(0), value.rewards[1].PenalizedDelegators)
	require.Equal(t, common.Address{0x2}, value.rewards[1].Address)
	require.Len(t, value.rewards[1].TotalRewards, 12)

	require.Equal(t, uint32(1), value.rewards[2].PenalizedDelegators)
	require.Equal(t, common.Address{0x3}, value.rewards[2].Address)
	require.Len(t, value.rewards[2].TotalRewards, 13)

	require.Equal(t, uint32(1), value.rewards[3].PenalizedDelegators)
	require.Equal(t, common.Address{0x4}, value.rewards[3].Address)
	require.Len(t, value.rewards[3].TotalRewards, 0)

	value.incPenalizedDelegators(common.Address{0x4})
	require.Equal(t, uint32(2), value.rewards[3].PenalizedDelegators)
	require.Equal(t, common.Address{0x4}, value.rewards[3].Address)
	require.Len(t, value.rewards[3].TotalRewards, 0)
}
