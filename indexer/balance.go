package indexer

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

type balanceDiff struct {
	balance *big.Int
	stake   *big.Int
}

func determineBalanceUpdates(isFirstBlock bool,
	balanceUpdateAddrs mapset.Set,
	prevAppState,
	newAppState *appstate.AppState) ([]db.Balance, balanceDiff) {

	// Add all addrs to list for determining balance updates in case of genesis block
	if isFirstBlock {
		if balanceUpdateAddrs == nil {
			balanceUpdateAddrs = mapset.NewSet()
		}
		prevAppState.State.IterateAccounts(func(key []byte, _ []byte) bool {
			if key == nil {
				return true
			}
			addr := conversion.BytesToAddr(key)
			balanceUpdateAddrs.Add(addr)
			return false
		})
	}

	var res []db.Balance
	diff := balanceDiff{
		balance: big.NewInt(0),
		stake:   big.NewInt(0),
	}
	if balanceUpdateAddrs != nil {
		for _, a := range balanceUpdateAddrs.ToSlice() {
			addr := a.(common.Address)
			prevBalance := prevAppState.State.GetBalance(addr)
			prevStake := prevAppState.State.GetStakeBalance(addr)
			newBalance := newAppState.State.GetBalance(addr)
			newStake := newAppState.State.GetStakeBalance(addr)
			res = append(res, db.Balance{
				Address: conversion.ConvertAddress(addr),
				Balance: blockchain.ConvertToFloat(newBalance),
				Stake:   blockchain.ConvertToFloat(newStake),
			})

			diff.balance.Add(diff.balance, newBalance)
			diff.balance.Sub(diff.balance, prevBalance)
			diff.stake.Add(diff.stake, newStake)
			diff.stake.Sub(diff.stake, prevStake)
		}
	}
	return res, diff
}
