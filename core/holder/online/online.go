package online

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/types"
	"github.com/idena-network/idena-indexer/log"
	"github.com/shopspring/decimal"
	"math"
	"sort"
	"strings"
	"time"
)

type Identity struct {
	Address      string
	LastActivity *time.Time
	Penalty      decimal.Decimal
	Online       bool
	Delegatee    *Identity
}

type CurrentOnlineIdentitiesHolder interface {
	GetAll() []*Identity
	Get(address string) *Identity
	GetOnlineCount() int
	ValidatorsCount() int
	Validators() []*types.Validator
	OnlineValidatorsCount() int
	OnlineValidators() []*types.Validator
	Staking() float64
}

type currentOnlineIdentitiesCache struct {
	identities           []*Identity
	identitiesPerAddress map[string]*Identity
	onlineCount          int
	validators           []*types.Validator
	onlineValidators     []*types.Validator
	staking              float64
}

func NewCurrentOnlineIdentitiesCache(appState *appstate.AppState,
	chain *blockchain.Blockchain,
	offlineDetector *blockchain.OfflineDetector) CurrentOnlineIdentitiesHolder {
	cache := &currentOnlineIdentitiesCache{}
	cache.set(nil, make(map[string]*Identity), 0, nil, nil, 0)
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

func (cache *currentOnlineIdentitiesCache) ValidatorsCount() int {
	return len(cache.validators)
}

func (cache *currentOnlineIdentitiesCache) Validators() []*types.Validator {
	return cache.validators
}

func (cache *currentOnlineIdentitiesCache) OnlineValidatorsCount() int {
	return len(cache.onlineValidators)
}

func (cache *currentOnlineIdentitiesCache) OnlineValidators() []*types.Validator {
	return cache.onlineValidators
}

func (cache *currentOnlineIdentitiesCache) Staking() float64 {
	return cache.staking
}

func (cache *currentOnlineIdentitiesCache) set(
	identities []*Identity,
	identitiesPerAddress map[string]*Identity,
	onlineCount int,
	validators []*types.Validator,
	onlineValidators []*types.Validator,
	staking float64,
) {
	cache.identities = identities
	cache.identitiesPerAddress = identitiesPerAddress
	cache.onlineCount = onlineCount
	cache.validators = validators
	cache.onlineValidators = onlineValidators
	cache.staking = staking
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
	poolsByAddress := make(map[common.Address]*Identity)
	var validators, onlineValidators []*types.Validator

	buildIdentity := func(address common.Address, identity state.Identity, pOnline *bool, delegetee *Identity) *Identity {
		var online bool
		if pOnline != nil {
			online = *pOnline
		} else {
			online = appState.ValidatorsCache.IsOnlineIdentity(address)
		}
		addressStr := conversion.ConvertAddress(address)
		var lastActivity *time.Time
		if t, present := activityMap[address]; present {
			lastActivity = &t
		}
		return &Identity{
			Address:      addressStr,
			LastActivity: lastActivity,
			Penalty:      blockchain.ConvertToFloat(identity.Penalty),
			Online:       online,
			Delegatee:    delegetee,
		}
	}

	buildValidator := func(address common.Address, identity state.Identity) *types.Validator {
		var size uint32
		isPool := appState.ValidatorsCache.IsPool(address)
		if identity.State.NewbieOrBetter() && identity.Delegatee == nil && !isPool {
			size = 1
		} else if isPool {
			size = uint32(appState.ValidatorsCache.PoolSize(address))
		}
		if size == 0 {
			return nil
		}
		var lastActivity *time.Time
		if t, present := activityMap[address]; present {
			lastActivity = &t
		}
		return &types.Validator{
			Address:      conversion.ConvertAddress(address),
			Size:         size,
			Online:       appState.ValidatorsCache.IsOnlineIdentity(address),
			LastActivity: lastActivity,
			Penalty:      blockchain.ConvertToFloat(identity.Penalty),
			IsPool:       isPool,
		}
	}

	var staking float64
	appState.State.IterateOverIdentities(func(address common.Address, identity state.Identity) {
		if identity.State.NewbieOrBetter() && identity.Stake != nil {
			stake, _ := blockchain.ConvertToFloat(identity.Stake).Float64()
			staking += math.Pow(stake, 0.9)
		}
		if validator := buildValidator(address, identity); validator != nil {
			validators = append(validators, validator)
			if validator.Online {
				onlineValidators = append(onlineValidators, validator)
			}
		}
		var delegatee *Identity
		if identity.Delegatee != nil {
			delegeteeAddr := *identity.Delegatee
			if delegatee = poolsByAddress[delegeteeAddr]; delegatee == nil {
				delegatee = buildIdentity(delegeteeAddr, appState.State.GetIdentity(delegeteeAddr), nil, nil)
				poolsByAddress[delegeteeAddr] = delegatee
			}
		}
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
		var onlineIdentity *Identity
		if onlineIdentity = poolsByAddress[address]; onlineIdentity == nil {
			onlineIdentity = buildIdentity(address, identity, &online, delegatee)
		}
		identities = append(identities, onlineIdentity)
		identitiesPerAddress[strings.ToLower(onlineIdentity.Address)] = onlineIdentity
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
	if len(validators) > 0 {
		sort.Slice(validators, func(i, j int) bool {
			return validators[i].Size > validators[j].Size
		})
	}
	if len(onlineValidators) > 0 {
		sort.Slice(onlineValidators, func(i, j int) bool {
			return onlineValidators[i].Size > onlineValidators[j].Size
		})
	}

	updater.cache.set(identities, identitiesPerAddress, onlineCount, validators, onlineValidators, staking)
	finishTime := time.Now()
	updater.log.Debug("Updated", "duration", finishTime.Sub(startTime))
}
