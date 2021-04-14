package online

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/log"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"time"
)

type Identity struct {
	Address      string
	LastActivity *time.Time
	Penalty      decimal.Decimal
	Online       bool
}

type CurrentOnlineIdentitiesHolder interface {
	GetAll() []*Identity
	Get(address string) *Identity
	GetOnlineCount() int
}

type currentOnlineIdentitiesCache struct {
	identities           []*Identity
	identitiesPerAddress map[string]*Identity
	onlineCount          int
}

func NewCurrentOnlineIdentitiesCache(appState *appstate.AppState,
	chain *blockchain.Blockchain,
	offlineDetector *blockchain.OfflineDetector) CurrentOnlineIdentitiesHolder {
	cache := &currentOnlineIdentitiesCache{}
	cache.set(nil, make(map[string]*Identity), 0)
	cache.initialize(appState, chain, offlineDetector)
	return cache
}

type currentOnlineIdentitiesCacheUpdater struct {
	log             log.Logger
	cache           *currentOnlineIdentitiesCache
	appState        *appstate.AppState
	chain           *blockchain.Blockchain
	offlineDetector *blockchain.OfflineDetector
}

func (cache *currentOnlineIdentitiesCache) GetAll() []*Identity {
	return cache.identities
}

func (cache *currentOnlineIdentitiesCache) Get(address string) *Identity {
	return cache.identitiesPerAddress[strings.ToLower(address)]
}

func (cache *currentOnlineIdentitiesCache) GetOnlineCount() int {
	return cache.onlineCount
}

func (cache *currentOnlineIdentitiesCache) set(
	identities []*Identity,
	identitiesPerAddress map[string]*Identity,
	onlineCount int,
) {
	cache.identities = identities
	cache.identitiesPerAddress = identitiesPerAddress
	cache.onlineCount = onlineCount
}

func (cache *currentOnlineIdentitiesCache) initialize(appState *appstate.AppState,
	chain *blockchain.Blockchain,
	offlineDetector *blockchain.OfflineDetector) {
	updater := currentOnlineIdentitiesCacheUpdater{
		log:             log.New("component", "currentOnlineIdentitiesCacheUpdater"),
		cache:           cache,
		appState:        appState,
		chain:           chain,
		offlineDetector: offlineDetector,
	}
	go updater.loop()
}

func (updater *currentOnlineIdentitiesCacheUpdater) loop() {
	for {
		updater.update()
		time.Sleep(time.Second * 60)
	}
}

func (updater *currentOnlineIdentitiesCacheUpdater) update() {
	startTime := time.Now()
	var identities []*Identity
	identitiesPerAddress := make(map[string]*Identity)
	if updater.chain.Head == nil {
		updater.log.Error("Unable to update due to empty chain head")
		return
	}
	height := updater.chain.Head.Height()
	appState, err := updater.appState.Readonly(height)
	if err != nil {
		updater.log.Error(fmt.Sprintf("Unable to update, height %v, err %v", height, err.Error()))
		return
	}

	activityMap := updater.offlineDetector.GetActivityMap()
	var onlineCount int
	appState.State.IterateOverIdentities(func(address common.Address, identity state.Identity) {
		online := appState.ValidatorsCache.IsOnlineIdentity(address)
		if online {
			if appState.ValidatorsCache.IsPool(address) {
				onlineCount += appState.ValidatorsCache.PoolSize(address)
			} else {
				onlineCount++
			}
		}
		if identity.State != state.Newbie && identity.State != state.Verified && identity.State != state.Human {
			return
		}
		addressStr := conversion.ConvertAddress(address)
		var lastActivity *time.Time
		if t, present := activityMap[address]; present {
			lastActivity = &t
		}
		onlineIdentity := &Identity{
			Address:      addressStr,
			LastActivity: lastActivity,
			Penalty:      blockchain.ConvertToFloat(identity.Penalty),
			Online:       online,
		}
		identities = append(identities, onlineIdentity)
		identitiesPerAddress[strings.ToLower(addressStr)] = onlineIdentity
	})

	if len(identities) > 0 {
		sort.Slice(identities, func(i, j int) bool {
			jTime := identities[j]
			if jTime.LastActivity == nil {
				return true
			}
			iTime := identities[i]
			if iTime.LastActivity == nil {
				return false
			}
			return iTime.LastActivity.After(*jTime.LastActivity)
		})
	}

	updater.cache.set(identities, identitiesPerAddress, onlineCount)
	finishTime := time.Now()
	updater.log.Debug("Updated", "duration", finishTime.Sub(startTime))
}
