package mempool

import (
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/patrickmn/go-cache"
	"time"
)

const period = time.Hour

type PeersTracker interface {
	AddPeersData(peersData []iface.ConnectionInfo, time time.Time)
}

func NewPeersTracker(dbAccessor db.Accessor, logger log.Logger) PeersTracker {
	pt := &peersTracker{
		dbAccessor:     dbAccessor,
		peersDataQueue: make(chan *peersDataWrapper, 10),
		peersData:      cache.New(period, period*2),
		logger:         logger,
	}
	go pt.handlePeers()
	go pt.track()
	return pt
}

type peersTracker struct {
	dbAccessor     db.Accessor
	peersDataQueue chan *peersDataWrapper
	peersData      *cache.Cache
	logger         log.Logger
}

type peersDataWrapper struct {
	peersData []iface.ConnectionInfo
	timestamp time.Time
}

type peerData struct {
}

func (pt *peersTracker) AddPeersData(peersData []iface.ConnectionInfo, timestamp time.Time) {
	wrapper := &peersDataWrapper{
		peersData: peersData,
		timestamp: timestamp,
	}
	pt.peersDataQueue <- wrapper
}

func (pt *peersTracker) handlePeers() {
	for {
		wrapper := <-pt.peersDataQueue
		if len(wrapper.peersData) == 0 {
			continue
		}
		for _, info := range wrapper.peersData {
			pt.peersData.Set(info.ID().String(), peerData{}, cache.DefaultExpiration)
		}
	}
}

func (pt *peersTracker) track() {
	time.Sleep(period)
	for {
		pt.peersData.DeleteExpired()
		now := time.Now()
		n := pt.peersData.ItemCount()
		if err := pt.dbAccessor.SavePeersCount(n, now); err != nil {
			pt.logger.Warn("Failed to track peers count", "n", n, "err", err)
			time.Sleep(time.Minute)
		} else {
			pt.logger.Debug("Tracked peers count", "n", n)
			time.Sleep(period)
		}
	}
}
