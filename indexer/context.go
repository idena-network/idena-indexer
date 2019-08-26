package indexer

import (
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

type conversionContext struct {
	blockHeight       uint64
	submittedFlips    []db.Flip
	flipKeys          []db.FlipKey
	flipsData         []db.FlipData
	addresses         map[string]*db.Address
	balanceUpdates    []db.Balance
	totalBalanceDiff  *balanceDiff
	totalFee          *big.Int
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
