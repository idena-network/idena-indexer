package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
)

const (
	RemovedMemPoolTxEventID = eventbus.EventID("removed-mem-pool-tx")
)

type RemovedMemPoolTxEvent struct {
	Tx *types.Transaction
}

func (e *RemovedMemPoolTxEvent) EventID() eventbus.EventID {
	return RemovedMemPoolTxEventID
}
