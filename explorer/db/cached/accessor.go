package cached

import (
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	permanentDataLifeTime                   = time.Hour * 2
	activeAddressesCountMethod              = "ActiveAddressesCount"
	epochFlipAnswersSummaryMethod           = "EpochFlipAnswersSummary"
	epochFlipStatesSummaryMethod            = "EpochFlipStatesSummary"
	epochFlipWrongWordsSummaryMethod        = "EpochFlipWrongWordsSummary"
	epochIdentityStatesSummaryMethod        = "EpochIdentityStatesSummary"
	epochIdentityStatesInterimSummaryMethod = "EpochIdentityStatesInterimSummary"
	epochInvitesSummaryMethod               = "EpochInvitesSummary"
	epochInviteStatesSummaryMethod          = "EpochInviteStatesSummary"
	epochRewardsSummaryMethod               = "EpochRewardsSummary"
	epochBadAuthorsCountMethod              = "EpochBadAuthorsCount"
	epochBadAuthorsMethod                   = "EpochBadAuthors"
	epochBadAuthorsOldMethod                = "EpochBadAuthorsOld"
	epochRewardsCountMethod                 = "EpochRewardsCount"
	epochRewardsMethod                      = "EpochRewards"
	epochIdentitiesRewardsCountMethod       = "EpochIdentitiesRewardsCount"
	epochIdentitiesRewardsMethod            = "EpochIdentitiesRewards"
	epochIdentitiesRewardsOldMethod         = "EpochIdentitiesRewardsOld"
	epochFundPaymentsMethod                 = "EpochFundPayments"
	flipEpochAdjacentFlipsMethod            = "FlipEpochAdjacentFlips"
	flipAddressAdjacentFlipsMethod          = "FlipAddressAdjacentFlips"
	flipEpochIdentityAdjacentFlipsMethod    = "FlipEpochIdentityAdjacentFlips"
)

type cachedAccessor struct {
	db.Accessor
	maxItemCountsByMethod    map[string]int
	defaultCacheMaxItemCount int
	maxItemLifeTimesByMethod map[string]time.Duration
	defaultCacheItemLifeTime time.Duration
	cachesByMethod           map[string]Cache
	mutex                    sync.Mutex
	logger                   log.Logger
}

func NewCachedAccessor(
	db db.Accessor,
	defaultCacheMaxItemCount int,
	defaultCacheItemLifeTime time.Duration,
	logger log.Logger,
) db.Accessor {
	a := &cachedAccessor{
		Accessor:                 db,
		maxItemCountsByMethod:    createMaxItemCountsByMethod(),
		defaultCacheMaxItemCount: defaultCacheMaxItemCount,
		maxItemLifeTimesByMethod: createMaxItemLifeTimesByMethod(),
		defaultCacheItemLifeTime: defaultCacheItemLifeTime,
		logger:                   logger,
	}
	go func() {
		for {
			time.Sleep(time.Minute)
			a.log()
		}
	}()
	go a.monitorEpochChange()
	return a
}

func createMaxItemCountsByMethod() map[string]int {
	return map[string]int{
		activeAddressesCountMethod: 1,
	}
}

func createMaxItemLifeTimesByMethod() map[string]time.Duration {
	return map[string]time.Duration{
		activeAddressesCountMethod:              time.Minute * 5,
		epochIdentityStatesInterimSummaryMethod: time.Minute * 5,
		epochInvitesSummaryMethod:               time.Minute * 3,
		epochInviteStatesSummaryMethod:          time.Minute * 3,
		flipEpochAdjacentFlipsMethod:            time.Minute * 20,
		flipAddressAdjacentFlipsMethod:          time.Minute * 20,
		flipEpochIdentityAdjacentFlipsMethod:    time.Minute * 20,
		epochFlipAnswersSummaryMethod:           permanentDataLifeTime,
		epochFlipStatesSummaryMethod:            permanentDataLifeTime,
		epochFlipWrongWordsSummaryMethod:        permanentDataLifeTime,
		epochIdentityStatesSummaryMethod:        permanentDataLifeTime,
		epochRewardsSummaryMethod:               permanentDataLifeTime,
		epochBadAuthorsCountMethod:              permanentDataLifeTime,
		epochBadAuthorsMethod:                   permanentDataLifeTime,
		epochBadAuthorsOldMethod:                permanentDataLifeTime,
		epochRewardsCountMethod:                 permanentDataLifeTime,
		epochRewardsMethod:                      permanentDataLifeTime,
		epochIdentitiesRewardsCountMethod:       permanentDataLifeTime,
		epochIdentitiesRewardsMethod:            permanentDataLifeTime,
		epochIdentitiesRewardsOldMethod:         permanentDataLifeTime,
		epochFundPaymentsMethod:                 permanentDataLifeTime,
	}
}

func (a *cachedAccessor) monitorEpochChange() {
	isFirst := true
	epoch := uint64(0)
	const delay = time.Second * 5
	for {
		time.Sleep(delay)
		lastEpoch, err := a.Accessor.LastEpoch()
		if err != nil {
			a.logger.Warn(errors.Wrap(err, "Unable to get last epoch from db to detect new one").Error())
			continue
		}
		a.logger.Debug(fmt.Sprintf("epoch: %v, lastEpoch: %v", epoch, lastEpoch.Epoch))
		if lastEpoch.Epoch > epoch {
			epoch = lastEpoch.Epoch
			if isFirst {
				isFirst = false
			} else {
				a.logger.Debug("Detected new epoch")
				a.clearCache()
			}
		}
		timeToStartMonitoring := lastEpoch.ValidationTime.Add(time.Minute * 25)
		now := time.Now()
		if timeToStartMonitoring.After(now) {
			<-time.After(timeToStartMonitoring.Sub(now))
		}
	}
}

func (a *cachedAccessor) clearCache() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for method, dbCache := range a.cachesByMethod {
		dbCache.Clear()
		a.logger.Debug(fmt.Sprintf("Cleared %v cache", method))
	}
}

func (a *cachedAccessor) log() {
	type methodItemsCount struct {
		method string
		count  int
	}
	a.mutex.Lock()
	itemsCounts := make([]*methodItemsCount, len(a.cachesByMethod))
	i := 0
	for method, dbCache := range a.cachesByMethod {
		itemsCounts[i] = &methodItemsCount{
			method: method,
			count:  dbCache.ItemsCount(),
		}
		i++
	}
	a.mutex.Unlock()

	emptyCount := 0
	var s []string
	if len(itemsCounts) > 0 {
		sort.Slice(itemsCounts, func(i, j int) bool {
			return itemsCounts[i].count >= itemsCounts[j].count
		})
		for _, itemsCount := range itemsCounts {
			if itemsCount.count > 0 {
				s = append(s, fmt.Sprintf("%s: %d", itemsCount.method, itemsCount.count))
			} else {
				emptyCount++
			}
		}
	}
	header := fmt.Sprintf("Total: %d, empty: %d", len(itemsCounts), emptyCount)
	var infoToLog string
	if len(s) > 0 {
		infoToLog = fmt.Sprintf("%s (%s)", header, strings.Join(s, ", "))
	} else {
		infoToLog = header
	}
	a.logger.Debug(infoToLog)
}

type cachedValue struct {
	res               interface{}
	continuationToken *string
	err               error
}

func key(args ...interface{}) string {
	res := "key"
	for _, arg := range args {
		res = fmt.Sprintf("%s-%v", res, arg)
	}
	return res
}

func (a *cachedAccessor) getOrLoad(method string, load func() (interface{}, error), args ...interface{}) (interface{}, error) {
	dbCache := a.getCache(method)
	key := key(args)
	if v, ok := dbCache.Get(key); ok {
		return v.(*cachedValue).res, v.(*cachedValue).err
	}
	res, err := load()
	dbCache.Set(key, &cachedValue{
		res: res,
		err: err,
	}, cache.DefaultExpiration)
	return res, err
}

func (a *cachedAccessor) getOrLoadWithConToken(method string, load func() (interface{}, *string, error), args ...interface{}) (interface{}, *string, error) {
	dbCache := a.getCache(method)
	key := key(args)
	if v, ok := dbCache.Get(key); ok {
		return v.(*cachedValue).res, v.(*cachedValue).continuationToken, v.(*cachedValue).err
	}
	res, continuationToken, err := load()
	dbCache.Set(key, &cachedValue{
		res:               res,
		continuationToken: continuationToken,
		err:               err,
	}, cache.DefaultExpiration)
	return res, continuationToken, err
}

func (a *cachedAccessor) getCache(method string) Cache {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.cachesByMethod == nil {
		a.cachesByMethod = make(map[string]Cache)
	}
	dbCache, ok := a.cachesByMethod[method]
	if !ok {
		maxSize := a.defaultCacheMaxItemCount
		if count, ok := a.maxItemCountsByMethod[method]; ok {
			maxSize = count
		}
		defaultExpiration := a.defaultCacheItemLifeTime
		if lifeTime, ok := a.maxItemLifeTimesByMethod[method]; ok {
			defaultExpiration = lifeTime
		}
		dbCache = NewCache(
			maxSize,
			defaultExpiration,
			a.logger.New("component", fmt.Sprintf("cache-%s", method)),
		)
		a.cachesByMethod[method] = dbCache
	}
	return dbCache
}

func (a *cachedAccessor) Search(value string) ([]types.Entity, error) {
	res, err := a.getOrLoad("Search", func() (interface{}, error) {
		return a.Accessor.Search(value)
	}, value)
	return res.([]types.Entity), err
}

func (a *cachedAccessor) Coins() (types.AllCoins, error) {
	res, err := a.getOrLoad("Coins", func() (interface{}, error) {
		return a.Accessor.Coins()
	})
	return res.(types.AllCoins), err
}

func (a *cachedAccessor) CirculatingSupply(addressesToExclude []string) (decimal.Decimal, error) {
	res, err := a.getOrLoad("CirculatingSupply", func() (interface{}, error) {
		return a.Accessor.CirculatingSupply(addressesToExclude)
	})
	return res.(decimal.Decimal), err
}

func (a *cachedAccessor) ActiveAddressesCount(afterTime time.Time) (uint64, error) {
	res, err := a.getOrLoad(activeAddressesCountMethod, func() (interface{}, error) {
		return a.Accessor.ActiveAddressesCount(afterTime)
	})
	return res.(uint64), err
}

func (a *cachedAccessor) EpochsCount() (uint64, error) {
	res, err := a.getOrLoad("EpochsCount", func() (interface{}, error) {
		return a.Accessor.EpochsCount()
	})
	return res.(uint64), err
}

func (a *cachedAccessor) Epochs(count uint64, continuationToken *string) ([]types.EpochSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("Epochs", func() (interface{}, *string, error) {
		return a.Accessor.Epochs(count, continuationToken)
	}, count, continuationToken)
	return res.([]types.EpochSummary), nextContinuationToken, err
}

func (a *cachedAccessor) EpochsOld(startIndex uint64, count uint64) ([]types.EpochSummary, error) {
	res, err := a.getOrLoad("EpochsOld", func() (interface{}, error) {
		return a.Accessor.EpochsOld(startIndex, count)
	}, startIndex, count)
	return res.([]types.EpochSummary), err
}

func (a *cachedAccessor) LastEpoch() (types.EpochDetail, error) {
	res, err := a.getOrLoad("LastEpoch", func() (interface{}, error) {
		return a.Accessor.LastEpoch()
	})
	return res.(types.EpochDetail), err
}

func (a *cachedAccessor) Epoch(epoch uint64) (types.EpochDetail, error) {
	res, err := a.getOrLoad("Epoch", func() (interface{}, error) {
		return a.Accessor.Epoch(epoch)
	}, epoch)
	return res.(types.EpochDetail), err
}

func (a *cachedAccessor) EpochBlocksCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad("EpochBlocksCount", func() (interface{}, error) {
		return a.Accessor.EpochBlocksCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochBlocks(epoch uint64, count uint64, continuationToken *string) ([]types.BlockSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("EpochBlocks", func() (interface{}, *string, error) {
		return a.Accessor.EpochBlocks(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.BlockSummary), nextContinuationToken, err
}

func (a *cachedAccessor) EpochBlocksOld(epoch uint64, startIndex uint64, count uint64) ([]types.BlockSummary, error) {
	res, err := a.getOrLoad("EpochBlocksOld", func() (interface{}, error) {
		return a.Accessor.EpochBlocksOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.BlockSummary), err
}

func (a *cachedAccessor) EpochFlipsCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad("EpochFlipsCount", func() (interface{}, error) {
		return a.Accessor.EpochFlipsCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochFlips(epoch uint64, count uint64, continuationToken *string) ([]types.FlipSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("EpochFlips", func() (interface{}, *string, error) {
		return a.Accessor.EpochFlips(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.FlipSummary), nextContinuationToken, err
}

func (a *cachedAccessor) EpochFlipsOld(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	res, err := a.getOrLoad("EpochFlipsOld", func() (interface{}, error) {
		return a.Accessor.EpochFlipsOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.FlipSummary), err
}

func (a *cachedAccessor) EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad(epochFlipAnswersSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochFlipAnswersSummary(epoch)
	}, epoch)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad(epochFlipStatesSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochFlipStatesSummary(epoch)
	}, epoch)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochFlipWrongWordsSummary(epoch uint64) ([]types.NullableBoolValueCount, error) {
	res, err := a.getOrLoad(epochFlipWrongWordsSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochFlipWrongWordsSummary(epoch)
	}, epoch)
	return res.([]types.NullableBoolValueCount), err
}

func (a *cachedAccessor) EpochIdentitiesCount(epoch uint64, prevStates []string, states []string) (uint64, error) {
	res, err := a.getOrLoad("EpochIdentitiesCount", func() (interface{}, error) {
		return a.Accessor.EpochIdentitiesCount(epoch, prevStates, states)
	}, epoch, prevStates, states)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochIdentities(epoch uint64, prevStates []string, states []string, count uint64,
	continuationToken *string) ([]types.EpochIdentity, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("EpochIdentities", func() (interface{}, *string, error) {
		return a.Accessor.EpochIdentities(epoch, prevStates, states, count, continuationToken)
	}, epoch, prevStates, states, count, continuationToken)
	return res.([]types.EpochIdentity), nextContinuationToken, err
}

func (a *cachedAccessor) EpochIdentitiesOld(epoch uint64, prevStates []string, states []string, startIndex uint64,
	count uint64) ([]types.EpochIdentity, error) {
	res, err := a.getOrLoad("EpochIdentitiesOld", func() (interface{}, error) {
		return a.Accessor.EpochIdentitiesOld(epoch, prevStates, states, startIndex, count)
	}, epoch, prevStates, states, startIndex, count)
	return res.([]types.EpochIdentity), err
}

func (a *cachedAccessor) EpochIdentityStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad(epochIdentityStatesSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochIdentityStatesSummary(epoch)
	}, epoch)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochIdentityStatesInterimSummary(epoch uint64) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad(epochIdentityStatesInterimSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochIdentityStatesInterimSummary(epoch)
	}, epoch)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochInvitesSummary(epoch uint64) (types.InvitesSummary, error) {
	res, err := a.getOrLoad(epochInvitesSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochInvitesSummary(epoch)
	}, epoch)
	return res.(types.InvitesSummary), err
}

func (a *cachedAccessor) EpochInviteStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad(epochInviteStatesSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochInviteStatesSummary(epoch)
	}, epoch)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochInvitesCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad("EpochInvitesCount", func() (interface{}, error) {
		return a.Accessor.EpochInvitesCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochInvites(epoch uint64, count uint64, continuationToken *string) ([]types.Invite, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("EpochInvites", func() (interface{}, *string, error) {
		return a.Accessor.EpochInvites(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.Invite), nextContinuationToken, err
}

func (a *cachedAccessor) EpochInvitesOld(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error) {
	res, err := a.getOrLoad("EpochInvitesOld", func() (interface{}, error) {
		return a.Accessor.EpochInvitesOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.Invite), err
}

func (a *cachedAccessor) EpochTxsCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad("EpochTxsCount", func() (interface{}, error) {
		return a.Accessor.EpochTxsCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochTxs(epoch uint64, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("EpochTxs", func() (interface{}, *string, error) {
		return a.Accessor.EpochTxs(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.TransactionSummary), nextContinuationToken, err
}

func (a *cachedAccessor) EpochTxsOld(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	res, err := a.getOrLoad("EpochTxsOld", func() (interface{}, error) {
		return a.Accessor.EpochTxsOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.TransactionSummary), err
}

func (a *cachedAccessor) EpochCoins(epoch uint64) (types.AllCoins, error) {
	res, err := a.getOrLoad("EpochCoins", func() (interface{}, error) {
		return a.Accessor.EpochCoins(epoch)
	}, epoch)
	return res.(types.AllCoins), err
}

func (a *cachedAccessor) EpochRewardsSummary(epoch uint64) (types.RewardsSummary, error) {
	res, err := a.getOrLoad(epochRewardsSummaryMethod, func() (interface{}, error) {
		return a.Accessor.EpochRewardsSummary(epoch)
	}, epoch)
	return res.(types.RewardsSummary), err
}

func (a *cachedAccessor) EpochBadAuthorsCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad(epochBadAuthorsCountMethod, func() (interface{}, error) {
		return a.Accessor.EpochBadAuthorsCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochBadAuthors(epoch uint64, count uint64, continuationToken *string) ([]types.BadAuthor, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken(epochBadAuthorsMethod, func() (interface{}, *string, error) {
		return a.Accessor.EpochBadAuthors(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.BadAuthor), nextContinuationToken, err
}

func (a *cachedAccessor) EpochBadAuthorsOld(epoch uint64, startIndex uint64, count uint64) ([]types.BadAuthor, error) {
	res, err := a.getOrLoad(epochBadAuthorsOldMethod, func() (interface{}, error) {
		return a.Accessor.EpochBadAuthorsOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.BadAuthor), err
}

func (a *cachedAccessor) EpochRewardsCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad(epochRewardsCountMethod, func() (interface{}, error) {
		return a.Accessor.EpochRewardsCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Reward, error) {
	res, err := a.getOrLoad(epochRewardsMethod, func() (interface{}, error) {
		return a.Accessor.EpochRewards(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.Reward), err
}

func (a *cachedAccessor) EpochIdentitiesRewardsCount(epoch uint64) (uint64, error) {
	res, err := a.getOrLoad(epochIdentitiesRewardsCountMethod, func() (interface{}, error) {
		return a.Accessor.EpochIdentitiesRewardsCount(epoch)
	}, epoch)
	return res.(uint64), err
}

func (a *cachedAccessor) EpochIdentitiesRewards(epoch uint64, count uint64, continuationToken *string) ([]types.Rewards, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken(epochIdentitiesRewardsMethod, func() (interface{}, *string, error) {
		return a.Accessor.EpochIdentitiesRewards(epoch, count, continuationToken)
	}, epoch, count, continuationToken)
	return res.([]types.Rewards), nextContinuationToken, err
}

func (a *cachedAccessor) EpochIdentitiesRewardsOld(epoch uint64, startIndex uint64, count uint64) ([]types.Rewards, error) {
	res, err := a.getOrLoad(epochIdentitiesRewardsOldMethod, func() (interface{}, error) {
		return a.Accessor.EpochIdentitiesRewardsOld(epoch, startIndex, count)
	}, epoch, startIndex, count)
	return res.([]types.Rewards), err
}

func (a *cachedAccessor) EpochFundPayments(epoch uint64) ([]types.FundPayment, error) {
	res, err := a.getOrLoad(epochFundPaymentsMethod, func() (interface{}, error) {
		return a.Accessor.EpochFundPayments(epoch)
	}, epoch)
	return res.([]types.FundPayment), err
}

func (a *cachedAccessor) EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	res, err := a.getOrLoad("EpochIdentity", func() (interface{}, error) {
		return a.Accessor.EpochIdentity(epoch, address)
	}, epoch, address)
	return res.(types.EpochIdentity), err
}

func (a *cachedAccessor) EpochIdentityShortFlipsToSolve(epoch uint64, address string) ([]string, error) {
	res, err := a.getOrLoad("EpochIdentityShortFlipsToSolve", func() (interface{}, error) {
		return a.Accessor.EpochIdentityShortFlipsToSolve(epoch, address)
	}, epoch, address)
	return res.([]string), err
}

func (a *cachedAccessor) EpochIdentityLongFlipsToSolve(epoch uint64, address string) ([]string, error) {
	res, err := a.getOrLoad("EpochIdentityLongFlipsToSolve", func() (interface{}, error) {
		return a.Accessor.EpochIdentityLongFlipsToSolve(epoch, address)
	}, epoch, address)
	return res.([]string), err
}

func (a *cachedAccessor) EpochIdentityShortAnswers(epoch uint64, address string) ([]types.Answer, error) {
	res, err := a.getOrLoad("EpochIdentityShortAnswers", func() (interface{}, error) {
		return a.Accessor.EpochIdentityShortAnswers(epoch, address)
	}, epoch, address)
	return res.([]types.Answer), err
}

func (a *cachedAccessor) EpochIdentityLongAnswers(epoch uint64, address string) ([]types.Answer, error) {
	res, err := a.getOrLoad("EpochIdentityLongAnswers", func() (interface{}, error) {
		return a.Accessor.EpochIdentityLongAnswers(epoch, address)
	}, epoch, address)
	return res.([]types.Answer), err
}

func (a *cachedAccessor) EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error) {
	res, err := a.getOrLoad("EpochIdentityFlips", func() (interface{}, error) {
		return a.Accessor.EpochIdentityFlips(epoch, address)
	}, epoch, address)
	return res.([]types.FlipSummary), err
}

func (a *cachedAccessor) EpochIdentityFlipsWithRewardFlag(epoch uint64, address string) ([]types.FlipWithRewardFlag, error) {
	res, err := a.getOrLoad("EpochIdentityFlipsWithRewardFlag", func() (interface{}, error) {
		return a.Accessor.EpochIdentityFlipsWithRewardFlag(epoch, address)
	}, epoch, address)
	return res.([]types.FlipWithRewardFlag), err
}

func (a *cachedAccessor) EpochIdentityReportedFlipRewards(epoch uint64, address string) ([]types.ReportedFlipReward, error) {
	res, err := a.getOrLoad("EpochIdentityReportedFlipRewards", func() (interface{}, error) {
		return a.Accessor.EpochIdentityReportedFlipRewards(epoch, address)
	}, epoch, address)
	return res.([]types.ReportedFlipReward), err
}

func (a *cachedAccessor) EpochIdentityValidationTxs(epoch uint64, address string) ([]types.TransactionSummary, error) {
	res, err := a.getOrLoad("EpochIdentityValidationTxs", func() (interface{}, error) {
		return a.Accessor.EpochIdentityValidationTxs(epoch, address)
	}, epoch, address)
	return res.([]types.TransactionSummary), err
}

func (a *cachedAccessor) EpochIdentityRewards(epoch uint64, address string) ([]types.Reward, error) {
	res, err := a.getOrLoad("EpochIdentityRewards", func() (interface{}, error) {
		return a.Accessor.EpochIdentityRewards(epoch, address)
	}, epoch, address)
	return res.([]types.Reward), err
}

func (a *cachedAccessor) EpochIdentityBadAuthor(epoch uint64, address string) (*types.BadAuthor, error) {
	res, err := a.getOrLoad("EpochIdentityBadAuthor", func() (interface{}, error) {
		return a.Accessor.EpochIdentityBadAuthor(epoch, address)
	}, epoch, address)
	return res.(*types.BadAuthor), err
}

func (a *cachedAccessor) EpochIdentityInvitesWithRewardFlag(epoch uint64, address string) ([]types.InviteWithRewardFlag, error) {
	res, err := a.getOrLoad("EpochIdentityInvitesWithRewardFlag", func() (interface{}, error) {
		return a.Accessor.EpochIdentityInvitesWithRewardFlag(epoch, address)
	}, epoch, address)
	return res.([]types.InviteWithRewardFlag), err
}

func (a *cachedAccessor) EpochIdentitySavedInviteRewards(epoch uint64, address string) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad("EpochIdentitySavedInviteRewards", func() (interface{}, error) {
		return a.Accessor.EpochIdentitySavedInviteRewards(epoch, address)
	}, epoch, address)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) EpochIdentityAvailableInvites(epoch uint64, address string) ([]types.EpochInvites, error) {
	res, err := a.getOrLoad("EpochIdentityAvailableInvites", func() (interface{}, error) {
		return a.Accessor.EpochIdentityAvailableInvites(epoch, address)
	}, epoch, address)
	return res.([]types.EpochInvites), err
}

func (a *cachedAccessor) BlockByHeight(height uint64) (types.BlockDetail, error) {
	res, err := a.getOrLoad("BlockByHeight", func() (interface{}, error) {
		return a.Accessor.BlockByHeight(height)
	}, height)
	return res.(types.BlockDetail), err
}

func (a *cachedAccessor) BlockTxsCountByHeight(height uint64) (uint64, error) {
	res, err := a.getOrLoad("BlockTxsCountByHeight", func() (interface{}, error) {
		return a.Accessor.BlockTxsCountByHeight(height)
	}, height)
	return res.(uint64), err
}

func (a *cachedAccessor) BlockTxsByHeight(height uint64, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("BlockTxsByHeight", func() (interface{}, *string, error) {
		return a.Accessor.BlockTxsByHeight(height, count, continuationToken)
	}, height, count, continuationToken)
	return res.([]types.TransactionSummary), nextContinuationToken, err
}

func (a *cachedAccessor) BlockTxsByHeightOld(height uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	res, err := a.getOrLoad("BlockTxsByHeightOld", func() (interface{}, error) {
		return a.Accessor.BlockTxsByHeightOld(height, startIndex, count)
	}, height, startIndex, count)
	return res.([]types.TransactionSummary), err
}

func (a *cachedAccessor) BlockByHash(hash string) (types.BlockDetail, error) {
	res, err := a.getOrLoad("BlockByHash", func() (interface{}, error) {
		return a.Accessor.BlockByHash(hash)
	}, hash)
	return res.(types.BlockDetail), err
}

func (a *cachedAccessor) BlockTxsCountByHash(hash string) (uint64, error) {
	res, err := a.getOrLoad("BlockTxsCountByHash", func() (interface{}, error) {
		return a.Accessor.BlockTxsCountByHash(hash)
	}, hash)
	return res.(uint64), err
}

func (a *cachedAccessor) BlockTxsByHash(hash string, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("BlockTxsByHash", func() (interface{}, *string, error) {
		return a.Accessor.BlockTxsByHash(hash, count, continuationToken)
	}, hash, count, continuationToken)
	return res.([]types.TransactionSummary), nextContinuationToken, err
}

func (a *cachedAccessor) BlockTxsByHashOld(hash string, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	res, err := a.getOrLoad("BlockTxsByHashOld", func() (interface{}, error) {
		return a.Accessor.BlockTxsByHashOld(hash, startIndex, count)
	}, hash, startIndex, count)
	return res.([]types.TransactionSummary), err
}

func (a *cachedAccessor) BlockCoinsByHeight(height uint64) (types.AllCoins, error) {
	res, err := a.getOrLoad("BlockCoinsByHeight", func() (interface{}, error) {
		return a.Accessor.BlockCoinsByHeight(height)
	}, height)
	return res.(types.AllCoins), err
}

func (a *cachedAccessor) BlockCoinsByHash(hash string) (types.AllCoins, error) {
	res, err := a.getOrLoad("BlockCoinsByHash", func() (interface{}, error) {
		return a.Accessor.BlockCoinsByHash(hash)
	}, hash)
	return res.(types.AllCoins), err
}

func (a *cachedAccessor) Flip(hash string) (types.Flip, error) {
	res, err := a.getOrLoad("Flip", func() (interface{}, error) {
		return a.Accessor.Flip(hash)
	}, hash)
	return res.(types.Flip), err
}

func (a *cachedAccessor) FlipContent(hash string) (types.FlipContent, error) {
	res, err := a.getOrLoad("FlipContent", func() (interface{}, error) {
		return a.Accessor.FlipContent(hash)
	}, hash)
	return res.(types.FlipContent), err
}

func (a *cachedAccessor) FlipAnswers(hash string, isShort bool) ([]types.Answer, error) {
	res, err := a.getOrLoad("FlipAnswers", func() (interface{}, error) {
		return a.Accessor.FlipAnswers(hash, isShort)
	}, hash, isShort)
	return res.([]types.Answer), err
}

func (a *cachedAccessor) FlipEpochAdjacentFlips(hash string) (types.AdjacentStrValues, error) {
	res, err := a.getOrLoad(flipEpochAdjacentFlipsMethod, func() (interface{}, error) {
		return a.Accessor.FlipEpochAdjacentFlips(hash)
	}, hash)
	return res.(types.AdjacentStrValues), err
}

func (a *cachedAccessor) Identity(address string) (types.Identity, error) {
	res, err := a.getOrLoad("Identity", func() (interface{}, error) {
		return a.Accessor.Identity(address)
	}, address)
	return res.(types.Identity), err
}

func (a *cachedAccessor) IdentityAge(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityAge", func() (interface{}, error) {
		return a.Accessor.IdentityAge(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityCurrentFlipCids(address string) ([]string, error) {
	res, err := a.getOrLoad("IdentityCurrentFlipCids", func() (interface{}, error) {
		return a.Accessor.IdentityCurrentFlipCids(address)
	}, address)
	return res.([]string), err
}

func (a *cachedAccessor) IdentityEpochsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityEpochsCount", func() (interface{}, error) {
		return a.Accessor.IdentityEpochsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityEpochs(address string, count uint64, continuationToken *string) ([]types.EpochIdentity, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityEpochs", func() (interface{}, *string, error) {
		return a.Accessor.IdentityEpochs(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.EpochIdentity), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityEpochsOld(address string, startIndex uint64, count uint64) ([]types.EpochIdentity, error) {
	res, err := a.getOrLoad("IdentityEpochsOld", func() (interface{}, error) {
		return a.Accessor.IdentityEpochsOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.EpochIdentity), err
}

func (a *cachedAccessor) IdentityFlipsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityFlipsCount", func() (interface{}, error) {
		return a.Accessor.IdentityFlipsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityFlips(address string, count uint64, continuationToken *string) ([]types.FlipSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityFlips", func() (interface{}, *string, error) {
		return a.Accessor.IdentityFlips(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.FlipSummary), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityFlipsOld(address string, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	res, err := a.getOrLoad("IdentityFlipsOld", func() (interface{}, error) {
		return a.Accessor.IdentityFlipsOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.FlipSummary), err
}

func (a *cachedAccessor) IdentityFlipQualifiedAnswers(address string) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad("IdentityFlipQualifiedAnswers", func() (interface{}, error) {
		return a.Accessor.IdentityFlipQualifiedAnswers(address)
	}, address)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) IdentityFlipStates(address string) ([]types.StrValueCount, error) {
	res, err := a.getOrLoad("IdentityFlipStates", func() (interface{}, error) {
		return a.Accessor.IdentityFlipStates(address)
	}, address)
	return res.([]types.StrValueCount), err
}

func (a *cachedAccessor) IdentityInvitesCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityInvitesCount", func() (interface{}, error) {
		return a.Accessor.IdentityInvitesCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityInvites(address string, count uint64, continuationToken *string) ([]types.Invite, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityInvites", func() (interface{}, *string, error) {
		return a.Accessor.IdentityInvites(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.Invite), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityInvitesOld(address string, startIndex uint64, count uint64) ([]types.Invite, error) {
	res, err := a.getOrLoad("IdentityInvitesOld", func() (interface{}, error) {
		return a.Accessor.IdentityInvitesOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.Invite), err
}

func (a *cachedAccessor) IdentityTxsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityTxsCount", func() (interface{}, error) {
		return a.Accessor.IdentityTxsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityTxs(address string, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityTxs", func() (interface{}, *string, error) {
		return a.Accessor.IdentityTxs(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.TransactionSummary), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityTxsOld(address string, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	res, err := a.getOrLoad("IdentityTxsOld", func() (interface{}, error) {
		return a.Accessor.IdentityTxsOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.TransactionSummary), err
}

func (a *cachedAccessor) IdentityRewardsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityRewardsCount", func() (interface{}, error) {
		return a.Accessor.IdentityRewardsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityRewards(address string, count uint64, continuationToken *string) ([]types.Reward, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityRewards", func() (interface{}, *string, error) {
		return a.Accessor.IdentityRewards(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.Reward), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityRewardsOld(address string, startIndex uint64, count uint64) ([]types.Reward, error) {
	res, err := a.getOrLoad("IdentityRewardsOld", func() (interface{}, error) {
		return a.Accessor.IdentityRewardsOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.Reward), err
}

func (a *cachedAccessor) IdentityEpochRewardsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("IdentityEpochRewardsCount", func() (interface{}, error) {
		return a.Accessor.IdentityEpochRewardsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) IdentityEpochRewards(address string, count uint64, continuationToken *string) ([]types.Rewards, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("IdentityEpochRewards", func() (interface{}, *string, error) {
		return a.Accessor.IdentityEpochRewards(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.Rewards), nextContinuationToken, err
}

func (a *cachedAccessor) IdentityEpochRewardsOld(address string, startIndex uint64, count uint64) ([]types.Rewards, error) {
	res, err := a.getOrLoad("IdentityEpochRewardsOld", func() (interface{}, error) {
		return a.Accessor.IdentityEpochRewardsOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.Rewards), err
}

func (a *cachedAccessor) Address(address string) (types.Address, error) {
	res, err := a.getOrLoad("Address", func() (interface{}, error) {
		return a.Accessor.Address(address)
	}, address)
	return res.(types.Address), err
}

func (a *cachedAccessor) AddressPenaltiesCount(address string) (uint64, error) {
	res, err := a.getOrLoad("AddressPenaltiesCount", func() (interface{}, error) {
		return a.Accessor.AddressPenaltiesCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) AddressPenalties(address string, count uint64, continuationToken *string) ([]types.Penalty, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("AddressPenalties", func() (interface{}, *string, error) {
		return a.Accessor.AddressPenalties(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.Penalty), nextContinuationToken, err
}

func (a *cachedAccessor) AddressPenaltiesOld(address string, startIndex uint64, count uint64) ([]types.Penalty, error) {
	res, err := a.getOrLoad("AddressPenaltiesOld", func() (interface{}, error) {
		return a.Accessor.AddressPenaltiesOld(address, startIndex, count)
	}, address, startIndex, count)
	return res.([]types.Penalty), err
}

func (a *cachedAccessor) AddressStatesCount(address string) (uint64, error) {
	res, err := a.getOrLoad("AddressStatesCount", func() (interface{}, error) {
		return a.Accessor.AddressStatesCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) AddressStates(address string, count uint64, continuationToken *string) ([]types.AddressState, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("AddressStates", func() (interface{}, *string, error) {
		return a.Accessor.AddressStates(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.AddressState), nextContinuationToken, err
}

func (a *cachedAccessor) AddressTotalLatestMiningReward(afterTime time.Time, address string) (types.TotalMiningReward, error) {
	res, err := a.getOrLoad("AddressTotalLatestMiningReward", func() (interface{}, error) {
		return a.Accessor.AddressTotalLatestMiningReward(afterTime, address)
	}, afterTime, address)
	return res.(types.TotalMiningReward), err
}

func (a *cachedAccessor) AddressTotalLatestBurntCoins(afterTime time.Time, address string) (types.AddressBurntCoins, error) {
	res, err := a.getOrLoad("AddressTotalLatestBurntCoins", func() (interface{}, error) {
		return a.Accessor.AddressTotalLatestBurntCoins(afterTime, address)
	}, afterTime, address)
	return res.(types.AddressBurntCoins), err
}

func (a *cachedAccessor) AddressBadAuthorsCount(address string) (uint64, error) {
	res, err := a.getOrLoad("AddressBadAuthorsCount", func() (interface{}, error) {
		return a.Accessor.AddressBadAuthorsCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) AddressBadAuthors(address string, count uint64, continuationToken *string) ([]types.BadAuthor, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("AddressBadAuthors", func() (interface{}, *string, error) {
		return a.Accessor.AddressBadAuthors(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.BadAuthor), nextContinuationToken, err
}

func (a *cachedAccessor) AddressBalanceUpdatesCount(address string) (uint64, error) {
	res, err := a.getOrLoad("AddressBalanceUpdatesCount", func() (interface{}, error) {
		return a.Accessor.AddressBalanceUpdatesCount(address)
	}, address)
	return res.(uint64), err
}

func (a *cachedAccessor) AddressBalanceUpdates(address string, count uint64, continuationToken *string) ([]types.BalanceUpdate, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("AddressBalanceUpdates", func() (interface{}, *string, error) {
		return a.Accessor.AddressBalanceUpdates(address, count, continuationToken)
	}, address, count, continuationToken)
	return res.([]types.BalanceUpdate), nextContinuationToken, err
}

func (a *cachedAccessor) Transaction(hash string) (types.TransactionDetail, error) {
	res, err := a.getOrLoad("Transaction", func() (interface{}, error) {
		return a.Accessor.Transaction(hash)
	}, hash)
	return res.(types.TransactionDetail), err
}

func (a *cachedAccessor) TransactionRaw(hash string) (hexutil.Bytes, error) {
	res, err := a.getOrLoad("TransactionRaw", func() (interface{}, error) {
		return a.Accessor.TransactionRaw(hash)
	}, hash)
	return res.(hexutil.Bytes), err
}

func (a *cachedAccessor) BalancesCount() (uint64, error) {
	res, err := a.getOrLoad("BalancesCount", func() (interface{}, error) {
		return a.Accessor.BalancesCount()
	})
	return res.(uint64), err
}

func (a *cachedAccessor) Balances(count uint64, continuationToken *string) ([]types.Balance, *string, error) {
	res, nextContinuationToken, err := a.getOrLoadWithConToken("Balances", func() (interface{}, *string, error) {
		return a.Accessor.Balances(count, continuationToken)
	}, count, continuationToken)
	return res.([]types.Balance), nextContinuationToken, err
}

func (a *cachedAccessor) BalancesOld(startIndex uint64, count uint64) ([]types.Balance, error) {
	res, err := a.getOrLoad("BalancesOld", func() (interface{}, error) {
		return a.Accessor.BalancesOld(startIndex, count)
	}, startIndex, count)
	return res.([]types.Balance), err
}

func (a *cachedAccessor) TotalLatestMiningRewardsCount(afterTime time.Time) (uint64, error) {
	res, err := a.getOrLoad("TotalLatestMiningRewardsCount", func() (interface{}, error) {
		return a.Accessor.TotalLatestMiningRewardsCount(afterTime)
	}, afterTime)
	return res.(uint64), err
}

func (a *cachedAccessor) TotalLatestMiningRewards(afterTime time.Time, startIndex uint64,
	count uint64) ([]types.TotalMiningReward, error) {
	res, err := a.getOrLoad("TotalLatestMiningRewards", func() (interface{}, error) {
		return a.Accessor.TotalLatestMiningRewards(afterTime, startIndex, count)
	}, afterTime, startIndex, count)
	return res.([]types.TotalMiningReward), err
}

func (a *cachedAccessor) TotalLatestBurntCoinsCount(afterTime time.Time) (uint64, error) {
	res, err := a.getOrLoad("TotalLatestBurntCoinsCount", func() (interface{}, error) {
		return a.Accessor.TotalLatestBurntCoinsCount(afterTime)
	}, afterTime)
	return res.(uint64), err
}

func (a *cachedAccessor) TotalLatestBurntCoins(afterTime time.Time, startIndex uint64,
	count uint64) ([]types.AddressBurntCoins, error) {
	res, err := a.getOrLoad("TotalLatestBurntCoins", func() (interface{}, error) {
		return a.Accessor.TotalLatestBurntCoins(afterTime, startIndex, count)
	}, afterTime, startIndex, count)
	return res.([]types.AddressBurntCoins), err
}

func (a *cachedAccessor) Destroy() {
	a.Accessor.Destroy()
}
