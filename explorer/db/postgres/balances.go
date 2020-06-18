package postgres

import (
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
	"time"
)

const (
	balancesCountQuery                 = "balancesCount.sql"
	balancesQuery                      = "balances.sql"
	balancesOldQuery                   = "balancesOld.sql"
	totalLatestMiningRewardsCountQuery = "totalLatestMiningRewardsCount.sql"
	totalLatestMiningRewardsQuery      = "totalLatestMiningRewards.sql"
	totalLatestBurntCoinsCountQuery    = "totalLatestBurntCoinsCount.sql"
	totalLatestBurntCoinsQuery         = "totalLatestBurntCoins.sql"
)

func (a *postgresAccessor) BalancesCount() (uint64, error) {
	return a.count(balancesCountQuery)
}

func (a *postgresAccessor) Balances(count uint64, continuationToken *string) ([]types.Balance, *string, error) {
	parseToken := func(continuationToken *string) (addressId *uint64, balance *decimal.Decimal, err error) {
		if continuationToken == nil {
			return
		}
		strs := strings.Split(*continuationToken, "-")
		if len(strs) != 2 {
			err = errors.New("invalid continuation token")
			return
		}
		sAddressId := strs[0]
		if addressId, err = parseUintContinuationToken(&sAddressId); err != nil {
			return
		}
		var d decimal.Decimal
		d, err = decimal.NewFromString(strs[1])
		if err != nil {
			return
		}
		balance = &d
		return
	}
	addressId, balance, err := parseToken(continuationToken)
	if err != nil {
		return nil, nil, err
	}
	rows, err := a.db.Query(a.getQuery(balancesQuery), count+1, addressId, balance)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var res []types.Balance
	for rows.Next() {
		item := types.Balance{}
		err = rows.Scan(
			&addressId,
			&item.Address,
			&item.Balance,
			&item.Stake,
		)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, item)
	}
	var nextContinuationToken *string
	if len(res) > 0 && len(res) == int(count)+1 {
		t := strconv.FormatUint(*addressId, 10) + "-" + res[len(res)-1].Balance.String()
		nextContinuationToken = &t
		res = res[:len(res)-1]
	}
	return res, nextContinuationToken, nil
}

func (a *postgresAccessor) BalancesOld(startIndex uint64, count uint64) ([]types.Balance, error) {
	rows, err := a.db.Query(a.getQuery(balancesOldQuery), startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Balance
	for rows.Next() {
		item := types.Balance{}
		err = rows.Scan(&item.Address,
			&item.Balance,
			&item.Stake)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) TotalLatestMiningRewardsCount(afterTime time.Time) (uint64, error) {
	return a.count(totalLatestMiningRewardsCountQuery, afterTime.Unix())
}

func (a *postgresAccessor) TotalLatestMiningRewards(afterTime time.Time, startIndex uint64, count uint64) ([]types.TotalMiningReward, error) {
	rows, err := a.db.Query(a.getQuery(totalLatestMiningRewardsQuery), afterTime.Unix(), startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.TotalMiningReward
	for rows.Next() {
		item := types.TotalMiningReward{}
		err = rows.Scan(&item.Address,
			&item.Balance,
			&item.Stake,
			&item.Proposer,
			&item.FinalCommittee)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) TotalLatestBurntCoinsCount(afterTime time.Time) (uint64, error) {
	return a.count(totalLatestBurntCoinsCountQuery, afterTime.Unix())
}

func (a *postgresAccessor) TotalLatestBurntCoins(afterTime time.Time, startIndex uint64, count uint64) ([]types.AddressBurntCoins, error) {
	rows, err := a.db.Query(a.getQuery(totalLatestBurntCoinsQuery), afterTime.Unix(), startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.AddressBurntCoins
	for rows.Next() {
		item := types.AddressBurntCoins{}
		err = rows.Scan(&item.Address,
			&item.Amount)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
