package indexer

import (
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
)

type conversionContext struct {
	blockHeight       uint64
	submittedFlips    []db.Flip
	flipKeys          []db.FlipKey
	flipsWords        []db.FlipWords
	flipsData         []db.FlipData
	flipSizeUpdates   []db.FlipSizeUpdate
	addresses         map[string]*db.Address
	prevStateReadOnly *appstate.AppState
	newStateReadOnly  *appstate.AppState
}

func (ctx *conversionContext) getAddresses() []db.Address {
	var addresses []db.Address
	for _, addr := range ctx.addresses {
		addresses = append(addresses, *addr)
	}
	return addresses
}
