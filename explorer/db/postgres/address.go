package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
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
	addressBalanceUpdatesCountQuery     = "addressBalanceUpdatesCount.sql"
	addressBalanceUpdatesQuery          = "addressBalanceUpdates.sql"

	txBalanceUpdateReason              = "Tx"
	committeeRewardBalanceUpdateReason = "CommitteeReward"
)

func (a *postgresAccessor) Address(address string) (types.Address, error) {
	res := types.Address{}
	err := a.db.QueryRow(a.getQuery(addressQuery), address).Scan(
		&res.Address,
		&res.Balance,
		&res.Stake,
		&res.TxCount,
		&res.FlipsCount,
		&res.ReportedFlipsCount,
	)
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
		item.Timestamp = timestampToTimeUTC(timestamp)
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
		item.Timestamp = timestampToTimeUTC(timestamp)
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

func (a *postgresAccessor) AddressBalanceUpdatesCount(address string) (uint64, error) {
	return a.count(addressBalanceUpdatesCountQuery, address)
}

type balanceUpdateOptionalData struct {
	txHash             string
	lastBlockHeight    uint64
	lastBlockHash      string
	lastBlockTimestamp int64
	rewardShare        decimal.Decimal
	blocksCount        uint32
}

func (a *postgresAccessor) AddressBalanceUpdates(address string, startIndex uint64, count uint64) ([]types.BalanceUpdate, error) {
	rows, err := a.db.Query(a.getQuery(addressBalanceUpdatesQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.BalanceUpdate
	for rows.Next() {
		item := types.BalanceUpdate{}
		var timestamp int64
		optionalData := &balanceUpdateOptionalData{}
		err = rows.Scan(
			&item.BalanceOld,
			&item.StakeOld,
			&item.PenaltyOld,
			&item.BalanceNew,
			&item.StakeNew,
			&item.PenaltyNew,
			&item.Reason,
			&item.BlockHeight,
			&item.BlockHash,
			&timestamp,
			&optionalData.txHash,
			&optionalData.lastBlockHeight,
			&optionalData.lastBlockHash,
			&optionalData.lastBlockTimestamp,
			&optionalData.rewardShare,
			&optionalData.blocksCount,
		)
		if err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		item.Data = readBalanceUpdateSpecificData(item.Reason, optionalData)
		res = append(res, item)
	}
	return res, nil
}

func readBalanceUpdateSpecificData(reason string, optionalData *balanceUpdateOptionalData) interface{} {
	var res interface{}
	switch reason {
	case txBalanceUpdateReason:
		res = &types.TransactionBalanceUpdate{
			TxHash: optionalData.txHash,
		}
	case committeeRewardBalanceUpdateReason:
		res = &types.CommitteeRewardBalanceUpdate{
			LastBlockHeight:    optionalData.lastBlockHeight,
			LastBlockHash:      optionalData.lastBlockHash,
			LastBlockTimestamp: timestampToTimeUTC(optionalData.lastBlockTimestamp),
			RewardShare:        optionalData.rewardShare,
			BlocksCount:        optionalData.blocksCount,
		}
	}
	return res
}
