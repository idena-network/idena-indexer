package indexer

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

type TxBalanceUpdateDetector struct {
	detector *balanceUpdateDetector
	txHash   string
}

func NewTxBalanceUpdateDetector(tx *types.Transaction, prevState *appstate.AppState) *TxBalanceUpdateDetector {
	res := TxBalanceUpdateDetector{
		txHash: convertHash(tx.Hash()),
	}
	sender, _ := types.Sender(tx)
	addresses := []common.Address{sender}
	if tx.To != nil && *tx.To != sender {
		addresses = append(addresses, *tx.To)
	}
	res.detector = newBalanceUpdateDetector(prevState, addresses...)
	return &res
}

func (d *TxBalanceUpdateDetector) GetUpdates(state *appstate.AppState) ([]db.Balance, *balanceDiff) {
	updates, diff := d.detector.getUpdates(state)
	for i := range updates {
		updates[i].TxHash = d.txHash
	}
	return updates, diff
}

type BlockBalanceUpdateDetector struct {
	detector     *balanceUpdateDetector
	isFirstBlock bool
}

func NewBlockBalanceUpdateDetector(block *types.Block, prevState *appstate.AppState, blockValidators mapset.Set, ctx *conversionContext) *BlockBalanceUpdateDetector {
	res := BlockBalanceUpdateDetector{}
	var addresses []common.Address
	if isFirstBlock(block) {
		res.isFirstBlock = true
		prevState.State.IterateAccounts(func(key []byte, _ []byte) bool {
			if key == nil {
				return true
			}
			addr := bytesToAddr(key)
			addresses = append(addresses, addr)
			convertedAddress := convertAddress(addr)
			ctx.addresses[convertedAddress] = &db.Address{
				Address: convertedAddress,
			}
			return false
		})
	} else {
		if !block.IsEmpty() {
			if blockValidators == nil || !blockValidators.Contains(block.Header.ProposedHeader.Coinbase) {
				addresses = append(addresses, block.Header.ProposedHeader.Coinbase)
			}
			if blockValidators != nil {
				for _, address := range blockValidators.ToSlice() {
					addresses = append(addresses, address.(common.Address))
				}
			}
		}
		if block.Header.Flags().HasFlag(types.IdentityUpdate) {
			prevState.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
				if block.IsEmpty() || !(addr == block.Header.ProposedHeader.Coinbase || (blockValidators != nil && blockValidators.Contains(addr))) {
					addresses = append(addresses, addr)
				}
			})
		}
	}

	res.detector = newBalanceUpdateDetector(prevState, addresses...)
	return &res
}

func bytesToAddr(bytes []byte) common.Address {
	addr := common.Address{}
	addr.SetBytes(bytes[1:])
	return addr
}

func (d *BlockBalanceUpdateDetector) GetUpdates(state *appstate.AppState) ([]db.Balance, *balanceDiff) {
	if d.isFirstBlock {
		var res []db.Balance
		var totalDiff *balanceDiff
		for _, prevBalance := range d.detector.prevBalances {
			res = append(res, db.Balance{
				Address: convertAddress(prevBalance.address),
				Balance: blockchain.ConvertToFloat(prevBalance.balance),
				Stake:   blockchain.ConvertToFloat(prevBalance.stake),
			})
			diff := balanceDiff{
				balance:    prevBalance.balance,
				stake:      prevBalance.stake,
				burntStake: new(big.Int),
			}
			if totalDiff == nil {
				totalDiff = &diff
			} else {
				totalDiff.Add(&diff)
			}
		}
		return res, totalDiff
	}
	return d.detector.getUpdates(state)
}

type balanceUpdateDetector struct {
	prevBalances []addrBalance
}

type addrBalance struct {
	address common.Address
	balance *big.Int
	stake   *big.Int
}

type balanceDiff struct {
	balance    *big.Int
	stake      *big.Int
	burntStake *big.Int
}

func (diff *balanceDiff) Add(d *balanceDiff) {
	diff.balance = new(big.Int).Add(diff.balance, d.balance)
	diff.stake = new(big.Int).Add(diff.stake, d.stake)
	diff.burntStake = new(big.Int).Add(diff.burntStake, d.burntStake)
}

func newBalanceUpdateDetector(prevState *appstate.AppState, addresses ...common.Address) *balanceUpdateDetector {
	res := balanceUpdateDetector{}
	gotAddresses := mapset.NewSet()
	for _, address := range addresses {
		if !gotAddresses.Contains(address) {
			gotAddresses.Add(address)
		} else {
			log.Warn("Got duplicated address for balance update detection")
			continue
		}
		res.prevBalances = append(res.prevBalances, addrBalance{
			address: address,
			balance: prevState.State.GetBalance(address),
			stake:   prevState.State.GetStakeBalance(address),
		})
	}
	return &res
}

func (d *balanceUpdateDetector) getUpdates(state *appstate.AppState) ([]db.Balance, *balanceDiff) {
	var res []db.Balance
	var totalDiff *balanceDiff
	for _, prevBalance := range d.prevBalances {
		if update, diff, ok := d.detectUpdate(prevBalance, state); ok {
			res = append(res, update)
			if totalDiff == nil {
				totalDiff = &diff
			} else {
				totalDiff.Add(&diff)
			}
		}
	}
	return res, totalDiff
}

func (d *balanceUpdateDetector) detectUpdate(ab addrBalance, state *appstate.AppState) (db.Balance, balanceDiff, bool) {
	prevBalance := ab.balance
	prevStake := ab.stake
	balance := state.State.GetBalance(ab.address)
	stake := state.State.GetStakeBalance(ab.address)
	if balance.Cmp(prevBalance) == 0 && stake.Cmp(prevStake) == 0 {
		return db.Balance{}, balanceDiff{}, false
	}
	burntStake := new(big.Int)
	if stake.Cmp(prevStake) == -1 {
		burntStake = burntStake.Sub(prevStake, stake)
	}
	return db.Balance{
			Address: convertAddress(ab.address),
			Balance: blockchain.ConvertToFloat(balance),
			Stake:   blockchain.ConvertToFloat(stake),
		}, balanceDiff{
			balance:    new(big.Int).Sub(balance, prevBalance),
			stake:      new(big.Int).Sub(stake, prevStake),
			burntStake: burntStake,
		}, true
}
