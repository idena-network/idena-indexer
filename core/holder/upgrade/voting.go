package upgrade

import (
	"github.com/idena-network/idena-go/core/upgrade"
	"github.com/idena-network/idena-indexer/log"
	"time"
)

type UpgradesVotingHolder interface {
	Get() []*Votes
}

type upgradesVotingCache struct {
	votes []*Votes
}

type Votes struct {
	Upgrade uint32
	Votes   uint64
}

func NewUpgradesVotingHolder(upgrader *upgrade.Upgrader) UpgradesVotingHolder {
	holder := &upgradesVotingCache{}
	holder.initialize(upgrader)
	return holder
}

func (cache *upgradesVotingCache) initialize(upgrader *upgrade.Upgrader) {
	updater := upgradesVotingCacheUpdater{
		log:      log.New("component", "upgradesVotingCacheUpdater"),
		cache:    cache,
		upgrader: upgrader,
	}
	go updater.loop()
}

func (cache *upgradesVotingCache) Get() []*Votes {
	return cache.votes
}

type upgradesVotingCacheUpdater struct {
	log      log.Logger
	cache    *upgradesVotingCache
	upgrader *upgrade.Upgrader
}

func (updater *upgradesVotingCacheUpdater) loop() {
	for {
		updater.update()
		time.Sleep(time.Second * 10)
	}
}

func (updater *upgradesVotingCacheUpdater) update() {
	startTime := time.Now()
	var resVotes []*Votes
	votes := updater.upgrader.GetVotes()
	if len(votes) > 0 {
		votesByUpgrade := make(map[uint32]uint64)
		for _, u := range votes {
			votesByUpgrade[u]++
		}
		resVotes = make([]*Votes, len(votesByUpgrade))
		i := 0
		for u, votes := range votesByUpgrade {
			resVotes[i] = &Votes{
				Upgrade: u,
				Votes:   votes,
			}
			i++
		}
	}
	updater.cache.votes = resVotes
	finishTime := time.Now()
	updater.log.Debug("Updated", "duration", finishTime.Sub(startTime))
}
