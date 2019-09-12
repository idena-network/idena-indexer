package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
)

func (indexer *Indexer) determinePenalty(block *types.Block, ctx *conversionContext) *db.Penalty {
	if !block.Header.Flags().HasFlag(types.OfflineCommit) {
		return nil
	}
	address := *block.Header.OfflineAddr()
	return &db.Penalty{
		Address: conversion.ConvertAddress(address),
		Penalty: blockchain.ConvertToFloat(ctx.newStateReadOnly.State.GetPenalty(address)),
	}
}
