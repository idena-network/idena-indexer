package indexer

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
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

func detectBurntPenalties(
	block *types.Block,
	prevState *appstate.AppState,
	newState *appstate.AppState,
	chain *blockchain.Blockchain,
) []db.Penalty {
	if block.Header.Flags().HasFlag(types.ValidationFinished) {
		return getAllIdentitiesRemainingPenalties(prevState)
	}
	rewardedAddresses := getRewardedAddresses(block, prevState, chain)

	var res []db.Penalty
	for _, a := range rewardedAddresses.ToSlice() {
		address := a.(common.Address)
		penalty := prevState.State.GetPenalty(address)
		if penalty == nil || penalty.Sign() != 1 {
			continue
		}
		newPenalty := newState.State.GetPenalty(address)
		if newPenalty == nil || newPenalty.Sign() != 1 {
			res = append(res, db.Penalty{
				Address: conversion.ConvertAddress(address),
				Penalty: blockchain.ConvertToFloat(newPenalty),
			})
		}
	}

	addressesWithNewPenalty := getAddressesWithNewPenalty(block)
	for _, a := range addressesWithNewPenalty.ToSlice() {
		address := a.(common.Address)
		penalty := prevState.State.GetPenalty(address)
		if penalty == nil || penalty.Sign() != 1 {
			continue
		}
		if rewardedAddresses.Contains(address) {
			p := prevState.State.GetPenalty(address)
			if p != nil && p.Sign() == 1 {
				// todo: there can be case when address is rewarded and got new penalty at the same time - then last payment for previous penalty will be missed
				log.Warn(fmt.Sprintf("Missed last penalty payment for %s", conversion.ConvertAddress(address)))
			}
		}
		res = append(res, db.Penalty{
			Address: conversion.ConvertAddress(address),
			Penalty: blockchain.ConvertToFloat(penalty),
		})
	}

	return res
}

func getAllIdentitiesRemainingPenalties(appState *appstate.AppState) []db.Penalty {
	var res []db.Penalty
	appState.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		penalty := appState.State.GetPenalty(addr)
		if penalty == nil || penalty.Sign() != 1 {
			return
		}
		res = append(res, db.Penalty{
			Address: conversion.ConvertAddress(addr),
			Penalty: blockchain.ConvertToFloat(penalty),
		})
	})
	return res
}

func getRewardedAddresses(block *types.Block, prevState *appstate.AppState, chain *blockchain.Blockchain) mapset.Set {
	res := mapset.NewSet()
	if block.IsEmpty() {
		return res
	}
	res.Add(getBlockProposerAddress(block))
	res = res.Union(getFinalCommittee(block, prevState, chain))
	return res
}

func getBlockProposerAddress(block *types.Block) common.Address {
	return block.Header.ProposedHeader.Coinbase
}

func getFinalCommittee(block *types.Block, prevState *appstate.AppState, chain *blockchain.Blockchain) mapset.Set {
	prevBlock := chain.GetBlockHeaderByHeight(block.Height() - 1)
	validatorsCache := validators.NewValidatorsCache(prevState.IdentityState, prevState.State.GodAddress())
	validatorsCache.Load()
	res := mapset.NewSet()
	blockValidators := validatorsCache.GetOnlineValidators(prevBlock.Seed(), block.Height(), 1000, chain.GetCommitteSize(validatorsCache, true))
	if blockValidators != nil {
		res = res.Union(blockValidators)
	}
	return res
}

func getAddressesWithNewPenalty(block *types.Block) mapset.Set {
	res := mapset.NewSet()
	if block.Header.Flags().HasFlag(types.OfflineCommit) {
		res.Add(*block.Header.OfflineAddr())
	}
	return res
}
