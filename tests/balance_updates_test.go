package tests

import (
	"github.com/idena-network/idena-go/blockchain"
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/db"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func Test_committeeRewardZeroBlocksCount(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	addr := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	applyBlockWithHeight(bus, 2, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(updates))
	require.Equal(t, 1, *updates[0].BlocksCount)
	require.Equal(t, dna(4), blockchain.ConvertToInt(updates[0].BalanceNew))
	require.Equal(t, dna(1), blockchain.ConvertToInt(updates[0].StakeNew))
	committeeUpdates, err := testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 0, len(committeeUpdates))

	// When
	statsCollector.EnableCollecting()
	applyBlockWithHeight(bus, 3, appState)
	statsCollector.CompleteCollecting()
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	applyBlockWithHeight(bus, 4, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(updates))
	require.Equal(t, 2, *updates[0].BlocksCount)
	require.Equal(t, 2, updates[0].BlockHeight)
	require.Equal(t, 4, *updates[0].LastBlockHeight)
	require.Equal(t, dna(8), blockchain.ConvertToInt(updates[0].BalanceNew))
	require.Equal(t, dna(2), blockchain.ConvertToInt(updates[0].StakeNew))

	// When
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(5), dna(2), appState)
	applyBlockWithHeight(bus, 5, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(updates))
	require.Equal(t, 3, *updates[0].BlocksCount)
	require.Equal(t, 2, updates[0].BlockHeight)
	require.Equal(t, 5, *updates[0].LastBlockHeight)
	require.Equal(t, dna(0), blockchain.ConvertToInt(updates[0].BalanceOld))
	require.Equal(t, dna(0), blockchain.ConvertToInt(updates[0].StakeOld))
	require.Equal(t, dna(13), blockchain.ConvertToInt(updates[0].BalanceNew))
	require.Equal(t, dna(4), blockchain.ConvertToInt(updates[0].StakeNew))
}

func Test_changeCommitteeRewardBlocksCount(t *testing.T) {
	dbConnector, indxr, listener, _, bus := testCommon.InitIndexer(true, 3, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	addr := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	for i := 0; i < 5; i++ {
		statsCollector.EnableCollecting()
		addCommitteeReward(statsCollector, addr, dna(1), nil, appState)
		height := uint64(2 + i)
		applyBlockWithHeight(bus, height, appState)
		statsCollector.CompleteCollecting()
	}

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(updates))
	require.Equal(t, 2, updates[0].BlockHeight)
	require.Equal(t, 6, *updates[0].LastBlockHeight)
	require.Equal(t, 5, *updates[0].BlocksCount)
	committeeUpdates, err := testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 3, len(committeeUpdates))

	// When
	dbConnector.Close()
	indxr.Destroy()
	ctx := testCommon.InitIndexer2(testCommon.Options{
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
		AppState:          appState,
	})
	defer ctx.Listener.Destroy()
	dbConnector, indxr, listener, bus = ctx.DbConnector, ctx.Indexer, ctx.Listener, ctx.EventBus
	appState = listener.NodeCtx().AppState

	statsCollector = listener.StatsCollector()
	statsCollector.EnableCollecting()
	applyBlockWithHeight(bus, 7, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(updates))
	require.Equal(t, 5, *updates[0].BlocksCount)
	require.Equal(t, 2, updates[0].BlockHeight)
	require.Equal(t, 6, *updates[0].LastBlockHeight)
	require.Equal(t, dna(5), blockchain.ConvertToInt(updates[0].BalanceNew))
	require.Equal(t, dna(0), blockchain.ConvertToInt(updates[0].StakeNew))
	committeeUpdates, err = testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 0, len(committeeUpdates))
}

func Test_complexCommitteeRewardBalanceUpdates3blocks(t *testing.T) {
	dbConnector, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 3, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetState(addr2, state.Newbie)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 2, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	applyBlockWithHeight(bus, 3, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	addCommitteeReward(statsCollector, addr2, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 4, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addEpochReward(statsCollector, addr2, dna(100), dna(50), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 5, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 6, appState)
	statsCollector.CompleteCollecting()

	// New committee reward share
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 7, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 8, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 9, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addEpochReward(statsCollector, addr1, dna(100), dna(50), appState)
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 10, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 11, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 12, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	appState.State.IncEpoch()
	applyBlockWithHeight(bus, 13, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 14, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 11, len(updates))
	require.Equal(t, 1, *updates[8].BlocksCount)
	require.Equal(t, 12, updates[8].BlockHeight)
	require.Equal(t, 2, *updates[10].BlocksCount)
	require.Equal(t, 13, updates[10].BlockHeight)
	committeeUpdates, err := testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 3, len(committeeUpdates))

	summaries, err := testCommon.GetBalanceUpdateSummaries(dbConnector)
	require.Nil(t, err)
	require.Len(t, summaries, 2)
	for _, summary := range summaries {
		require.True(t, summary.Address == addr1.Hex() || summary.Address == addr2.Hex())
		if summary.Address == addr1.Hex() {
			require.Equal(t, dna(164), blockchain.ConvertToInt(summary.BalanceIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.BalanceOut))
			require.Equal(t, dna(82), blockchain.ConvertToInt(summary.StakeIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.StakeOut))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyOut))
		}
		if summary.Address == addr2.Hex() {
			require.Equal(t, dna(102), blockchain.ConvertToInt(summary.BalanceIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.BalanceOut))
			require.Equal(t, dna(51), blockchain.ConvertToInt(summary.StakeIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.StakeOut))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyOut))
		}
	}
	summariesChanges, err := testCommon.GetBalanceUpdateSummariesChanges(dbConnector)
	require.Nil(t, err)
	require.Len(t, summariesChanges, 5)

	miningRewardSummaries, err := testCommon.GetMiningRewardSummaries(dbConnector)
	require.Nil(t, err)
	require.Len(t, miningRewardSummaries, 3)
	for _, summary := range miningRewardSummaries {
		require.True(t, summary.Address == addr1.Hex() || summary.Address == addr2.Hex())
		if summary.Address == addr1.Hex() {
			require.True(t, summary.Epoch == 0 || summary.Epoch == 1)
			if summary.Epoch == 0 {
				require.Equal(t, dna(93), blockchain.ConvertToInt(summary.Amount))
			} else {
				require.Equal(t, dna(3), blockchain.ConvertToInt(summary.Amount))
			}
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.Burnt))
		}
		if summary.Address == addr2.Hex() {
			require.Equal(t, dna(3), blockchain.ConvertToInt(summary.Amount))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.Burnt))
		}
	}

	miningRewardSummariesChanges, err := testCommon.GetMiningRewardSummariesChanges(dbConnector)
	require.Nil(t, err)
	require.Len(t, miningRewardSummariesChanges, 5)

	// When
	err = dbAccessor.ResetTo(12)

	// Then
	require.Nil(t, err)
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 9, len(updates))
	require.Equal(t, 1, *updates[8].BlocksCount)
	committeeUpdates, err = testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(committeeUpdates))

	summaries, err = testCommon.GetBalanceUpdateSummaries(dbConnector)
	require.Nil(t, err)
	require.Len(t, summaries, 2)
	for _, summary := range summaries {
		require.True(t, summary.Address == addr1.Hex() || summary.Address == addr2.Hex())
		if summary.Address == addr1.Hex() {
			require.Equal(t, dna(150), blockchain.ConvertToInt(summary.BalanceIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.BalanceOut))
			require.Equal(t, dna(75), blockchain.ConvertToInt(summary.StakeIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.StakeOut))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyOut))
		}
		if summary.Address == addr2.Hex() {
			require.Equal(t, dna(102), blockchain.ConvertToInt(summary.BalanceIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.BalanceOut))
			require.Equal(t, dna(51), blockchain.ConvertToInt(summary.StakeIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.StakeOut))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyIn))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.PenaltyOut))
		}
	}
	summariesChanges, err = testCommon.GetBalanceUpdateSummariesChanges(dbConnector)
	require.Nil(t, err)
	require.Len(t, summariesChanges, 2)

	miningRewardSummaries, err = testCommon.GetMiningRewardSummaries(dbConnector)
	require.Nil(t, err)
	require.Len(t, miningRewardSummaries, 2)
	for _, summary := range miningRewardSummaries {
		require.True(t, summary.Address == addr1.Hex() || summary.Address == addr2.Hex())
		if summary.Address == addr1.Hex() {
			require.Equal(t, dna(75), blockchain.ConvertToInt(summary.Amount))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.Burnt))
			require.Equal(t, uint16(0), summary.Epoch)
		}
		if summary.Address == addr2.Hex() {
			require.Equal(t, dna(3), blockchain.ConvertToInt(summary.Amount))
			require.Equal(t, dna(0), blockchain.ConvertToInt(summary.Burnt))
		}
	}

	miningRewardSummariesChanges, err = testCommon.GetMiningRewardSummariesChanges(dbConnector)
	require.Nil(t, err)
	require.Len(t, miningRewardSummariesChanges, 2)
}

func Test_complexCommitteeRewardBalanceUpdates6blocks(t *testing.T) {
	dbConnector, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 6, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetState(addr2, state.Newbie)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 2, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	applyBlockWithHeight(bus, 3, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	addCommitteeReward(statsCollector, addr2, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 4, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addEpochReward(statsCollector, addr2, dna(100), dna(50), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 5, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 6, appState)
	statsCollector.CompleteCollecting()

	// New committee reward share
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 7, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 8, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 9, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addEpochReward(statsCollector, addr1, dna(100), dna(50), appState)
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 10, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 11, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 12, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addProposerReward(statsCollector, addr1, dna(10), dna(5), appState)
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 13, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 14, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 11, len(updates))
	require.Equal(t, 1, *updates[8].BlocksCount)
	require.Equal(t, 12, updates[8].BlockHeight)
	require.Equal(t, 2, *updates[10].BlocksCount)
	require.Equal(t, 13, updates[10].BlockHeight)
	committeeUpdates, err := testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 6, len(committeeUpdates))

	// When
	err = dbAccessor.ResetTo(10)

	// Then
	require.Nil(t, err)
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 7, len(updates))
	require.Equal(t, 1, *updates[6].BlocksCount)
	require.Equal(t, 10, updates[6].BlockHeight)
	require.Equal(t, 10, *updates[6].LastBlockHeight)
	require.Equal(t, dna(134), blockchain.ConvertToInt(updates[6].BalanceOld))
	require.Equal(t, dna(67), blockchain.ConvertToInt(updates[6].StakeOld))
	require.Equal(t, dna(136), blockchain.ConvertToInt(updates[6].BalanceNew))
	require.Equal(t, dna(68), blockchain.ConvertToInt(updates[6].StakeNew))
	committeeUpdates, err = testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 2, len(committeeUpdates))
}

func Test_reset(t *testing.T) {
	dbConnector, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 6, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetState(addr2, state.Human)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 2, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr2, dna(2000), dna(1000), appState)
	applyBlockWithHeight(bus, 3, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 4, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 5, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	addProposerReward(statsCollector, addr1, dna(200), dna(100), appState)
	applyBlockWithHeight(bus, 6, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr1, dna(2), dna(1), appState)
	applyBlockWithHeight(bus, 7, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 4, len(updates))
	committeeUpdates, err := testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 6, len(committeeUpdates))

	// When
	err = dbAccessor.ResetTo(3)

	// Then
	require.Nil(t, err)
	updates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 2, len(updates))
	require.Equal(t, 1, *updates[0].BlocksCount)
	require.Equal(t, 2, updates[0].BlockHeight)
	require.Equal(t, 2, *updates[0].LastBlockHeight)
	require.Equal(t, dna(0), blockchain.ConvertToInt(updates[0].BalanceOld))
	require.Equal(t, dna(0), blockchain.ConvertToInt(updates[0].StakeOld))
	require.Equal(t, dna(2), blockchain.ConvertToInt(updates[0].BalanceNew))
	require.Equal(t, dna(1), blockchain.ConvertToInt(updates[0].StakeNew))
	require.Equal(t, 1, *updates[1].BlocksCount)
	require.Equal(t, 3, updates[1].BlockHeight)
	require.Equal(t, 3, *updates[1].LastBlockHeight)
	committeeUpdates, err = testCommon.GetCommitteeRewardBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 2, len(committeeUpdates))
}

func Test_penalty(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 3, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	addr := tests.GetRandAddr()
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	// When
	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	applyBlockWithHeight(bus, 2, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	setPenalty(statsCollector, addr, 777, appState)
	applyBlockWithHeight(bus, 3, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	applyBlockWithHeight(bus, 4, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	addCommitteeReward(statsCollector, addr, dna(4), dna(1), appState)
	applyBlockWithHeight(bus, 5, appState)
	statsCollector.CompleteCollecting()

	statsCollector.EnableCollecting()
	setEpochPenaltyReset(statsCollector, addr, appState)
	applyBlockWithHeight(bus, 6, appState)
	statsCollector.CompleteCollecting()

	// Then
	updates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 4, len(updates))
	require.Equal(t, db.PenaltyReason, updates[1].Reason)
	require.Nil(t, updates[1].PenaltyOld)
	require.Nil(t, updates[1].PenaltySecondsOld)
	require.Nil(t, updates[1].PenaltyNew)
	require.Equal(t, uint16(777), *updates[1].PenaltySecondsNew)
	require.Equal(t, db.EpochPenaltyResetReason, updates[3].Reason)
	require.Nil(t, updates[3].PenaltyOld)
	require.Equal(t, uint16(777), *updates[3].PenaltySecondsOld)
	require.Nil(t, updates[3].PenaltyNew)
	require.Nil(t, updates[3].PenaltySecondsNew)
}

func dna(amount int) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(amount)), common.DnaBase)
}

func addProposerReward(
	collector collector.StatsCollector,
	addr common.Address,
	balance *big.Int,
	stake *big.Int,
	appState *appstate.AppState,
) {
	collector.BeginProposerRewardBalanceUpdate(addr, addr, nil, appState)
	updateBalanceAndComplete(collector, addr, balance, stake, appState.State.GetPenaltySeconds(addr), appState)
}

func addCommitteeReward(
	collector collector.StatsCollector,
	addr common.Address,
	balance *big.Int,
	stake *big.Int,
	appState *appstate.AppState,
) {
	collector.BeginCommitteeRewardBalanceUpdate(addr, addr, nil, appState)
	updateBalanceAndComplete(collector, addr, balance, stake, appState.State.GetPenaltySeconds(addr), appState)
}

func setPenalty(
	collector collector.StatsCollector,
	addr common.Address,
	penaltySeconds uint16,
	appState *appstate.AppState,
) {
	collector.BeginPenaltyBalanceUpdate(addr, appState)
	updateBalanceAndComplete(collector, addr, appState.State.GetBalance(addr), appState.State.GetStakeBalance(addr), penaltySeconds, appState)
}

func setEpochPenaltyReset(
	collector collector.StatsCollector,
	addr common.Address,
	appState *appstate.AppState,
) {
	collector.BeginEpochPenaltyResetBalanceUpdate(addr, appState)
	updateBalanceAndComplete(collector, addr, appState.State.GetBalance(addr), appState.State.GetStakeBalance(addr), 0, appState)
}

func addEpochReward(
	collector collector.StatsCollector,
	addr common.Address,
	balance *big.Int,
	stake *big.Int,
	appState *appstate.AppState,
) {
	collector.BeginEpochRewardBalanceUpdate(addr, addr, appState)
	updateBalanceAndComplete(collector, addr, balance, stake, 0, appState)
}

func updateBalanceAndComplete(
	collector collector.StatsCollector,
	addr common.Address,
	balance *big.Int,
	stake *big.Int,
	penaltySeconds uint16,
	appState *appstate.AppState,
) {
	if balance != nil {
		appState.State.AddBalance(addr, balance)
	}
	if stake != nil {
		appState.State.AddStake(addr, stake)
	}
	appState.State.SetPenaltySeconds(addr, penaltySeconds)
	collector.CompleteBalanceUpdate(appState)
}

func applyBlockWithHeight(bus eventbus.Bus, height uint64, appState *appstate.AppState) error {
	block := buildBlock(height)
	appState.Precommit()
	if err := appState.Commit(block); err != nil {
		return err
	}
	bus.Publish(&events.NewBlockEvent{
		Block: block,
	})
	return nil
}

func applyBlock(bus eventbus.Bus, block *types2.Block, appState *appstate.AppState) error {
	appState.Precommit()
	if err := appState.Commit(block); err != nil {
		return err
	}
	appState.ValidatorsCache.Load()
	bus.Publish(&events.NewBlockEvent{
		Block: block,
	})
	return nil
}

func buildBlock(height uint64) *types2.Block {
	return &types2.Block{
		Header: &types2.Header{
			ProposedHeader: &types2.ProposedHeader{
				Height: height,
				Time:   time.Now().UTC().Unix(),
			},
		},
		Body: &types2.Body{
			Transactions: []*types2.Transaction{},
		},
	}
}
