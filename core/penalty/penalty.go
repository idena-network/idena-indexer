package penalty

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"time"
)

type Penalty struct {
	Address string
	Penalty decimal.Decimal
}

type CurrentPenaltiesHolder interface {
	GetAll() []*Penalty
	Get(address string) *Penalty
}

func NewCurrentPenaltiesCache(appState *appstate.AppState, chain *blockchain.Blockchain) CurrentPenaltiesHolder {
	cache := &currentPenaltiesCache{}
	cache.set(nil, make(map[string]*Penalty))
	cache.initialize(appState, chain)
	return cache
}

type currentPenaltiesCache struct {
	penalties           []*Penalty
	penaltiesPerAddress map[string]*Penalty
}

type currentPenaltiesCacheUpdater struct {
	cache    *currentPenaltiesCache
	appState *appstate.AppState
	chain    *blockchain.Blockchain
}

func (cache *currentPenaltiesCache) GetAll() []*Penalty {
	return cache.penalties
}

func (cache *currentPenaltiesCache) Get(address string) *Penalty {
	return cache.penaltiesPerAddress[strings.ToLower(address)]
}

func (cache *currentPenaltiesCache) set(penalties []*Penalty, penaltiesPerAddress map[string]*Penalty) {
	cache.penalties = penalties
	cache.penaltiesPerAddress = penaltiesPerAddress
}

func (cache *currentPenaltiesCache) initialize(appState *appstate.AppState, chain *blockchain.Blockchain) {
	updater := currentPenaltiesCacheUpdater{
		cache:    cache,
		appState: appState,
		chain:    chain,
	}
	go updater.loop()
}

func (updater *currentPenaltiesCacheUpdater) loop() {
	for {
		time.Sleep(time.Second * 5)

		var penalties []*Penalty
		penaltiesPerAddress := make(map[string]*Penalty)
		if updater.chain.Head != nil {
			appState := updater.appState.Readonly(updater.chain.Head.Height())
			if appState != nil {
				appState.State.IterateOverIdentities(func(address common.Address, identity state.Identity) {
					addressStr := indexer.ConvertAddress(address)
					if identity.Penalty == nil || identity.Penalty.Sign() != 1 {
						return
					}
					penalty := &Penalty{
						Address: addressStr,
						Penalty: blockchain.ConvertToFloat(identity.Penalty),
					}
					penalties = append(penalties, penalty)
					penaltiesPerAddress[strings.ToLower(addressStr)] = penalty
				})
			}
		}

		if len(penalties) > 0 {
			sort.Slice(penalties, func(i, j int) bool {
				return penalties[i].Penalty.Cmp(penalties[j].Penalty) == 1
			})
		}

		updater.cache.set(penalties, penaltiesPerAddress)
	}
}
