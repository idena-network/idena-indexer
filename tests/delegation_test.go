package tests

import (
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_delegation(t *testing.T) {
	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:           true,
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
	})

	statsCollector := ctx.Listener.StatsCollector()
	appState := ctx.Listener.NodeCtx().AppState

	appState.State.SetGlobalEpoch(5)

	key1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)
	appState.State.SetState(addr1, state.Zombie)
	appState.State.SetBirthday(addr1, 1)

	key2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)
	appState.State.SetState(addr2, state.Human)
	appState.State.SetBirthday(addr2, 2)

	key3, _ := crypto.GenerateKey()
	delegatee2 := crypto.PubkeyToAddress(key3.PublicKey)
	appState.State.SetState(delegatee2, state.Human)
	appState.State.SetBirthday(delegatee2, 3)

	var height uint64
	var block *types2.Block

	height = 1
	appState.Precommit()
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	height = 2
	delegator1 := tests.GetRandAddr()
	delegatee1 := tests.GetRandAddr()
	appState.State.ToggleDelegationAddress(delegator1, delegatee1)
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err := testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, poolsSummaries)

	poolSizes, err := testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, poolSizes)

	delegations, err := testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, delegations)

	height = 3
	appState.State.SetDelegatee(delegator1, delegatee1)
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 1, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 1)
	require.Equal(t, delegatee1.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 1)
	require.Equal(t, delegator1.Hex(), delegations[0].Delegator)
	require.Equal(t, delegatee1.Hex(), delegations[0].Delegatee)
	require.Nil(t, delegations[0].BirthEpoch)

	height = 4
	appState.State.ToggleDelegationAddress(addr1, delegatee2)
	appState.State.ToggleDelegationAddress(addr2, delegatee1)
	statsCollector.EnableCollecting()
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 1, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 1)
	require.Equal(t, delegatee1.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 1)
	require.Equal(t, delegator1.Hex(), delegations[0].Delegator)
	require.Equal(t, delegatee1.Hex(), delegations[0].Delegatee)
	require.Nil(t, delegations[0].BirthEpoch)

	height = 5
	appState.State.SetDelegatee(addr1, delegatee2)
	appState.State.SetDelegatee(addr2, delegatee1)
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate
	statsCollector.EnableCollecting()
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 2, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 2)
	require.Equal(t, delegatee2.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)
	require.Equal(t, delegatee1.Hex(), poolSizes[1].Address)
	require.Equal(t, 2, poolSizes[1].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 3)

	delegationsByDelegator := make(map[string]testCommon.Delegation, len(delegations))
	for _, delegation := range delegations {
		delegationsByDelegator[delegation.Delegator] = delegation
	}

	delegation := delegationsByDelegator[addr1.Hex()]
	require.Equal(t, addr1.Hex(), delegation.Delegator)
	require.Equal(t, delegatee2.Hex(), delegation.Delegatee)
	require.Equal(t, 1, *delegation.BirthEpoch)

	delegation = delegationsByDelegator[addr2.Hex()]
	require.Equal(t, addr2.Hex(), delegation.Delegator)
	require.Equal(t, delegatee1.Hex(), delegation.Delegatee)
	require.Equal(t, 2, *delegation.BirthEpoch)

	delegation = delegationsByDelegator[delegator1.Hex()]
	require.Equal(t, delegator1.Hex(), delegation.Delegator)
	require.Equal(t, delegatee1.Hex(), delegation.Delegatee)
	require.Nil(t, delegation.BirthEpoch)

	height = 6

	statsCollector.EnableCollecting()
	tx, _ := types2.SignTx(&types2.Transaction{}, key2)
	statsCollector.BeginApplyingTx(tx, appState)
	appState.State.SetState(addr2, state.Killed)
	statsCollector.CompleteApplyingTx(appState)

	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 2, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 2)
	require.Equal(t, delegatee2.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)
	require.Equal(t, delegatee1.Hex(), poolSizes[1].Address)
	require.Equal(t, 1, poolSizes[1].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 2)

	delegationsByDelegator = make(map[string]testCommon.Delegation, len(delegations))
	for _, delegation := range delegations {
		delegationsByDelegator[delegation.Delegator] = delegation
	}

	delegation = delegationsByDelegator[addr1.Hex()]
	require.Equal(t, addr1.Hex(), delegation.Delegator)
	require.Equal(t, delegatee2.Hex(), delegation.Delegatee)
	require.Equal(t, 1, *delegation.BirthEpoch)

	delegation = delegationsByDelegator[delegator1.Hex()]
	require.Equal(t, delegator1.Hex(), delegation.Delegator)
	require.Equal(t, delegatee1.Hex(), delegation.Delegatee)
	require.Nil(t, delegation.BirthEpoch)

	height = 7

	appState.State.SetState(addr1, state.Killed)
	appState.State.SetState(delegator1, state.Newbie)
	appState.State.SetBirthday(delegator1, 5)
	statsCollector.EnableCollecting()
	statsCollector.SetValidation(types.NewValidationStats())
	appState.State.IncEpoch()
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.IdentityUpdate | types2.ValidationFinished
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 1, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 1)
	require.Equal(t, delegatee1.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 1)

	require.Equal(t, delegator1.Hex(), delegations[0].Delegator)
	require.Equal(t, delegatee1.Hex(), delegations[0].Delegatee)
	require.Equal(t, 5, *delegations[0].BirthEpoch)

	err = ctx.DbAccessor.ResetTo(7)

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, poolsSummaries)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, poolSizes)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Empty(t, delegations)

	ctx = testCommon.InitIndexer2(testCommon.Options{
		ClearDb:           false,
		RestoreInitially:  true,
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
		TestBlockchain:    ctx.TestBlockchain,
		AppState:          appState,
	})
	appState = ctx.Listener.NodeCtx().AppState
	statsCollector = ctx.Listener.StatsCollector()

	height = 8
	block = buildBlock(height)
	statsCollector.EnableCollecting()
	require.Nil(t, applyBlock(ctx.EventBus, block, appState))
	statsCollector.CompleteCollecting()

	poolsSummaries, err = testCommon.GetPoolsSummaries(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolsSummaries, 1)
	require.Equal(t, 1, poolsSummaries[0].Count)

	poolSizes, err = testCommon.GetPoolSizes(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, poolSizes, 1)
	require.Equal(t, delegatee1.Hex(), poolSizes[0].Address)
	require.Equal(t, 1, poolSizes[0].Size)

	delegations, err = testCommon.GetDelegations(ctx.DbConnector)
	require.Nil(t, err)
	require.Len(t, delegations, 1)

	require.Equal(t, delegator1.Hex(), delegations[0].Delegator)
	require.Equal(t, delegatee1.Hex(), delegations[0].Delegatee)
	require.Equal(t, 5, *delegations[0].BirthEpoch)
}
