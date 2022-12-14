package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

func convertChargedPenalties(chargedPenaltiesByAddr map[common.Address]*big.Int, chargedPenaltySecondsByAddr map[common.Address]stats.Penalty) []db.Penalty {
	if len(chargedPenaltiesByAddr) == 0 && len(chargedPenaltySecondsByAddr) == 0 {
		return nil
	}
	res := make([]db.Penalty, 0, len(chargedPenaltiesByAddr)+len(chargedPenaltySecondsByAddr))
	for addr, amount := range chargedPenaltiesByAddr {
		res = append(res, db.Penalty{
			Address: conversion.ConvertAddress(addr),
			Penalty: blockchain.ConvertToFloat(amount),
		})
	}
	for addr, penalty := range chargedPenaltySecondsByAddr {
		var inheritedFrom string
		if penalty.InheritedFrom != nil {
			inheritedFrom = conversion.ConvertAddress(*penalty.InheritedFrom)
		}
		res = append(res, db.Penalty{
			Address:       conversion.ConvertAddress(addr),
			Seconds:       penalty.Seconds,
			InheritedFrom: inheritedFrom,
		})
	}
	return res
}
