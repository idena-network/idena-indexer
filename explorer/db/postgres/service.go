package postgres

import (
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
	"math"
	"sync"
	"time"
)

type estimatedOracleRewardsService struct {
	cache         *estimatedOracleRewardsServiceCache
	mutex         sync.Mutex
	feeRateFn     func() (decimal.Decimal, error)
	networkSizeFn func() (uint64, error)
}

type estimatedOracleRewardsServiceCache struct {
	feeRate     decimal.Decimal
	networkSize uint64
}

func newEstimatedOracleRewardsCache(
	feeRateFn func() (decimal.Decimal, error),
	networkSizeFn func() (uint64, error),
) *estimatedOracleRewardsService {
	res := &estimatedOracleRewardsService{
		feeRateFn:     feeRateFn,
		networkSizeFn: networkSizeFn,
	}
	go func() {
		for {
			time.Sleep(time.Minute)
			res.cache = nil
		}
	}()
	return res
}

func (c *estimatedOracleRewardsService) get(committeeSize uint64) ([]types.EstimatedOracleReward, error) {
	data := c.cache
	if data == nil {
		c.mutex.Lock()
		data = c.cache
		if data == nil {
			var err error
			data, err = c.loadData()
			if err != nil {
				c.mutex.Unlock()
				return nil, err
			}
		}
		c.mutex.Unlock()
	}
	return createEstimatedOracleRewardsService(committeeSize, data.networkSize, data.feeRate), nil
}

func (c *estimatedOracleRewardsService) loadData() (*estimatedOracleRewardsServiceCache, error) {
	feeRate, err := c.feeRateFn()
	if err != nil {
		return nil, err
	}
	networkSize, err := c.networkSizeFn()
	if err != nil {
		return nil, err
	}
	return &estimatedOracleRewardsServiceCache{
		feeRate:     feeRate,
		networkSize: networkSize,
	}, nil
}

func createEstimatedOracleRewardsService(committeeSize, networkSize uint64, feeRate decimal.Decimal) []types.EstimatedOracleReward {
	if networkSize == 0 {
		networkSize = 1
	}
	minOracleReward := decimal.NewFromFloat(math.Pow(math.Log10(float64(networkSize)), math.Log10(100.0*float64(committeeSize)/float64(networkSize))/2))
	return []types.EstimatedOracleReward{
		{
			Amount: minOracleReward,
			Type:   "min",
		},
		{
			Amount: decimal.RequireFromString("250000").Mul(feeRate),
			Type:   "slow",
		},
		{
			Amount: decimal.RequireFromString("500000").Mul(feeRate),
			Type:   "medium",
		},
		{
			Amount: decimal.RequireFromString("2500000").Mul(feeRate),
			Type:   "fast",
		},
		{
			Amount: decimal.RequireFromString("5000000").Mul(feeRate),
			Type:   "fastest",
		},
	}
}
