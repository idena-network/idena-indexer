package postgres

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	addressPenaltiesCountQuery          = "addressPenaltiesCount.sql"
	addressPenaltiesQuery               = "addressPenalties.sql"
	addressMiningRewardsCountQuery      = "addressMiningRewardsCount.sql"
	addressMiningRewardsQuery           = "addressMiningRewards.sql"
	addressBlockMiningRewardsCountQuery = "addressBlockMiningRewardsCount.sql"
	addressBlockMiningRewardsQuery      = "addressBlockMiningRewards.sql"
	addressStatesCountQuery             = "addressStatesCount.sql"
	addressStatesQuery                  = "addressStates.sql"
)

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
