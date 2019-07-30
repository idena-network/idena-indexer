package incoming

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
)

type Listener interface {
	Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64)
	Node() *node.Node
	Destroy()
}

type listenerImpl struct {
	nodeConfigFile string
	n              *node.Node
}

func NewListener(nodeConfigFile string) Listener {
	l := &listenerImpl{
		nodeConfigFile: nodeConfigFile,
	}
	return l
}

func (l *listenerImpl) Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64) {
	cfg, err := config.MakeConfigFromFile(l.nodeConfigFile)
	if err != nil {
		panic(err)
	}
	cfg.Sync.FastSync = false

	bus := eventbus.New()
	bus.Subscribe(events.AddBlockEventID,
		func(e eventbus.Event) {
			newBlockEvent := e.(*events.NewBlockEvent)
			handleBlock(newBlockEvent.Block)
		})

	n, err := node.NewNode(cfg, bus)
	if err != nil {
		panic(err)
	}

	l.n = n

	n.StartWithHeight(expectedHeadHeight)
	n.WaitForStop()
}

func (l *listenerImpl) Node() *node.Node {
	return l.n
}

func (l *listenerImpl) Destroy() {

}
