package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
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
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenalty(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginProposerRewardBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeNew)
}

func TestStatsCollector_TxBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()
	key, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(key.PublicKey)
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	recipient := tests.GetRandAddr()
	tx := tests.GetFullTx(1, 1, key, types.SendTx, nil, &recipient, nil)
	c.BeginTxBalanceUpdate(tx, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	appState.State.SetBalance(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.epochRewardBalanceUpdatesByAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, appState)
	appState.State.AddStake(addr, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 2, len(c.pending.epochRewardBalanceUpdatesByAddr))
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
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
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

func TestStatsCollector_contractBalanceUpdate(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()

	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	key, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(key.PublicKey)
	address2 := tests.GetRandAddr()
	address3 := tests.GetRandAddr()

	appState.State.SetBalance(sender, big.NewInt(1))
	appState.State.AddStake(address2, big.NewInt(2))
	appState.State.SetPenalty(address2, big.NewInt(3))

	tx := tests.GetFullTx(1, 1, key, types.SendTx, nil, nil, nil)
	c.BeginApplyingTx(tx, appState)

	c.AddContractBalanceUpdate(address2, appState.State.GetBalance, big.NewInt(200), appState)
	c.AddContractBalanceUpdate(address3, appState.State.GetBalance, big.NewInt(0), appState)

	c.AddContractBalanceUpdate(sender, appState.State.GetBalance, big.NewInt(11), appState)
	appState.State.SetBalance(sender, big.NewInt(11))

	c.AddContractBurntCoins(address3, func(address common.Address) *big.Int {
		return big.NewInt(400)
	})
	c.AddTxReceipt(&types.TxReceipt{Success: true}, appState)

	c.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(sender, big.NewInt(111))
	c.CompleteBalanceUpdate(appState)

	c.CompleteApplyingTx(appState)

	// When
	require.Equal(t, 0, big.NewInt(400).Cmp(c.stats.BurntCoins))
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr))
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr[address3]))
	require.Equal(t, db2.EmbeddedContractReason, c.stats.BurntCoinsByAddr[address3][0].Reason)
	require.Equal(t, tx.Hash().Hex(), c.stats.BurntCoinsByAddr[address3][0].TxHash)
	require.Equal(t, "0.0000000000000004", c.stats.BurntCoinsByAddr[address3][0].Amount.String())

	require.Equal(t, 2, c.stats.BalanceUpdateAddrs.Cardinality())
	require.True(t, c.stats.BalanceUpdateAddrs.Contains(sender))
	require.True(t, c.stats.BalanceUpdateAddrs.Contains(address2))

	findContractBalanceUpdate := func(address common.Address) *db2.BalanceUpdate {
		for _, bu := range c.stats.BalanceUpdates {
			if address == bu.Address && bu.Reason == db2.EmbeddedContractReason {
				return bu
			}
		}
		return nil
	}
	require.Equal(t, 3, len(c.stats.BalanceUpdates))
	bu := findContractBalanceUpdate(address2)
	require.Equal(t, big.NewInt(0), bu.BalanceOld)
	require.Equal(t, big.NewInt(200), bu.BalanceNew)
	require.Equal(t, big.NewInt(3), bu.PenaltyOld)
	require.Equal(t, big.NewInt(3), bu.PenaltyNew)
	require.Equal(t, big.NewInt(2), bu.StakeOld)
	require.Equal(t, big.NewInt(2), bu.StakeNew)
	require.Equal(t, tx.Hash(), *bu.TxHash)

	bu = findContractBalanceUpdate(sender)
	require.Equal(t, big.NewInt(1), bu.BalanceOld)
	require.Equal(t, big.NewInt(11), bu.BalanceNew)
	require.Nil(t, bu.PenaltyOld)
	require.Nil(t, bu.PenaltyNew)
	require.Equal(t, big.NewInt(0), bu.StakeOld)
	require.Equal(t, big.NewInt(0), bu.StakeNew)
	require.Equal(t, tx.Hash(), *bu.TxHash)

	bu = c.stats.BalanceUpdates[2]
	require.Equal(t, sender, bu.Address)
	require.Equal(t, db2.TxReason, bu.Reason)
	require.Equal(t, big.NewInt(11), bu.BalanceOld)
	require.Equal(t, big.NewInt(111), bu.BalanceNew)
	require.Nil(t, bu.PenaltyOld)
	require.Nil(t, bu.PenaltyNew)
	require.Equal(t, big.NewInt(0), bu.StakeOld)
	require.Equal(t, big.NewInt(0), bu.StakeNew)
	require.Equal(t, tx.Hash(), *bu.TxHash)
}
