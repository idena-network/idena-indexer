package stats

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	types2 "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	db2 "github.com/idena-network/idena-indexer/db"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	db "github.com/tendermint/tm-db"
	"math/big"
	"testing"
)

func newStatsCollector() *statsCollector {
	res := &statsCollector{}
	res.consensusConf = blockchain.GetDefaultConsensusConfig()
	return res
}

func TestStatsCollector_PenaltyBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	addr := tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenaltySeconds(addr, 1)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.PenaltyReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Zero(t, c.stats.BalanceUpdates[0].PenaltySecondsOld)
	require.Equal(t, uint16(1), c.stats.BalanceUpdates[0].PenaltySecondsNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenaltySeconds(addr, 2)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.PenaltyReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, uint16(1), c.stats.BalanceUpdates[0].PenaltySecondsOld)
	require.Equal(t, uint16(2), c.stats.BalanceUpdates[0].PenaltySecondsNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginEpochPenaltyResetBalanceUpdate(addr, appState)
	appState.State.ClearPenalty(addr)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochPenaltyResetReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, uint16(2), c.stats.BalanceUpdates[0].PenaltySecondsOld)
	require.Nil(t, c.stats.BalanceUpdates[0].PenaltyNew)
}

func TestStatsCollector_ProposerRewardBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()
	addr := tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginProposerRewardBalanceUpdate(addr, addr, nil, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, addr, nil, appState)
	appState.State.SetBalance(addr, big.NewInt(12))
	appState.State.AddStake(addr, big.NewInt(2))
	appState.State.SetPenaltySeconds(addr, 3)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, addr, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.ProposerRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Nil(t, c.stats.BalanceUpdates[0].PenaltyOld)
	require.Equal(t, big.NewInt(12), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].StakeNew)
	require.Equal(t, uint16(3), c.stats.BalanceUpdates[0].PenaltySecondsNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, addr, nil, appState)
	appState.State.SetState(addr, state.Killed)
	c.CompleteBalanceUpdate(appState)
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeNew)
}

func TestStatsCollector_TxBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	c := newStatsCollector()
	c.EnableCollecting()
	addr := tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginEpochRewardBalanceUpdate(addr, addr, appState)
	appState.State.SetBalance(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 1, len(c.stats.BalanceUpdates))

	// when
	c.BeginEpochRewardBalanceUpdate(addr, addr, appState)
	appState.State.AddStake(addr, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
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
	c.BeginEpochRewardBalanceUpdate(addr2, addr2, appState)
	appState.State.SetBalance(addr2, big.NewInt(3))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 2, len(c.stats.BalanceUpdates))
	require.Equal(t, addr2, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeNew)
}

func TestStatsCollector_EpochRewardBalanceUpdateWithDelegatee(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()
	delegator, delegatee := tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	// when
	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddStake(delegator, big.NewInt(1))
	appState.State.AddBalance(delegatee, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 3, len(c.stats.BalanceUpdates))

	require.Equal(t, delegator, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[2].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeNew)

	delegator2, delegator3, delegatee2 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	// when
	c.BeginEpochRewardBalanceUpdate(delegator2, delegatee2, appState)
	appState.State.AddStake(delegator2, big.NewInt(3))
	appState.State.AddBalance(delegatee2, big.NewInt(4))
	c.CompleteBalanceUpdate(appState)
	c.BeginEpochRewardBalanceUpdate(delegator2, delegatee2, appState)
	appState.State.AddStake(delegator2, big.NewInt(5))
	appState.State.AddBalance(delegatee2, big.NewInt(6))
	c.CompleteBalanceUpdate(appState)
	c.BeginEpochRewardBalanceUpdate(delegator3, delegatee, appState)
	appState.State.AddStake(delegator3, big.NewInt(7))
	appState.State.AddBalance(delegatee, big.NewInt(8))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 3, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 3, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 8, len(c.stats.BalanceUpdates))

	require.Equal(t, delegator, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(2), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(1), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(10), c.stats.BalanceUpdates[2].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeNew)

	require.Equal(t, delegator2, c.stats.BalanceUpdates[3].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[3].Reason)
	require.Nil(t, c.stats.BalanceUpdates[3].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[3].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[3].StakeOld)
	require.Equal(t, big.NewInt(10), c.stats.BalanceUpdates[3].BalanceNew)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[3].StakeNew)

	require.Equal(t, delegator2, c.stats.BalanceUpdates[4].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[4].Reason)
	require.Nil(t, c.stats.BalanceUpdates[4].TxHash)
	require.Equal(t, big.NewInt(10), c.stats.BalanceUpdates[4].BalanceOld)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[4].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[4].BalanceNew)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[4].StakeNew)

	require.Equal(t, delegatee2, c.stats.BalanceUpdates[5].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[5].Reason)
	require.Nil(t, c.stats.BalanceUpdates[5].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[5].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[5].StakeOld)
	require.Equal(t, big.NewInt(10), c.stats.BalanceUpdates[5].BalanceNew)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[5].StakeNew)

	require.Equal(t, delegator3, c.stats.BalanceUpdates[6].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[6].Reason)
	require.Nil(t, c.stats.BalanceUpdates[6].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[6].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[6].StakeOld)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[6].BalanceNew)
	require.Equal(t, big.NewInt(7), c.stats.BalanceUpdates[6].StakeNew)

	require.Equal(t, delegator3, c.stats.BalanceUpdates[7].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[7].Reason)
	require.Nil(t, c.stats.BalanceUpdates[7].TxHash)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[7].BalanceOld)
	require.Equal(t, big.NewInt(7), c.stats.BalanceUpdates[7].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[7].BalanceNew)
	require.Equal(t, big.NewInt(7), c.stats.BalanceUpdates[7].StakeNew)
}

func TestStatsCollector_EpochRewardBalanceUpdateWithDelegatee2(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()
	delegator, delegatee := tests.GetRandAddr(), tests.GetRandAddr()

	// when
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddStake(delegator, big.NewInt(1))
	appState.State.AddBalance(delegatee, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddStake(delegatee, big.NewInt(3))
	appState.State.AddBalance(delegatee, big.NewInt(4))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddStake(delegator, big.NewInt(5))
	appState.State.AddBalance(delegatee, big.NewInt(6))
	c.CompleteBalanceUpdate(appState)

	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 4, len(c.stats.BalanceUpdates))

	require.Equal(t, delegator, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[2].BalanceNew)
	require.Zero(t, c.stats.BalanceUpdates[2].StakeNew.Sign())

	require.Equal(t, delegatee, c.stats.BalanceUpdates[3].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[3].Reason)
	require.Nil(t, c.stats.BalanceUpdates[3].TxHash)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[3].BalanceOld)
	require.Zero(t, c.stats.BalanceUpdates[3].StakeOld.Sign())
	require.Equal(t, big.NewInt(12), c.stats.BalanceUpdates[3].BalanceNew)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[3].StakeNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	appState, _ = appstate.NewAppState(db.NewMemDB(), eventbus.New())

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddStake(delegatee, big.NewInt(3))
	appState.State.AddBalance(delegatee, big.NewInt(4))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddStake(delegator, big.NewInt(1))
	appState.State.AddBalance(delegatee, big.NewInt(2))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddStake(delegator, big.NewInt(5))
	appState.State.AddBalance(delegatee, big.NewInt(6))
	c.CompleteBalanceUpdate(appState)

	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 4, len(c.stats.BalanceUpdates))

	require.Equal(t, delegatee, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(4), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(8), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceNew)
	require.Equal(t, big.NewInt(6), c.stats.BalanceUpdates[2].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[3].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[3].Reason)
	require.Nil(t, c.stats.BalanceUpdates[3].TxHash)
	require.Equal(t, big.NewInt(4), c.stats.BalanceUpdates[3].BalanceOld)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[3].StakeOld)
	require.Equal(t, big.NewInt(12), c.stats.BalanceUpdates[3].BalanceNew)
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[3].StakeNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	appState, _ = appstate.NewAppState(db.NewMemDB(), eventbus.New())

	delegator2 := tests.GetRandAddr()

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(2))
	appState.State.AddStake(delegatee, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(20))
	appState.State.AddStake(delegator, big.NewInt(10))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(200))
	appState.State.AddStake(delegatee, big.NewInt(100))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator2, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(2000))
	appState.State.AddStake(delegator2, big.NewInt(1000))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(20000))
	appState.State.AddStake(delegator, big.NewInt(10000))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(200000))
	appState.State.AddStake(delegatee, big.NewInt(100000))
	c.CompleteBalanceUpdate(appState)

	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 3, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 6, len(c.stats.BalanceUpdates))

	require.Equal(t, delegatee, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(200202), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(100101), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(20020), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(20020), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceNew)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[2].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[3].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[3].Reason)
	require.Nil(t, c.stats.BalanceUpdates[3].TxHash)
	require.Equal(t, big.NewInt(200202), c.stats.BalanceUpdates[3].BalanceOld)
	require.Equal(t, big.NewInt(100101), c.stats.BalanceUpdates[3].StakeOld)
	require.Equal(t, big.NewInt(222222), c.stats.BalanceUpdates[3].BalanceNew)
	require.Equal(t, big.NewInt(100101), c.stats.BalanceUpdates[3].StakeNew)

	require.Equal(t, delegator2, c.stats.BalanceUpdates[4].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[4].Reason)
	require.Nil(t, c.stats.BalanceUpdates[4].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[4].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[4].StakeOld)
	require.Equal(t, big.NewInt(2000), c.stats.BalanceUpdates[4].BalanceNew)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[4].StakeNew)

	require.Equal(t, delegator2, c.stats.BalanceUpdates[5].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[5].Reason)
	require.Nil(t, c.stats.BalanceUpdates[5].TxHash)
	require.Equal(t, big.NewInt(2000), c.stats.BalanceUpdates[5].BalanceOld)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[5].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[5].BalanceNew)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[5].StakeNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	appState, _ = appstate.NewAppState(db.NewMemDB(), eventbus.New())

	c.BeginEpochRewardBalanceUpdate(delegator2, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(2000))
	appState.State.AddStake(delegator2, big.NewInt(1000))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(20000))
	appState.State.AddStake(delegator, big.NewInt(10000))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(2))
	appState.State.AddStake(delegatee, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(200000))
	appState.State.AddStake(delegatee, big.NewInt(100000))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegator, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(20))
	appState.State.AddStake(delegator, big.NewInt(10))
	c.CompleteBalanceUpdate(appState)

	c.BeginEpochRewardBalanceUpdate(delegatee, delegatee, appState)
	appState.State.AddBalance(delegatee, big.NewInt(200))
	appState.State.AddStake(delegatee, big.NewInt(100))
	c.CompleteBalanceUpdate(appState)

	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 3, len(c.pending.balanceUpdatesByReasonAndAddr[db2.EpochRewardReason]))
	require.Equal(t, 2, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegatorEpochRewardReason]))
	require.Equal(t, 1, len(c.pending.balanceUpdatesByReasonAndAddr[db2.DelegateeEpochRewardReason]))
	require.Equal(t, 6, len(c.stats.BalanceUpdates))

	require.Equal(t, delegator2, c.stats.BalanceUpdates[0].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[0].Reason)
	require.Nil(t, c.stats.BalanceUpdates[0].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[0].StakeOld)
	require.Equal(t, big.NewInt(2000), c.stats.BalanceUpdates[0].BalanceNew)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[0].StakeNew)

	require.Equal(t, delegator2, c.stats.BalanceUpdates[1].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[1].Reason)
	require.Nil(t, c.stats.BalanceUpdates[1].TxHash)
	require.Equal(t, big.NewInt(2000), c.stats.BalanceUpdates[1].BalanceOld)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[1].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[1].BalanceNew)
	require.Equal(t, big.NewInt(1000), c.stats.BalanceUpdates[1].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[2].Address)
	require.Equal(t, db2.DelegateeEpochRewardReason, c.stats.BalanceUpdates[2].Reason)
	require.Nil(t, c.stats.BalanceUpdates[2].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[2].StakeOld)
	require.Equal(t, big.NewInt(22020), c.stats.BalanceUpdates[2].BalanceNew)
	require.Zero(t, c.stats.BalanceUpdates[2].StakeNew.Sign())

	require.Equal(t, delegator, c.stats.BalanceUpdates[3].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[3].Reason)
	require.Nil(t, c.stats.BalanceUpdates[3].TxHash)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[3].BalanceOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[3].StakeOld)
	require.Equal(t, big.NewInt(20020), c.stats.BalanceUpdates[3].BalanceNew)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[3].StakeNew)

	require.Equal(t, delegator, c.stats.BalanceUpdates[4].Address)
	require.Equal(t, db2.DelegatorEpochRewardReason, c.stats.BalanceUpdates[4].Reason)
	require.Nil(t, c.stats.BalanceUpdates[4].TxHash)
	require.Equal(t, big.NewInt(20020), c.stats.BalanceUpdates[4].BalanceOld)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[4].StakeOld)
	require.Equal(t, big.NewInt(0), c.stats.BalanceUpdates[4].BalanceNew)
	require.Equal(t, big.NewInt(10010), c.stats.BalanceUpdates[4].StakeNew)

	require.Equal(t, delegatee, c.stats.BalanceUpdates[5].Address)
	require.Equal(t, db2.EpochRewardReason, c.stats.BalanceUpdates[5].Reason)
	require.Nil(t, c.stats.BalanceUpdates[5].TxHash)
	require.Equal(t, big.NewInt(22020), c.stats.BalanceUpdates[5].BalanceOld)
	require.Zero(t, c.stats.BalanceUpdates[5].StakeOld.Sign())
	require.Equal(t, big.NewInt(222222), c.stats.BalanceUpdates[5].BalanceNew)
	require.Equal(t, big.NewInt(100101), c.stats.BalanceUpdates[5].StakeNew)
}

func TestStatsCollector_DustClearingBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
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
	c := newStatsCollector()
	c.EnableCollecting()

	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())

	key, _ := crypto.GenerateKey()
	sender := crypto.PubkeyToAddress(key.PublicKey)
	address2 := tests.GetRandAddr()
	address3 := tests.GetRandAddr()

	appState.State.SetBalance(sender, big.NewInt(1))
	appState.State.AddStake(address2, big.NewInt(2))
	appState.State.SetPenaltySeconds(address2, 3)

	tx := tests.GetFullTx(1, 1, key, types.SendTx, nil, nil, nil)
	c.BeginApplyingTx(tx, appState)

	balanceCache := make(map[common.Address]*big.Int)
	c.AddContractBalanceUpdate(nil, address2, appState.State.GetBalance, big.NewInt(200), appState, &balanceCache)
	c.AddContractBalanceUpdate(nil, address3, appState.State.GetBalance, big.NewInt(0), appState, &balanceCache)

	c.AddContractBalanceUpdate(nil, sender, appState.State.GetBalance, big.NewInt(11), appState, &balanceCache)
	appState.State.SetBalance(sender, big.NewInt(11))

	c.AddContractBurntCoins(address3, func(address common.Address) *big.Int {
		return big.NewInt(400)
	}, &balanceCache)
	c.ApplyContractBalanceUpdates(&balanceCache, nil)
	c.AddTxReceipt(&types.TxReceipt{Success: true}, appState)

	c.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(sender, big.NewInt(111))
	c.CompleteBalanceUpdate(appState)

	c.CompleteApplyingTx(appState)

	// When
	require.Equal(t, 0, big.NewInt(400).Cmp(c.stats.BurntCoins))
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr))
	require.Equal(t, 1, len(c.stats.BurntCoinsByAddr[address3]))
	require.Equal(t, db2.ContractReason, c.stats.BurntCoinsByAddr[address3][0].Reason)
	require.Equal(t, tx.Hash().Hex(), c.stats.BurntCoinsByAddr[address3][0].TxHash)
	require.Equal(t, "0.0000000000000004", c.stats.BurntCoinsByAddr[address3][0].Amount.String())

	require.Equal(t, 2, c.stats.BalanceUpdateAddrs.Cardinality())
	require.True(t, c.stats.BalanceUpdateAddrs.Contains(sender))
	require.True(t, c.stats.BalanceUpdateAddrs.Contains(address2))

	findContractBalanceUpdate := func(address common.Address) *db2.BalanceUpdate {
		for _, bu := range c.stats.BalanceUpdates {
			if address == bu.Address && bu.Reason == db2.ContractReason {
				return bu
			}
		}
		return nil
	}
	require.Equal(t, 3, len(c.stats.BalanceUpdates))
	bu := findContractBalanceUpdate(address2)
	require.Equal(t, big.NewInt(0), bu.BalanceOld)
	require.Equal(t, big.NewInt(200), bu.BalanceNew)
	require.Equal(t, uint16(3), bu.PenaltySecondsOld)
	require.Equal(t, uint16(3), bu.PenaltySecondsNew)
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

func TestStatsCollector_AddValidationReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddValidationReward(addr1, addr1, 7, nil, nil)
	c.AddValidationReward(addr5, addr5, 10, nil, nil)

	c.AddValidationReward(addr2, addr1, 2, big.NewInt(1), big.NewInt(2))
	c.AddValidationReward(addr2, addr2, 3, big.NewInt(3), big.NewInt(4))
	c.AddValidationReward(addr3, addr3, 4, big.NewInt(4), big.NewInt(5))
	c.AddValidationReward(addr2, addr4, 5, big.NewInt(6), big.NewInt(7))

	require.Equal(t, uint16(3), c.stats.RewardsStats.AgesByAddress[addr1.Hex()])
	require.Equal(t, uint16(4), c.stats.RewardsStats.AgesByAddress[addr2.Hex()])
	require.Equal(t, uint16(5), c.stats.RewardsStats.AgesByAddress[addr3.Hex()])
	require.Equal(t, uint16(6), c.stats.RewardsStats.AgesByAddress[addr4.Hex()])
	require.Equal(t, uint16(11), c.stats.RewardsStats.AgesByAddress[addr5.Hex()])

	require.Len(t, c.stats.RewardsStats.Rewards, 4)

	find := func(address common.Address) *RewardStats {
		for _, item := range c.stats.RewardsStats.Rewards {
			if address == item.Address {
				return item
			}
		}
		return nil
	}

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(1)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Validation, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(3)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Validation, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Validation, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(6)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Validation, find(addr4).Type)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards, 1)
	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[addr2].TotalRewards, 1)
	require.Equal(t, big.NewInt(7), c.stats.RewardsStats.DelegateesEpochRewards[addr2].TotalRewards[Validation].Balance)
	require.Equal(t, big.NewInt(0), c.stats.RewardsStats.DelegateesEpochRewards[addr2].TotalRewards[Validation].Stake)
	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards, 2)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr1].EpochRewards, 1)
	require.Equal(t, big.NewInt(1), c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr1].EpochRewards[Validation].Balance)
	require.Equal(t, big.NewInt(0), c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr1].EpochRewards[Validation].Stake)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr4].EpochRewards, 1)
	require.Equal(t, big.NewInt(6), c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr4].EpochRewards[Validation].Balance)
	require.Equal(t, big.NewInt(0), c.stats.RewardsStats.DelegateesEpochRewards[addr2].DelegatorsEpochRewards[addr4].EpochRewards[Validation].Stake)
}

func TestStatsCollector_AddRewardsWithDelegatee(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	delegator1, delegator2, delegator3, delegator4, delegatee1, delegatee2 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddValidationReward(delegatee1, delegatee1, 1, new(big.Int).SetInt64(1), new(big.Int).SetInt64(2))
	c.AddValidationReward(delegatee1, delegator1, 1, new(big.Int).SetInt64(1), new(big.Int).SetInt64(2))
	c.AddValidationReward(delegatee1, delegator2, 2, new(big.Int).SetInt64(3), new(big.Int).SetInt64(4))
	c.AddValidationReward(delegatee1, delegator3, 3, new(big.Int).SetInt64(5), new(big.Int).SetInt64(6))
	c.AddValidationReward(delegatee2, delegator4, 4, new(big.Int).SetInt64(7), new(big.Int).SetInt64(8))

	c.AddFlipsBasicReward(delegatee1, delegatee1, big.NewInt(9), big.NewInt(10), []*types.FlipToReward{
		{[]byte{0x1}, types.GradeNone, decimal.NewFromInt32(8)},
	})
	c.AddFlipsBasicReward(delegatee1, delegator1, big.NewInt(9), big.NewInt(10), []*types.FlipToReward{
		{[]byte{0x1}, types.GradeNone, decimal.NewFromInt32(8)},
	})
	c.AddFlipsBasicReward(delegatee1, delegator2, big.NewInt(11), big.NewInt(12), []*types.FlipToReward{
		{[]byte{0x1}, types.GradeNone, decimal.NewFromInt32(8)},
	})
	c.AddFlipsBasicReward(delegatee1, delegator3, big.NewInt(13), big.NewInt(14), []*types.FlipToReward{
		{[]byte{0x1}, types.GradeNone, decimal.NewFromInt32(8)},
	})
	c.AddFlipsBasicReward(delegatee2, delegator4, big.NewInt(15), big.NewInt(16), []*types.FlipToReward{
		{[]byte{0x1}, types.GradeNone, decimal.NewFromInt32(8)},
	})

	c.SetValidation(&types2.ValidationStats{
		Shards: map[common.ShardId]*types2.ValidationShardStats{
			1: {
				FlipCids: [][]byte{
					{0x1},
				},
			},
		},
	})

	c.AddReportedFlipsReward(delegatee1, delegatee1, 1, 1, big.NewInt(17), big.NewInt(18))
	c.AddReportedFlipsReward(delegatee1, delegator1, 1, 1, big.NewInt(17), big.NewInt(18))
	c.AddReportedFlipsReward(delegatee1, delegator2, 1, 1, big.NewInt(19), big.NewInt(20))
	c.AddReportedFlipsReward(delegatee1, delegator3, 1, 1, big.NewInt(21), big.NewInt(22))
	c.AddReportedFlipsReward(delegatee2, delegator4, 1, 1, big.NewInt(23), big.NewInt(24))
	c.AddReportedFlipsReward(delegatee2, delegator4, 1, 1, big.NewInt(100), big.NewInt(200))

	txHash := common.Hash{0x1, 0x2}
	c.AddInvitationsReward(delegatee1, delegatee1, big.NewInt(25), big.NewInt(26), 1, &txHash, 4, false)
	c.AddInvitationsReward(delegatee1, delegator1, big.NewInt(25), big.NewInt(26), 1, &txHash, 4, false)
	c.AddInvitationsReward(delegatee1, delegator2, big.NewInt(27), big.NewInt(28), 1, &txHash, 5, false)
	c.AddInvitationsReward(delegatee1, delegator3, big.NewInt(29), big.NewInt(30), 2, &txHash, 6, false)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards, 2)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards, 5)
	require.Equal(t, new(big.Int).SetInt64(9), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Validation].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Validation].Stake)

	require.Equal(t, new(big.Int).SetInt64(33), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Flips].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Flips].Stake)

	require.Equal(t, new(big.Int).SetInt64(57), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[ReportedFlips].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[ReportedFlips].Stake)

	require.Equal(t, new(big.Int).SetInt64(52), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Invitations].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Invitations].Stake)

	require.Equal(t, new(big.Int).SetInt64(29), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Invitations2].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].TotalRewards[Invitations2].Stake)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards, 3)
	require.Equal(t, new(big.Int).SetInt64(1), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator1].EpochRewards[Validation].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator1].EpochRewards[Validation].Stake)
	require.Equal(t, new(big.Int).SetInt64(3), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator2].EpochRewards[Validation].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator2].EpochRewards[Validation].Stake)
	require.Equal(t, new(big.Int).SetInt64(5), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator3].EpochRewards[Validation].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee1].DelegatorsEpochRewards[delegator3].EpochRewards[Validation].Stake)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards, 3)
	require.Equal(t, new(big.Int).SetInt64(7), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[Validation].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[Validation].Stake)

	require.Equal(t, new(big.Int).SetInt64(15), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[Flips].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[Flips].Stake)

	require.Equal(t, new(big.Int).SetInt64(123), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[ReportedFlips].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].TotalRewards[ReportedFlips].Stake)

	require.Len(t, c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].DelegatorsEpochRewards, 1)
	require.Equal(t, new(big.Int).SetInt64(123), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].DelegatorsEpochRewards[delegator4].EpochRewards[ReportedFlips].Balance)
	require.Equal(t, new(big.Int).SetInt64(0), c.stats.RewardsStats.DelegateesEpochRewards[delegatee2].DelegatorsEpochRewards[delegator4].EpochRewards[ReportedFlips].Stake)
}

func TestStatsCollector_AddFlipsReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddFlipsBasicReward(addr1, addr1, nil, nil, nil)
	c.AddFlipsBasicReward(addr5, addr5, nil, nil, nil)

	cid1, _ := cid.Parse("bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa")
	cid2, _ := cid.Parse("bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova")
	cid3, _ := cid.Parse("bafkreihcvhijrwwts3xl3zufbi2mjng5gltc7ojw2syue7zyritkq3gbii")

	c.AddFlipsBasicReward(addr2, addr1, big.NewInt(1), big.NewInt(2), []*types.FlipToReward{
		{cid1.Bytes(), types.GradeNone, decimal.NewFromInt32(8)},
		{cid2.Bytes(), types.GradeNone, decimal.NewFromInt32(6)},
	})
	c.AddFlipsBasicReward(addr2, addr2, big.NewInt(3), big.NewInt(4), []*types.FlipToReward{
		{cid3.Bytes(), types.GradeNone, decimal.NewFromInt32(4)},
	})
	c.AddFlipsBasicReward(addr3, addr3, big.NewInt(4), big.NewInt(5), nil)
	c.AddFlipsBasicReward(addr2, addr4, big.NewInt(6), big.NewInt(7), nil)

	require.Len(t, c.stats.RewardsStats.RewardedFlipCids, 3)
	require.Equal(t, cid1.String(), c.stats.RewardsStats.RewardedFlipCids[0])
	require.Equal(t, cid2.String(), c.stats.RewardsStats.RewardedFlipCids[1])
	require.Equal(t, cid3.String(), c.stats.RewardsStats.RewardedFlipCids[2])

	require.Len(t, c.stats.RewardsStats.Rewards, 4)

	find := func(address common.Address) *RewardStats {
		for _, item := range c.stats.RewardsStats.Rewards {
			if address == item.Address {
				return item
			}
		}
		return nil
	}

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(1)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Flips, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(3)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Flips, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Flips, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(6)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Flips, find(addr4).Type)
}

func TestStatsCollector_AddFlipsExtraReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddFlipsExtraReward(addr1, addr1, nil, nil, nil)
	c.AddFlipsExtraReward(addr5, addr5, nil, nil, nil)

	cid1, _ := cid.Parse("bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa")
	cid2, _ := cid.Parse("bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova")
	cid3, _ := cid.Parse("bafkreihcvhijrwwts3xl3zufbi2mjng5gltc7ojw2syue7zyritkq3gbii")

	c.AddFlipsBasicReward(addr2, addr1, big.NewInt(1), big.NewInt(2), []*types.FlipToReward{
		{},
		{},
		{},
		{cid1.Bytes(), types.GradeNone, decimal.NewFromInt32(8)},
		{cid2.Bytes(), types.GradeNone, decimal.NewFromInt32(6)},
	})
	c.AddFlipsExtraReward(addr2, addr1, big.NewInt(1), big.NewInt(2), []*types.FlipToReward{
		{cid1.Bytes(), types.GradeNone, decimal.NewFromInt32(8)},
		{cid2.Bytes(), types.GradeNone, decimal.NewFromInt32(6)},
	})

	c.AddFlipsBasicReward(addr2, addr2, big.NewInt(3), big.NewInt(4), []*types.FlipToReward{
		{},
		{},
		{},
		{cid3.Bytes(), types.GradeNone, decimal.NewFromInt32(4)},
	})
	c.AddFlipsExtraReward(addr2, addr2, big.NewInt(3), big.NewInt(4), []*types.FlipToReward{
		{cid3.Bytes(), types.GradeNone, decimal.NewFromInt32(4)},
	})

	c.AddFlipsExtraReward(addr3, addr3, big.NewInt(4), big.NewInt(5), nil)
	c.AddFlipsExtraReward(addr2, addr4, big.NewInt(6), big.NewInt(7), nil)

	require.Len(t, c.stats.RewardsStats.RewardedExtraFlipCids, 3)
	require.Equal(t, cid1.String(), c.stats.RewardsStats.RewardedExtraFlipCids[0])
	require.Equal(t, cid2.String(), c.stats.RewardsStats.RewardedExtraFlipCids[1])
	require.Equal(t, cid3.String(), c.stats.RewardsStats.RewardedExtraFlipCids[2])

	require.Len(t, c.stats.RewardsStats.Rewards, 6)

	find := func(address common.Address) *RewardStats {
		for _, item := range c.stats.RewardsStats.Rewards {
			if item.Type == ExtraFlips && address == item.Address {
				return item
			}
		}
		return nil
	}

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(1)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, ExtraFlips, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(3)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, ExtraFlips, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, ExtraFlips, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(6)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, ExtraFlips, find(addr4).Type)
}

func TestStatsCollector_AddReportedFlipsReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	cid1, _ := cid.Parse("bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa")
	cid2, _ := cid.Parse("bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova")
	cid3, _ := cid.Parse("bafkreihcvhijrwwts3xl3zufbi2mjng5gltc7ojw2syue7zyritkq3gbii")

	c.SetValidation(&types2.ValidationStats{
		Shards: map[common.ShardId]*types2.ValidationShardStats{
			1: {
				FlipCids: [][]byte{
					cid1.Bytes(),
					cid2.Bytes(),
					cid3.Bytes(),
				},
			},
		},
	})

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddReportedFlipsReward(addr1, addr1, 1, -1, nil, nil)
	c.AddReportedFlipsReward(addr5, addr5, 1, 0, nil, nil)

	c.AddReportedFlipsReward(addr2, addr1, 1, 1, big.NewInt(1), big.NewInt(2))
	c.AddReportedFlipsReward(addr2, addr2, 1, 1, big.NewInt(3), big.NewInt(4))
	c.AddReportedFlipsReward(addr3, addr3, 1, 1, big.NewInt(4), big.NewInt(5))
	c.AddReportedFlipsReward(addr2, addr4, 1, 2, big.NewInt(6), big.NewInt(7))

	require.Len(t, c.stats.RewardsStats.ReportedFlipRewards, 5)
	require.Equal(t, cid1.String(), c.stats.RewardsStats.ReportedFlipRewards[0].Cid)
	require.Equal(t, addr5.Hex(), c.stats.RewardsStats.ReportedFlipRewards[0].Address)
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[0].Balance.Sign())
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[0].Stake.Sign())

	require.Equal(t, cid2.String(), c.stats.RewardsStats.ReportedFlipRewards[1].Cid)
	require.Equal(t, addr1.Hex(), c.stats.RewardsStats.ReportedFlipRewards[1].Address)
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[1].Balance.Cmp(blockchain.ConvertToFloat(big.NewInt(1))))
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[1].Stake.Cmp(blockchain.ConvertToFloat(big.NewInt(2))))

	require.Equal(t, cid2.String(), c.stats.RewardsStats.ReportedFlipRewards[2].Cid)
	require.Equal(t, addr2.Hex(), c.stats.RewardsStats.ReportedFlipRewards[2].Address)
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[2].Balance.Cmp(blockchain.ConvertToFloat(big.NewInt(3))))
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[2].Stake.Cmp(blockchain.ConvertToFloat(big.NewInt(4))))

	require.Equal(t, cid2.String(), c.stats.RewardsStats.ReportedFlipRewards[3].Cid)
	require.Equal(t, addr3.Hex(), c.stats.RewardsStats.ReportedFlipRewards[3].Address)
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[3].Balance.Cmp(blockchain.ConvertToFloat(big.NewInt(4))))
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[3].Stake.Cmp(blockchain.ConvertToFloat(big.NewInt(5))))

	require.Equal(t, cid3.String(), c.stats.RewardsStats.ReportedFlipRewards[4].Cid)
	require.Equal(t, addr4.Hex(), c.stats.RewardsStats.ReportedFlipRewards[4].Address)
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[4].Balance.Cmp(blockchain.ConvertToFloat(big.NewInt(6))))
	require.Zero(t, c.stats.RewardsStats.ReportedFlipRewards[4].Stake.Cmp(blockchain.ConvertToFloat(big.NewInt(7))))

	require.Len(t, c.stats.RewardsStats.Rewards, 4)

	find := func(address common.Address) *RewardStats {
		for _, item := range c.stats.RewardsStats.Rewards {
			if address == item.Address {
				return item
			}
		}
		return nil
	}

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(1)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, ReportedFlips, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(3)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, ReportedFlips, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, ReportedFlips, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(6)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, ReportedFlips, find(addr4).Type)
}

func TestStatsCollector_AddInvitationsReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddInvitationsReward(addr1, addr1, nil, nil, 1, nil, 2, false)
	c.AddInvitationsReward(addr5, addr5, nil, nil, 1, nil, 3, false)

	txHash := common.Hash{0x1, 0x2}

	c.AddInvitationsReward(addr2, addr1, big.NewInt(1), big.NewInt(2), 1, &txHash, 4, false)
	c.AddInvitationsReward(addr2, addr2, big.NewInt(3), big.NewInt(4), 1, &txHash, 5, false)
	c.AddInvitationsReward(addr3, addr3, big.NewInt(4), big.NewInt(5), 1, &txHash, 6, false)
	c.AddInvitationsReward(addr2, addr4, big.NewInt(6), big.NewInt(7), 1, &txHash, 7, false)

	require.Empty(t, c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType)
	require.Len(t, c.stats.RewardsStats.RewardedInvites, 4)

	require.Len(t, c.stats.RewardsStats.Rewards, 4)

	find := func(address common.Address) *RewardStats {
		for _, item := range c.stats.RewardsStats.Rewards {
			if address == item.Address {
				return item
			}
		}
		return nil
	}

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(1)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Invitations, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(3)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Invitations, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Invitations, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(6)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Invitations, find(addr4).Type)
}

func TestStatsCollector_AddProposerReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddProposerReward(addr1, addr1, nil, nil, nil)
	require.Empty(t, c.stats.MiningRewards)

	c.AddProposerReward(addr2, addr1, big.NewInt(1), big.NewInt(2), nil)
	require.Len(t, c.stats.MiningRewards, 2)

	require.True(t, c.stats.MiningRewards[0].Proposer)
	require.Equal(t, addr2, c.stats.MiningRewards[0].Address)
	require.Zero(t, c.stats.MiningRewards[0].Balance.Cmp(big.NewInt(1)))
	require.Zero(t, c.stats.MiningRewards[0].Stake.Sign())

	require.True(t, c.stats.MiningRewards[1].Proposer)
	require.Equal(t, addr1, c.stats.MiningRewards[1].Address)
	require.Zero(t, c.stats.MiningRewards[1].Balance.Sign())
	require.Zero(t, c.stats.MiningRewards[1].Stake.Cmp(big.NewInt(2)))

	c.CompleteCollecting()
	c.EnableCollecting()

	c.AddProposerReward(addr3, addr3, big.NewInt(3), big.NewInt(4), nil)
	require.Len(t, c.stats.MiningRewards, 1)

	require.True(t, c.stats.MiningRewards[0].Proposer)
	require.Equal(t, addr3, c.stats.MiningRewards[0].Address)
	require.Zero(t, c.stats.MiningRewards[0].Balance.Cmp(big.NewInt(3)))
	require.Zero(t, c.stats.MiningRewards[0].Stake.Cmp(big.NewInt(4)))
}

func TestStatsCollector_AddFinalCommitteeReward(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddFinalCommitteeReward(addr1, addr1, big.NewInt(1), big.NewInt(2), nil)
	c.AddFinalCommitteeReward(addr1, addr2, big.NewInt(3), big.NewInt(4), nil)
	c.AddFinalCommitteeReward(addr3, addr3, big.NewInt(5), big.NewInt(6), nil)

	require.Len(t, c.stats.MiningRewards, 3)

	require.False(t, c.stats.MiningRewards[0].Proposer)
	require.Equal(t, addr1, c.stats.MiningRewards[0].Address)
	require.Zero(t, c.stats.MiningRewards[0].Balance.Cmp(big.NewInt(4)))
	require.Zero(t, c.stats.MiningRewards[0].Stake.Cmp(big.NewInt(2)))

	require.False(t, c.stats.MiningRewards[1].Proposer)
	require.Equal(t, addr2, c.stats.MiningRewards[1].Address)
	require.Zero(t, c.stats.MiningRewards[1].Balance.Sign())
	require.Zero(t, c.stats.MiningRewards[1].Stake.Cmp(big.NewInt(4)))

	require.False(t, c.stats.MiningRewards[2].Proposer)
	require.Equal(t, addr3, c.stats.MiningRewards[2].Address)
	require.Zero(t, c.stats.MiningRewards[2].Balance.Cmp(big.NewInt(5)))
	require.Zero(t, c.stats.MiningRewards[2].Stake.Cmp(big.NewInt(6)))

	require.Len(t, c.stats.OriginalFinalCommittee, 3)
	require.Len(t, c.stats.PoolFinalCommittee, 2)
}

func Test_BeginVerifiedStakeTransferBalanceUpdateAndCompleteBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginVerifiedStakeTransferBalanceUpdate(addr1, addr1, appState)
	c.BeginVerifiedStakeTransferBalanceUpdate(addr2, addr1, appState)
	c.BeginVerifiedStakeTransferBalanceUpdate(addr3, addr1, appState)

	require.Len(t, c.pending.balanceUpdates, 5)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.VerifiedStakeTransferReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(1000))
	appState.State.SetBalance(addr2, big.NewInt(2000))
	appState.State.SetBalance(addr3, big.NewInt(3000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 3)
}

func Test_BeginProposerRewardBalanceUpdateAndCompleteBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2 := tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginProposerRewardBalanceUpdate(addr1, addr1, nil, appState)

	require.Len(t, c.pending.balanceUpdates, 1)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.ProposerRewardReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(1000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 1)

	c.CompleteCollecting()
	c.EnableCollecting()

	c.BeginProposerRewardBalanceUpdate(addr1, addr2, nil, appState)

	require.Len(t, c.pending.balanceUpdates, 2)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.ProposerRewardReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(2000))
	appState.State.SetBalance(addr2, big.NewInt(3000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 2)
}

func Test_BeginCommitteeRewardBalanceUpdateAndCompleteBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr1, nil, appState)
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr2, nil, appState)
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr3, nil, appState)

	require.Len(t, c.pending.balanceUpdates, 5)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.CommitteeRewardReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(1000))
	appState.State.SetBalance(addr2, big.NewInt(2000))
	appState.State.SetBalance(addr3, big.NewInt(3000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 3)
}

func Test_BeginEpochRewardBalanceUpdateAndCompleteBalanceUpdate(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginEpochRewardBalanceUpdate(addr1, addr1, appState)
	c.BeginEpochRewardBalanceUpdate(addr1, addr2, appState)
	c.BeginEpochRewardBalanceUpdate(addr1, addr3, appState)

	require.Len(t, c.pending.balanceUpdates, 5)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.EpochRewardReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(1000))
	appState.State.SetBalance(addr2, big.NewInt(2000))
	appState.State.SetBalance(addr3, big.NewInt(3000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 3)
}

func Test_killedPenaltyBurntCoins(t *testing.T) {
	c := newStatsCollector()
	c.EnableCollecting()

	killedAddr1, killedAddr2, killedAddr3, addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	delegatee, killedDelegatee := tests.GetRandAddr(), tests.GetRandAddr()

	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	appState.State.SetState(killedAddr1, state.Killed)
	appState.State.SetState(killedAddr2, state.Killed)
	appState.State.SetState(killedAddr3, state.Killed)
	appState.State.SetState(killedDelegatee, state.Killed)
	appState.State.SetState(addr1, state.Newbie)
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetState(addr3, state.Human)
	appState.State.SetState(delegatee, state.Newbie)

	baseReward := new(big.Int).SetInt64(40)
	baseStake := new(big.Int).SetInt64(10)

	calculateStakeToAdd := func(penaltySub *big.Int) *big.Int {
		res := new(big.Int).Add(baseReward, baseStake)
		res.Sub(res, penaltySub)
		if res.Cmp(baseStake) >= 0 {
			return new(big.Int).Set(baseStake)
		}
		return res
	}

	// Cases without delegatee
	penaltySub := new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(killedAddr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedAddr1, penaltySub)

	expectedBurnt := int64(50)
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(killedAddr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedAddr2, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(killedAddr3, calculateStakeToAdd(penaltySub), appState)

	c.AddPenaltyBurntCoins(killedAddr3, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(addr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(addr1, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(addr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(addr2, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(addr3, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(addr3, penaltySub)

	expectedBurnt += 35
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	// Cases with delegatee
	penaltySub = new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(killedAddr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(killedAddr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(killedAddr3, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(addr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(addr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(addr3, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(delegatee, penaltySub)

	expectedBurnt += 35
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	// Cases with killed delegatee
	penaltySub = new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(killedAddr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(killedAddr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(killedAddr3, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).Add(baseReward, baseStake)
	c.AfterAddStake(addr1, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 50
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(45)
	c.AfterAddStake(addr2, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 45
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)

	penaltySub = new(big.Int).SetInt64(35)
	c.AfterAddStake(addr3, calculateStakeToAdd(penaltySub), appState)
	c.AddPenaltyBurntCoins(killedDelegatee, penaltySub)

	expectedBurnt += 35
	require.Equal(t, new(big.Int).SetInt64(expectedBurnt), c.stats.BurntCoins)
}
