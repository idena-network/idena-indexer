package incoming

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/events"
	"github.com/idena-network/idena-go/node"
	"gopkg.in/urfave/cli.v1"
)

type Listener interface {
	Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64)
	Node() *node.Node
	Destroy()
}

type listenerImpl struct {
	//nodeConfigFile string
	appContext *cli.Context
	n          *node.Node
}

func NewListener(appContext *cli.Context) Listener {
	l := &listenerImpl{
		appContext: appContext,
	}
	return l
}

func (l *listenerImpl) Listen(handleBlock func(block *types.Block), expectedHeadHeight uint64) {
	cfg, err := config.MakeConfig(l.appContext)
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
