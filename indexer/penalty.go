package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

func detectChargedPenalty(block *types.Block, newState *appstate.AppState) *db.Penalty {
	if !block.Header.Flags().HasFlag(types.OfflineCommit) {
		return nil
	}
	address := *block.Header.OfflineAddr()
	return &db.Penalty{
		Address: conversion.ConvertAddress(address),
		Penalty: blockchain.ConvertToFloat(newState.State.GetPenalty(address)),
	}
}

func convertBurntPenalties(burntPenaltiesByAddr map[common.Address]*big.Int) []db.Penalty {
	var res []db.Penalty
	for addr, burntAmount := range burntPenaltiesByAddr {
		res = append(res, db.Penalty{
			Address: conversion.ConvertAddress(addr),
			Penalty: blockchain.ConvertToFloat(burntAmount),
		})
	}
	return res
}
