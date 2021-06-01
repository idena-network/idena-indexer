package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

func convertChargedPenalties(chargedPenaltiesByAddr map[common.Address]*big.Int) []db.Penalty {
	if len(chargedPenaltiesByAddr) == 0 {
		return nil
	}
	res := make([]db.Penalty, 0, len(chargedPenaltiesByAddr))
	for addr, amount := range chargedPenaltiesByAddr {
		res = append(res, db.Penalty{
			Address: conversion.ConvertAddress(addr),
			Penalty: blockchain.ConvertToFloat(amount),
		})
	}
	return res
}
