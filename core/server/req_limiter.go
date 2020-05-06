package server

import (
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

type reqLimiter struct {
	queue               chan struct{}
	adjacentDataQueue   chan struct{}
	timeout             time.Duration
	mutex               sync.Mutex
	reqCountsByClientId *cache.Cache
	reqLimit            int
}

var (
	errTimeout        = errors.New("timeout while waiting for resource")
	errReqLimitExceed = errors.New("request limit exceeded")
)

func (limiter *reqLimiter) takeResource(clientId string, lowerUrlPath string) error {
	if err := limiter.checkReqLimit(clientId); err != nil {
		return err
	}
	var ok bool
	queue := limiter.getQueueByUrlPath(lowerUrlPath)
	select {
	case queue <- struct{}{}:
		ok = true
	case <-time.After(limiter.timeout):
	}
	if !ok {
		return errTimeout
	}
	return nil
}

func (limiter *reqLimiter) checkReqLimit(clientId string) error {
	if limiter.reqLimit <= 0 {
		return nil
	}
	getReqCount := func() int {
		if count, err := limiter.reqCountsByClientId.IncrementInt(clientId, 1); err == nil {
			return count
		}
		limiter.mutex.Lock()
		defer limiter.mutex.Unlock()
		if err := limiter.reqCountsByClientId.Add(clientId, 1, cache.DefaultExpiration); err == nil {
			return 1
		}
		if count, err := limiter.reqCountsByClientId.IncrementInt(clientId, 1); err == nil {
			return count
		}
		return 1
	}
	if getReqCount() > limiter.reqLimit {
		return errReqLimitExceed
	}
	return nil
}

func (limiter *reqLimiter) releaseResource(lowerUrlPath string) {
	queue := limiter.getQueueByUrlPath(lowerUrlPath)
	<-queue
}

func (limiter *reqLimiter) getQueueByUrlPath(lowerUrlPath string) chan struct{} {
	if strings.Contains(lowerUrlPath, "/adjacent") {
		return limiter.adjacentDataQueue
	}
	return limiter.queue
}
