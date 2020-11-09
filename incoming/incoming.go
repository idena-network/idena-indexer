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
	"github.com/idena-network/idena-go/database"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/log"
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

func NewListener(nodeConfigFile string, pm monitoring.PerformanceMonitor) Listener {
	l := &listenerImpl{}

	cfg, err := config.MakeConfigFromFile(nodeConfigFile)
	if err != nil {
		panic(err)
	}
	cfg.P2P.MaxInboundPeers = config.LowPowerMaxInboundPeers
	cfg.P2P.MaxOutboundPeers = config.LowPowerMaxOutboundPeers
	cfg.IpfsConf.LowWater = 8
	cfg.IpfsConf.HighWater = 10
	cfg.IpfsConf.GracePeriod = "30s"
	cfg.IpfsConf.ReproviderInterval = "0"
	cfg.IpfsConf.Routing = "dhtclient"
	cfg.Sync.FastSync = false

	cfgTransform(cfg)

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

func cfgTransform(cfg *config.Config) {
	db, err := node.OpenDatabase(cfg.DataDir, "idenachain", 16, 16)
	if err != nil {
		log.Error("Cannot transform consensus config", "err", err)
		return
	}
	defer db.Close()
	repo := database.NewRepo(db)
	consVersion := repo.ReadConsensusVersion()
	if consVersion <= uint32(cfg.Consensus.Version) {
		return
	}
	config.ApplyConsensusVersion(config.ConsensusVerson(consVersion), cfg.Consensus)
	log.Info("Consensus config transformed to", "ver", consVersion)
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
