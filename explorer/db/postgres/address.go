package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
	"time"
)

const (
	addressQuery                        = "address.sql"
	addressPenaltiesCountQuery          = "addressPenaltiesCount.sql"
	addressPenaltiesQuery               = "addressPenalties.sql"
	addressMiningRewardsCountQuery      = "addressMiningRewardsCount.sql"
	addressMiningRewardsQuery           = "addressMiningRewards.sql"
	addressBlockMiningRewardsCountQuery = "addressBlockMiningRewardsCount.sql"
	addressBlockMiningRewardsQuery      = "addressBlockMiningRewards.sql"
	addressStatesCountQuery             = "addressStatesCount.sql"
	addressStatesQuery                  = "addressStates.sql"
	addressTotalLatestMiningRewardQuery = "addressTotalLatestMiningReward.sql"
	addressTotalLatestBurntCoinsQuery   = "addressTotalLatestBurntCoins.sql"
	addressBadAuthorsCountQuery         = "addressBadAuthorsCount.sql"
	addressBadAuthorsQuery              = "addressBadAuthors.sql"
)

func (a *postgresAccessor) Address(address string) (types.Address, error) {
	res := types.Address{}
	err := a.db.QueryRow(a.getQuery(addressQuery), address).Scan(&res.Address, &res.Balance, &res.Stake, &res.TxCount)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Address{}, err
	}
	return res, nil
}

func (a *postgresAccessor) AddressPenaltiesCount(address string) (uint64, error) {
	return a.count(addressPenaltiesCountQuery, address)
}

func (a *postgresAccessor) AddressPenalties(address string, startIndex uint64, count uint64) ([]types.Penalty, error) {
	rows, err := a.db.Query(a.getQuery(addressPenaltiesQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Penalty
	for rows.Next() {
		item := types.Penalty{}
		var timestamp int64
		err = rows.Scan(&item.Address,
			&item.Penalty,
			&item.Paid,
			&item.BlockHeight,
			&item.BlockHash,
			&timestamp,
			&item.Epoch)
		if err != nil {
			return nil, err
		}
		item.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) AddressMiningRewardsCount(address string) (uint64, error) {
	return a.count(addressMiningRewardsCountQuery, address)
}

func (a *postgresAccessor) AddressMiningRewards(address string, startIndex uint64, count uint64) (
	[]types.Reward, error) {
	return a.rewards(addressMiningRewardsQuery, address, startIndex, count)
}

func (a *postgresAccessor) AddressBlockMiningRewardsCount(address string) (uint64, error) {
	return a.count(addressBlockMiningRewardsCountQuery, address)
}

func (a *postgresAccessor) AddressBlockMiningRewards(address string, startIndex uint64, count uint64) (
	[]types.BlockRewards, error) {
	rows, err := a.db.Query(a.getQuery(addressBlockMiningRewardsQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.BlockRewards
	var item *types.BlockRewards
	for rows.Next() {
		reward := types.Reward{}
		var blockHeight, epoch uint64
		if err := rows.Scan(&blockHeight, &epoch, &reward.Balance, &reward.Stake, &reward.Type); err != nil {
			return nil, err
		}
		if item == nil || item.Height != blockHeight {
			if item != nil {
				res = append(res, *item)
			}
			item = &types.BlockRewards{
				Height: blockHeight,
				Epoch:  epoch,
			}
		}
		item.Rewards = append(item.Rewards, reward)
	}
	if item != nil {
		res = append(res, *item)
	}
	return res, nil
}

func (a *postgresAccessor) AddressStatesCount(address string) (uint64, error) {
	return a.count(addressStatesCountQuery, address)
}

func (a *postgresAccessor) AddressStates(address string, startIndex uint64, count uint64) ([]types.AddressState, error) {
	rows, err := a.db.Query(a.getQuery(addressStatesQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.AddressState
	for rows.Next() {
		item := types.AddressState{}
		var timestamp int64
		err = rows.Scan(&item.State,
			&item.Epoch,
			&item.BlockHeight,
			&item.BlockHash,
			&item.TxHash,
			&timestamp,
			&item.IsValidation)
		if err != nil {
			return nil, err
		}
		item.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) AddressTotalLatestMiningReward(afterTime time.Time, address string) (types.TotalMiningReward, error) {
	res := types.TotalMiningReward{}
	err := a.db.QueryRow(a.getQuery(addressTotalLatestMiningRewardQuery), afterTime.Unix(), address).
		Scan(&res.Balance, &res.Stake, &res.Proposer, &res.FinalCommittee)
	if err != nil {
		return types.TotalMiningReward{}, err
	}
	return res, nil
}

func (a *postgresAccessor) AddressTotalLatestBurntCoins(afterTime time.Time, address string) (types.AddressBurntCoins, error) {
	res := types.AddressBurntCoins{}
	err := a.db.QueryRow(a.getQuery(addressTotalLatestBurntCoinsQuery), afterTime.Unix(), address).
		Scan(&res.Amount)
	if err != nil {
		return types.AddressBurntCoins{}, err
	}
	return res, nil
}

func (a *postgresAccessor) AddressBadAuthorsCount(address string) (uint64, error) {
	return a.count(addressBadAuthorsCountQuery, address)
}

func (a *postgresAccessor) AddressBadAuthors(address string, startIndex uint64, count uint64) ([]types.BadAuthor, error) {
	rows, err := a.db.Query(a.getQuery(addressBadAuthorsQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readBadAuthors(rows)
}
