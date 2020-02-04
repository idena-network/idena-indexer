package cached

import (
	"github.com/idena-network/idena-indexer/log"
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, lifeTime time.Duration)
	ItemsCount() int
}

func NewCache(maxSize int, defaultExpiration time.Duration, logger log.Logger) Cache {
	return &cacheImpl{
		maxSize: maxSize,
		cache:   cache.New(defaultExpiration, time.Minute*3),
		logger:  logger,
	}
}

type cacheImpl struct {
	maxSize int
	cache   *cache.Cache
	mutex   sync.Mutex
	logger  log.Logger
}

func (c *cacheImpl) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *cacheImpl) Set(key string, value interface{}, lifeTime time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache.ItemCount() >= c.maxSize {
		c.cache.DeleteExpired()
		if c.cache.ItemCount() >= c.maxSize {
			c.logger.Warn("Max size reached")
			return
		}
	}
	c.cache.Set(key, value, lifeTime)
}

func (c *cacheImpl) ItemsCount() int {
	return c.cache.ItemCount()
}
