package indexer

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-indexer/db"
)

type conversionContext struct {
	blockHeight       uint64
	prevStateReadOnly *appstate.AppState
	newStateReadOnly  *appstate.AppState
}

type conversionCollector struct {
	submittedFlips      []db.Flip
	deletedFlips        []db.DeletedFlip
	flipTxs             []flipTx
	flipKeys            []db.FlipKey
	flipsWords          []db.FlipWords
	addresses           map[string]*db.Address
	killInviteeTxs      []db.KillInviteeTx
	becomeOnlineTxs     []string
	becomeOfflineTxs    []string
	killedAddrs         map[common.Address]struct{}
	switchDelegationTxs []*types.Transaction
}

func (collector *conversionCollector) getAddresses() []db.Address {
	addresses := make([]db.Address, 0, len(collector.addresses))
	for _, addr := range collector.addresses {
		addresses = append(addresses, *addr)
	}
	return addresses
}

type delegateeEpochRewardsWrapper struct {
	rewards            []db.DelegateeEpochRewards
	indexesByDelegatee map[common.Address]int
}

func newDelegateeEpochRewardsWrapper(capacity int) *delegateeEpochRewardsWrapper {
	return &delegateeEpochRewardsWrapper{
		rewards:            make([]db.DelegateeEpochRewards, 0, capacity),
		indexesByDelegatee: make(map[common.Address]int, capacity),
	}
}

func (w *delegateeEpochRewardsWrapper) append(item db.DelegateeEpochRewards) {
	w.indexesByDelegatee[item.Address] = len(w.rewards)
	w.rewards = append(w.rewards, item)
}

func (w *delegateeEpochRewardsWrapper) incPenalizedDelegators(delegatee common.Address) {
	if i, ok := w.indexesByDelegatee[delegatee]; ok {
		w.rewards[i].PenalizedDelegators++
	} else {
		w.append(db.DelegateeEpochRewards{
			Address:             delegatee,
			PenalizedDelegators: 1,
		})
	}
}
