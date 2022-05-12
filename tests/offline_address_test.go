package tests

import (
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_offlineAddress(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	var height uint64
	appState := listener.NodeCtx().AppState
	offlineAddress := tests.GetRandAddr()
	appState.State.SetState(offlineAddress, state.Verified)
	appState.Precommit(true)
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.OfflineAddr = &offlineAddress
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	blocks, err := testCommon.GetBlocks(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 3, len(blocks))
	require.Nil(t, blocks[0].OfflineAddress)
	require.Nil(t, blocks[1].OfflineAddress)
	require.Equal(t, offlineAddress.Hex(), *blocks[2].OfflineAddress)
}
