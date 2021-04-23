package tests

import (
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_upgradeVotingHistory(t *testing.T) {
	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:           true,
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
	})

	statsCollector := ctx.Listener.StatsCollector()
	appState := ctx.Listener.NodeCtx().AppState

	var height uint64
	var block *types2.Block

	height = 1
	appState.Precommit()
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	for k, _ := range config.ConsensusVersions {
		delete(config.ConsensusVersions, k)
	}

	now := time.Now().UTC()
	config.ConsensusVersions[1] = &config.ConsensusConf{
		StartActivationDate: now.Add(time.Second).Unix(),
		EndActivationDate:   now.Add(time.Second * 5).Unix(),
	}
	config.ConsensusVersions[2] = &config.ConsensusConf{
		StartActivationDate: now.Add(time.Second * 3).Unix(),
		EndActivationDate:   now.Add(time.Second * 7).Unix(),
	}
	config.ConsensusVersions[4] = &config.ConsensusConf{
		StartActivationDate: now.Add(time.Second * 0).Unix(),
		EndActivationDate:   now.Add(time.Second * 10).Unix(),
	}

	height = 2
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	height = 3
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err := testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, history)

	ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 10}, {2, 20}, {3, 30}})
	height = 4
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err = testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, history)

	ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 11}, {2, 21}, {3, 31}})
	time.Sleep(time.Second * 2)
	height = 5
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err = testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, history, 1)

	ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 12}, {2, 22}, {3, 32}})
	time.Sleep(time.Second * 2)
	height = 6
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err = testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, history, 3)

	ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 13}, {2, 23}, {3, 33}})
	time.Sleep(time.Second * 2)
	height = 7
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err = testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, history, 4)

	ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 14}, {2, 24}, {3, 34}})
	time.Sleep(time.Second * 2)
	height = 8
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	history, err = testCommon.GetUpgradeVotingHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, history, 4)
}

func Test_upgradeVotingShortHistory(t *testing.T) {
	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:                           true,
		Schema:                            testCommon.PostgresSchema,
		ScriptsPathPrefix:                 "..",
		UpgradeVotingShortHistoryItems:    testCommon.Pint(6),
		UpgradeVotingShortHistoryMinShift: testCommon.Pint(0),
	})

	statsCollector := ctx.Listener.StatsCollector()
	appState := ctx.Listener.NodeCtx().AppState

	var height uint64
	var block *types2.Block

	height = 1
	appState.Precommit()
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	for k, _ := range config.ConsensusVersions {
		delete(config.ConsensusVersions, k)
	}

	now := time.Now().UTC()
	config.ConsensusVersions[1] = &config.ConsensusConf{
		StartActivationDate: now.Add(time.Second).Unix(),
		EndActivationDate:   now.Add(time.Hour * 5).Unix(),
	}

	height = 2
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	time.Sleep(time.Second * 2)

	for height = 3; height < 14; height++ {
		ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 1000 + height}})
		statsCollector.EnableCollecting()
		block = buildBlock(height)
		require.Nil(t, applyBlock(ctx.EventBus, block, appState))
		statsCollector.CompleteCollecting()
	}

	time.Sleep(time.Second)

	history, err := testCommon.GetUpgradeVotingShortHistory(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, history, 6)
	require.Equal(t, 13, history[5].BlockHeight)
	require.Equal(t, 1, history[5].Upgrade)
	require.Equal(t, 1013, history[5].Votes)
}

//func Test_upgradeVotingShortHistory_longVoting(t *testing.T) {
//	ctx := testCommon.InitIndexer2(testCommon.Options{
//		ClearDb:                           true,
//		Schema:                            testCommon.PostgresSchema,
//		ScriptsPathPrefix:                 "..",
//		UpgradeVotingShortHistoryItems:    testCommon.Pint(401),
//		UpgradeVotingShortHistoryMinShift: testCommon.Pint(5),
//	})
//
//	statsCollector := ctx.Listener.StatsCollector()
//	appState := ctx.Listener.NodeCtx().AppState
//
//	var height uint64
//	var block *types2.Block
//
//	height = 1
//	appState.Precommit()
//	require.Nil(t, appState.CommitAt(height))
//	require.Nil(t, appState.Initialize(height))
//
//	for k, _ := range config.ConsensusVersions {
//		delete(config.ConsensusVersions, k)
//	}
//
//	now := time.Now().UTC()
//	config.ConsensusVersions[1] = &config.ConsensusConf{
//		StartActivationDate: now.Add(time.Second).Unix(),
//		EndActivationDate:   now.Add(time.Hour * 5).Unix(),
//	}
//
//	height = 2
//	statsCollector.EnableCollecting()
//	block = buildBlock(height)
//	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
//	statsCollector.CompleteCollecting()
//
//	time.Sleep(time.Second * 2)
//
//	for height = 3; height < 32041; height++ {
//		ctx.UpgradesVotingHolder.Set([]*upgrade.Votes{{1, 1000000 + height}})
//		statsCollector.EnableCollecting()
//		block = buildBlock(height)
//		require.Nil(t, applyBlock(ctx.EventBus, block, appState))
//		statsCollector.CompleteCollecting()
//	}
//
//	time.Sleep(time.Second)
//
//	history, err := testCommon.GetUpgradeVotingShortHistory(ctx.DbConnector)
//	require.Nil(t, err)
//	require.Len(t, history, 401)
//	require.Equal(t, 32040, history[400].BlockHeight)
//	require.Equal(t, 1, history[400].Upgrade)
//	require.Equal(t, 1032040, history[400].Votes)
//}
