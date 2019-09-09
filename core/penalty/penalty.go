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

func NewCurrentPenaltiesCache(appState *appstate.AppState) CurrentPenaltiesHolder {
	cache := &currentPenaltiesCache{}
	cache.initialize(appState)
	return cache
}

type currentPenaltiesCache struct {
	penalties           []*Penalty
	penaltiesPerAddress map[string]*Penalty
}

type currentPenaltiesCacheUpdater struct {
	cache    *currentPenaltiesCache
	appState *appstate.AppState
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

func (cache *currentPenaltiesCache) initialize(appState *appstate.AppState) {
	updater := currentPenaltiesCacheUpdater{
		cache:    cache,
		appState: appState,
	}
	go updater.loop()
}

func (updater *currentPenaltiesCacheUpdater) loop() {
	for {
		time.Sleep(time.Second * 5)

		var penalties []*Penalty
		penaltiesPerAddress := make(map[string]*Penalty)
		updater.appState.State.IterateOverIdentities(func(address common.Address, identity state.Identity) {
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

		sort.Slice(penalties, func(i, j int) bool {
			return penalties[i].Penalty.Cmp(penalties[j].Penalty) == 1
		})

		updater.cache.set(penalties, penaltiesPerAddress)
	}
}
