package indexer

import (
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
)

type conversionContext struct {
	blockHeight       uint64
	prevStateReadOnly *appstate.AppState
	newStateReadOnly  *appstate.AppState
}

type conversionCollector struct {
	submittedFlips []db.Flip
	deletedFlips   []db.DeletedFlip
	flipTxs        []flipTx
	flipKeys       []db.FlipKey
	flipsWords     []db.FlipWords
	addresses      map[string]*db.Address
}

func (collector *conversionCollector) getAddresses() []db.Address {
	var addresses []db.Address
	for _, addr := range collector.addresses {
		addresses = append(addresses, *addr)
	}
	return addresses
}
