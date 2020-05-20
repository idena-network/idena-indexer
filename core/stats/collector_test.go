package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	db2 "github.com/idena-network/idena-indexer/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
	"math/big"
	"testing"
)

func TestStatsCollector_PenaltyBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	addr := tests.GetRandAddr()
	appState := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenalty(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.PenaltyReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Nil(t, c.stats.BalanceUpdates[0].PenaltyOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].PenaltyNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenalty(addr, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.PenaltyReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].PenaltyOld)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].PenaltyNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginEpochPenaltyResetBalanceUpdate(addr, appState)
	appState.State.ClearPenalty(addr)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochPenaltyResetReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].PenaltyOld)
	require.Nil(t, c.stats.BalanceUpdates[0].PenaltyNew)
}

func TestStatsCollector_ProposerRewardBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()
	addr := tests.GetRandAddr()
	appState := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginProposerRewardBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, appState)
	appState.State.SetBalance(addr, big.NewInt(12))
	appState.State.AddStake(addr, big.NewInt(2))
	appState.State.SetPenalty(addr, big.NewInt(3))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.ProposerRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Nil(t, c.stats.BalanceUpdates[0].PenaltyOld)
	require.Equal(t, big.NewInt(12), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].StakeNew)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[0].PenaltyNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, appState)
	appState.State.SetState(addr, state.Killed)
	c.CompleteBalanceUpdate(appState)
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeNew)
}

func TestStatsCollector_TxBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()
	key, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(key.PublicKey)
	appState := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	recipient := tests.GetRandAddr()
	tx := tests.GetFullTx(1, 1, key, types.SendTx, nil, &recipient, nil)
	c.BeginTxBalanceUpdate(tx, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	tx = tests.GetFullTx(1, 1, key, types.SendTx, nil, nil, nil)
	c.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(sender, big.NewInt(1))
	appState.State.AddStake(sender, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, sender, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.TxReason, c.stats.BalanceUpdates[0].Reason)
	require.Equal(t, tx.Hash(), *c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].StakeNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	recipient = tests.GetRandAddr()
	tx = tests.GetFullTx(1, 1, key, types.SendTx, nil, &recipient, nil)
	c.BeginTxBalanceUpdate(tx, appState)
	appState.State.AddStake(sender, big.NewInt(2))
	appState.State.SetBalance(recipient, big.NewInt(3))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 0, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 2, len(c.stats.BalanceUpdates))
	require.Equal(t, sender, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.TxReason, c.stats.BalanceUpdates[0].Reason)
	require.Equal(t, tx.Hash(), *c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(4), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, recipient, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.TxReason, c.stats.BalanceUpdates[1].Reason)
	require.Equal(t, tx.Hash(), *c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeNew)
}

func TestStatsCollector_EpochRewardBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()
	addr := tests.GetRandAddr()
	appState := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	appState.State.SetBalance(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 1, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 1, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	appState.State.AddStake(addr, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 1, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].StakeNew)

	// when
	addr2 := tests.GetRandAddr()
	c.BeginEpochRewardBalanceUpdate(addr2, appState)
	appState.State.SetBalance(addr2, big.NewInt(3))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pendingBalanceUpdates))
	require.Equal(t, 2, len(c.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 2, len(c.stats.BalanceUpdates))
	require.Equal(t, addr2, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeNew)
}

func TestStatsCollector_DustClearingBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()
	addr := tests.GetRandAddr()
	appState := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	appState.State.SetBalance(addr, big.NewInt(100))

	// When
	c.EnableCollecting()
	c.BeginDustClearingBalanceUpdate(addr, appState)
	appState.State.SetBalance(addr, big.NewInt(0))
	c.CompleteBalanceUpdate(appState)
	// Then
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.DustClearingReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(100), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(100), c.stats.BurntCoins)
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr))
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr[addr]))
	require.Equal(t, db2.DustClearingBurntCoins, c.stats.BurntCoinsByAddr[addr][0].Reason)
	require.Zero(t, decimal.New(1, -16).Cmp(c.stats.BurntCoinsByAddr[addr][0].Amount))
	require.Equal(t, "", c.stats.BurntCoinsByAddr[addr][0].TxHash)
}
