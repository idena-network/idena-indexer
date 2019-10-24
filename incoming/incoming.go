package incoming

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/core/mempool"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/monitoring"
)

type Listener interface {
	Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64)
	AppStateReadonly(height uint64) *appstate.AppState
	AppState() *appstate.AppState
	NodeCtx() *node.NodeCtx
	StatsCollector() collector.StatsCollector
	Blockchain() *blockchain.Blockchain
	Flipper() *flip.Flipper
	Config() *config.Config
	KeysPool() *mempool.KeysPool
	OfflineDetector() *blockchain.OfflineDetector
	Destroy()
	WaitForStop()
}

type listenerImpl struct {
	appState        *appstate.AppState
	nodeCtx         *node.NodeCtx
	statsCollector  collector.StatsCollector
	blockchain      *blockchain.Blockchain
	flipper         *flip.Flipper
	keysPool        *mempool.KeysPool
	offlineDetector *blockchain.OfflineDetector
	config          *config.Config
	node            *node.Node
	handleBlock     func(block *types.Block)
}

func NewListener(nodeConfigFile string, pm monitoring.PerformanceMonitor) Listener {
	l := &listenerImpl{}

	cfg, err := config.MakeConfigFromFile(nodeConfigFile)
	if err != nil {
		panic(err)
	}
	cfg.P2P.MaxPeers = config.LowPowerMaxPeers
	cfg.IpfsConf.LowWater = 50
	cfg.IpfsConf.HighWater = 100
	cfg.IpfsConf.GracePeriod = "20s"
	cfg.IpfsConf.ReproviderInterval = "12h"
	cfg.IpfsConf.Routing = "dht"
	cfg.Sync.FastSync = false

	bus := eventbus.New()

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

	statsCollector := stats.NewStatsCollector()
	nodeCtx, err := node.NewNodeWithInjections(cfg, bus, statsCollector, "0.0.1")
	if err != nil {
		panic(err)
	}

	l.appState = nodeCtx.AppState
	l.flipper = nodeCtx.Flipper
	l.blockchain = nodeCtx.Blockchain
	l.keysPool = nodeCtx.KeysPool
	l.offlineDetector = nodeCtx.OfflineDetector
	l.config = cfg
	l.nodeCtx = nodeCtx
	l.statsCollector = statsCollector

	l.node = nodeCtx.Node

	return l
}

func (l *listenerImpl) AppStateReadonly(height uint64) *appstate.AppState {
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

func (l *listenerImpl) Blockchain() *blockchain.Blockchain {
	return l.blockchain
}

func (l *listenerImpl) Flipper() *flip.Flipper {
	return l.flipper
}

func (l *listenerImpl) KeysPool() *mempool.KeysPool {
	return l.keysPool
}

func (l *listenerImpl) OfflineDetector() *blockchain.OfflineDetector {
	return l.offlineDetector
}

func (l *listenerImpl) Config() *config.Config {
	return l.config
}

func (l *listenerImpl) Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64) {
	l.handleBlock = handleBlock
	l.node.StartWithHeight(expectedHeadHeight)
}

func (l *listenerImpl) WaitForStop() {
	l.node.WaitForStop()
}

func (l *listenerImpl) Destroy() {

}
