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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginPenaltyBalanceUpdate(addr, appState)
	appState.State.SetPenalty(addr, big.NewInt(1))
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	c.BeginProposerRewardBalanceUpdate(addr, addr, appState)
	c.CompleteBalanceUpdate(appState)
	// then
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
	require.Equal(t, 0, len(c.stats.BalanceUpdates))

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, addr, appState)
	appState.State.SetBalance(addr, big.NewInt(12))
	appState.State.AddStake(addr, big.NewInt(2))
	appState.State.SetPenalty(addr, big.NewInt(3))
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
	require.Equal(t, big.NewInt(3), c.stats.BalanceUpdates[0].PenaltyNew)

	// when
	c.CompleteCollecting()
	c.EnableCollecting()
	c.BeginProposerRewardBalanceUpdate(addr, addr, appState)
	appState.State.SetState(addr, state.Killed)
	c.CompleteBalanceUpdate(appState)
	require.Equal(t, 0, len(c.pending.balanceUpdates))
	require.Equal(t, 0, len(c.pending.balanceUpdatesByReasonAndAddr))
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
	c := &statsCollector{}
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

func TestStatsCollector_AddValidationReward(t *testing.T) {
	c := &statsCollector{}
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

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Validation, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(10)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Validation, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Validation, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Validation, find(addr4).Type)
}

func TestStatsCollector_AddFlipsReward(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()

	addr1, addr2, addr3, addr4, addr5 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddFlipsReward(addr1, addr1, nil, nil, nil)
	c.AddFlipsReward(addr5, addr5, nil, nil, nil)

	cid1, _ := cid.Parse("bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa")
	cid2, _ := cid.Parse("bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova")
	cid3, _ := cid.Parse("bafkreihcvhijrwwts3xl3zufbi2mjng5gltc7ojw2syue7zyritkq3gbii")

	c.AddFlipsReward(addr2, addr1, big.NewInt(1), big.NewInt(2), []*types.FlipToReward{
		{cid1.Bytes(), types.GradeA},
		{cid2.Bytes(), types.GradeB},
	})
	c.AddFlipsReward(addr2, addr2, big.NewInt(3), big.NewInt(4), []*types.FlipToReward{
		{cid3.Bytes(), types.GradeC},
	})
	c.AddFlipsReward(addr3, addr3, big.NewInt(4), big.NewInt(5), nil)
	c.AddFlipsReward(addr2, addr4, big.NewInt(6), big.NewInt(7), nil)

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

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Flips, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(10)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Flips, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Flips, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Flips, find(addr4).Type)
}

func TestStatsCollector_AddReportedFlipsReward(t *testing.T) {
	c := &statsCollector{}
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

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, ReportedFlips, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(10)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, ReportedFlips, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, ReportedFlips, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, ReportedFlips, find(addr4).Type)
}

func TestStatsCollector_AddInvitationsReward(t *testing.T) {
	c := &statsCollector{}
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

	require.Zero(t, find(addr1).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr1).Stake.Cmp(big.NewInt(2)))
	require.Equal(t, Invitations, find(addr1).Type)

	require.Zero(t, find(addr2).Balance.Cmp(big.NewInt(10)))
	require.Zero(t, find(addr2).Stake.Cmp(big.NewInt(4)))
	require.Equal(t, Invitations, find(addr2).Type)

	require.Zero(t, find(addr3).Balance.Cmp(big.NewInt(4)))
	require.Zero(t, find(addr3).Stake.Cmp(big.NewInt(5)))
	require.Equal(t, Invitations, find(addr3).Type)

	require.Zero(t, find(addr4).Balance.Cmp(big.NewInt(0)))
	require.Zero(t, find(addr4).Stake.Cmp(big.NewInt(7)))
	require.Equal(t, Invitations, find(addr4).Type)
}

func TestStatsCollector_AddProposerReward(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddProposerReward(addr1, addr1, nil, nil)
	require.Empty(t, c.stats.MiningRewards)

	c.AddProposerReward(addr2, addr1, big.NewInt(1), big.NewInt(2))
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

	c.AddProposerReward(addr3, addr3, big.NewInt(3), big.NewInt(4))
	require.Len(t, c.stats.MiningRewards, 1)

	require.True(t, c.stats.MiningRewards[0].Proposer)
	require.Equal(t, addr3, c.stats.MiningRewards[0].Address)
	require.Zero(t, c.stats.MiningRewards[0].Balance.Cmp(big.NewInt(3)))
	require.Zero(t, c.stats.MiningRewards[0].Stake.Cmp(big.NewInt(4)))
}

func TestStatsCollector_AddFinalCommitteeReward(t *testing.T) {
	c := &statsCollector{}
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()

	c.AddFinalCommitteeReward(addr1, addr1, big.NewInt(1), big.NewInt(2))
	c.AddFinalCommitteeReward(addr1, addr2, big.NewInt(3), big.NewInt(4))
	c.AddFinalCommitteeReward(addr3, addr3, big.NewInt(5), big.NewInt(6))

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
	c := &statsCollector{}
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
	c := &statsCollector{}
	c.EnableCollecting()

	addr1, addr2 := tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginProposerRewardBalanceUpdate(addr1, addr1, appState)

	require.Len(t, c.pending.balanceUpdates, 1)
	for _, bu := range c.pending.balanceUpdates {
		require.Equal(t, db2.ProposerRewardReason, bu.Reason)
	}

	appState.State.SetBalance(addr1, big.NewInt(1000))
	c.CompleteBalanceUpdate(appState)

	require.Len(t, c.stats.BalanceUpdates, 1)

	c.CompleteCollecting()
	c.EnableCollecting()

	c.BeginProposerRewardBalanceUpdate(addr1, addr2, appState)

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
	c := &statsCollector{}
	c.EnableCollecting()

	addr1, addr2, addr3 := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	appState, _ := appstate.NewAppState(db.NewMemDB(), eventbus.New())
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr1, appState)
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr2, appState)
	c.BeginCommitteeRewardBalanceUpdate(addr1, addr3, appState)

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
	c := &statsCollector{}
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
