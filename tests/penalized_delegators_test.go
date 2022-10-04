package tests

import (
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_penalizedDelegators(t *testing.T) {

	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:           true,
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
	})
	listener := ctx.Listener
	bus := ctx.EventBus
	dbConnector := ctx.DbConnector
	defer listener.Destroy()

	statsCollector := listener.StatsCollector()

	var height uint64
	pool1, pool2, rewardedAddress, penalizedAddress1, penalizedAddress2, penalizedAddress3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetDelegatee(rewardedAddress, pool1)
	appState.State.SetDelegatee(penalizedAddress1, pool1)
	appState.State.SetDelegatee(penalizedAddress2, pool1)
	appState.State.SetDelegatee(penalizedAddress3, pool2)
	appState.Precommit()
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	var block *types2.Block

	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{})
	statsCollector.AddValidationReward(pool1, rewardedAddress, 1, big.NewInt(1000), big.NewInt(200))
	statsCollector.SetValidationResults(map[common.ShardId]*types2.ValidationResults{
		1: {
			BadAuthors: map[common.Address]types2.BadAuthorReason{
				penalizedAddress1: types2.WrongWordsBadAuthor,
				penalizedAddress2: types2.WrongWordsBadAuthor,
				penalizedAddress3: types2.WrongWordsBadAuthor,
			},
		},
	})
	statsCollector.SetTotalReportsReward(new(big.Int), new(big.Int))
	appState.State.IncEpoch()
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	epochIdentities, err := testCommon.GetEpochIdentities(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 4, len(epochIdentities))

	delegateeTotalValidationRewards, err := testCommon.GetDelegateeTotalValidationRewards(dbConnector)
	require.Nil(t, err)
	require.Len(t, delegateeTotalValidationRewards, 2)
	for _, delegateeTotalValidationReward := range delegateeTotalValidationRewards {
		if delegateeTotalValidationReward.DelegateeAddress == pool1.Hex() {
			require.Equal(t, "0.000000000000001", delegateeTotalValidationReward.TotalBalance.String())
			require.Equal(t, 2, delegateeTotalValidationReward.PenalizedDelegators)
		} else {
			require.Equal(t, pool2.Hex(), delegateeTotalValidationReward.DelegateeAddress)
			require.Zero(t, delegateeTotalValidationReward.TotalBalance.Sign())
			require.Equal(t, 1, delegateeTotalValidationReward.PenalizedDelegators)
		}
	}
}
