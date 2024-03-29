package incoming

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/core/mempool"
	"github.com/idena-network/idena-go/core/upgrade"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-indexer/monitoring"
)

type Listener interface {
	Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64)
	AppStateReadonly(height uint64) (*appstate.AppState, error)
	AppState() *appstate.AppState
	NodeCtx() *node.NodeCtx
	StatsCollector() collector.StatsCollector
	Flipper() *flip.Flipper
	Config() *config.Config
	KeysPool() *mempool.KeysPool
	NodeEventBus() eventbus.Bus
	Destroy()
	WaitForStop()
}

type listenerImpl struct {
	bus             eventbus.Bus
	appState        *appstate.AppState
	nodeCtx         *node.NodeCtx
	statsCollector  collector.StatsCollector
	blockchain      *blockchain.Blockchain
	flipper         *flip.Flipper
	keysPool        *mempool.KeysPool
	offlineDetector *blockchain.OfflineDetector
	upgrader        *upgrade.Upgrader
	config          *config.Config
	node            *node.Node
	handleBlock     func(block *types.Block)
}

func NewListener(cfg *config.Config, bus eventbus.Bus, statsCollector collector.StatsCollector, pm monitoring.PerformanceMonitor) Listener {
	l := &listenerImpl{}

	pm.Start("Full")
	pm.Start("Node")

	bus.Subscribe(events.AddBlockEventID,
		func(e eventbus.Event) {
			pm.Complete("Node")
			pm.Start("Index")

			newBlockEvent := e.(*events.NewBlockEvent)
			l.handleBlock(newBlockEvent.Block)

			pm.Complete("Index")
			pm.Complete("Full")
			pm.Start("Full")
			pm.Start("Node")
		})

	nodeCtx, err := node.NewNodeWithInjections(cfg, bus, statsCollector, "0.0.1")
	if err != nil {
		panic(err)
	}

	l.bus = bus
	l.appState = nodeCtx.AppState
	l.flipper = nodeCtx.Flipper
	l.blockchain = nodeCtx.Blockchain
	l.keysPool = nodeCtx.KeysPool
	l.offlineDetector = nodeCtx.OfflineDetector
	l.upgrader = nodeCtx.Upgrader
	l.config = cfg
	l.nodeCtx = nodeCtx
	l.statsCollector = statsCollector

	l.node = nodeCtx.Node

	return l
}

func (l *listenerImpl) AppStateReadonly(height uint64) (*appstate.AppState, error) {
	return l.appState.Readonly(height)
}

func (l *listenerImpl) AppState() *appstate.AppState {
	return l.appState
}

func (l *listenerImpl) NodeCtx() *node.NodeCtx {
	return l.nodeCtx
}

func (l *listenerImpl) StatsCollector() collector.StatsCollector {
	return l.statsCollector
}

func (l *listenerImpl) Flipper() *flip.Flipper {
	return l.flipper
}

func (l *listenerImpl) KeysPool() *mempool.KeysPool {
	return l.keysPool
}

func (l *listenerImpl) Config() *config.Config {
	return l.config
}

func (l *listenerImpl) Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64) {
	l.handleBlock = handleBlock
	l.node.StartWithHeight(expectedHeadHeight)
}

func (l *listenerImpl) NodeEventBus() eventbus.Bus {
	return l.bus
}

func (l *listenerImpl) WaitForStop() {
	l.node.WaitForStop()
}

func (l *listenerImpl) Destroy() {

}
