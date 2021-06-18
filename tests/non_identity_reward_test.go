package tests

import (
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_nonIdentityReward(t *testing.T) {

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
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	rewardedAddress := tests.GetRandAddr()
	appState.State.SetBalance(rewardedAddress, big.NewInt(5000))
	appState.Precommit()
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	var block *types2.Block

	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{})
	statsCollector.AddValidationReward(rewardedAddress, addr, 1, big.NewInt(1000), big.NewInt(200))
	appState.State.IncEpoch()
	appState.State.ClearFlips(addr)
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	epochIdentities, err := testCommon.GetEpochIdentities(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 2, len(epochIdentities))
	require.Equal(t, "0.000000000000001", epochIdentities[1].TotalValidationReward.String())
}
