package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
	"time"
)

const (
	addressQuery                         = "address.sql"
	addressPenaltiesCountQuery           = "addressPenaltiesCount.sql"
	addressPenaltiesQuery                = "addressPenalties.sql"
	addressPenaltiesOldQuery             = "addressPenaltiesOld.sql"
	addressStatesCountQuery              = "addressStatesCount.sql"
	addressStatesQuery                   = "addressStates.sql"
	addressTotalLatestMiningRewardQuery  = "addressTotalLatestMiningReward.sql"
	addressTotalLatestBurntCoinsQuery    = "addressTotalLatestBurntCoins.sql"
	addressBadAuthorsCountQuery          = "addressBadAuthorsCount.sql"
	addressBadAuthorsQuery               = "addressBadAuthors.sql"
	addressBalanceUpdatesCountQuery      = "addressBalanceUpdatesCount.sql"
	addressBalanceUpdatesQuery           = "addressBalanceUpdates.sql"
	addressContractTxBalanceUpdatesQuery = "addressContractTxBalanceUpdates.sql"

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

func (a *postgresAccessor) AddressPenalties(address string, count uint64, continuationToken *string) ([]types.Penalty, *string, error) {
	res, nextContinuationToken, err := a.page(addressPenaltiesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.Penalty
		var id uint64
		for rows.Next() {
			item := types.Penalty{}
			var timestamp int64
			if err := rows.Scan(
				&id,
				&item.Address,
				&item.Penalty,
				&item.Paid,
				&item.BlockHeight,
				&item.BlockHash,
				&timestamp,
				&item.Epoch,
			); err != nil {
				return nil, 0, err
			}
			item.Timestamp = timestampToTimeUTC(timestamp)
			res = append(res, item)
		}
		return res, id, nil
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.Penalty), nextContinuationToken, nil
}

func (a *postgresAccessor) AddressPenaltiesOld(address string, startIndex uint64, count uint64) ([]types.Penalty, error) {
	rows, err := a.db.Query(a.getQuery(addressPenaltiesOldQuery), address, startIndex, count)
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

func (a *postgresAccessor) AddressStatesCount(address string) (uint64, error) {
	return a.count(addressStatesCountQuery, address)
}

func (a *postgresAccessor) AddressStates(address string, count uint64, continuationToken *string) ([]types.AddressState, *string, error) {
	res, nextContinuationToken, err := a.page(addressStatesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.AddressState
		var id uint64
		for rows.Next() {
			item := types.AddressState{}
			var timestamp int64
			if err := rows.Scan(
				&id,
				&item.State,
				&item.Epoch,
				&item.BlockHeight,
				&item.BlockHash,
				&item.TxHash,
				&timestamp,
				&item.IsValidation,
			); err != nil {
				return nil, 0, err
			}
			item.Timestamp = timestampToTimeUTC(timestamp)
			res = append(res, item)
		}
		return res, id, nil
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.AddressState), nextContinuationToken, nil
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

func (a *postgresAccessor) AddressBadAuthors(address string, count uint64, continuationToken *string) ([]types.BadAuthor, *string, error) {
	res, nextContinuationToken, err := a.page(addressBadAuthorsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readBadAuthors(rows)
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.BadAuthor), nextContinuationToken, nil
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

func (a *postgresAccessor) AddressBalanceUpdates(address string, count uint64, continuationToken *string) ([]types.BalanceUpdate, *string, error) {
	res, nextContinuationToken, err := a.page(addressBalanceUpdatesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.BalanceUpdate
		var id uint64
		for rows.Next() {
			item := types.BalanceUpdate{}
			var timestamp int64
			optionalData := &balanceUpdateOptionalData{}
			if err := rows.Scan(
				&id,
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
			); err != nil {
				return nil, 0, err
			}
			item.Timestamp = timestampToTimeUTC(timestamp)
			item.Data = readBalanceUpdateSpecificData(item.Reason, optionalData)
			res = append(res, item)
		}
		return res, id, nil
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.BalanceUpdate), nextContinuationToken, nil
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

func (a *postgresAccessor) AddressContractTxBalanceUpdates(address string, contractAddress string, count uint64, continuationToken *string) ([]types.ContractTxBalanceUpdate, *string, error) {
	res, nextContinuationToken, err := a.page(addressContractTxBalanceUpdatesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.ContractTxBalanceUpdate
		var id uint64
		for rows.Next() {
			item := types.ContractTxBalanceUpdate{}
			var timestamp int64
			var callMethod sql.NullInt32
			var balanceOld, balanceNew NullDecimal
			if err := rows.Scan(
				&id,
				&item.Hash,
				&item.Type,
				&timestamp,
				&item.From,
				&item.To,
				&item.Amount,
				&item.Tips,
				&item.MaxFee,
				&item.Fee,
				&item.Address,
				&item.ContractAddress,
				&item.ContractType,
				&callMethod,
				&balanceOld,
				&balanceNew,
			); err != nil {
				return nil, 0, err
			}
			item.Timestamp = timestampToTimeUTC(timestamp)
			if callMethod.Valid {
				item.ContractCallMethod = types.GetCallMethodName(item.ContractType, uint8(callMethod.Int32))
			}
			if balanceOld.Valid && balanceNew.Valid {
				change := balanceNew.Decimal.Sub(balanceOld.Decimal)
				item.BalanceChange = &change
			}
			res = append(res, item)
		}
		return res, id, nil
	}, count, continuationToken, address, contractAddress)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.ContractTxBalanceUpdate), nextContinuationToken, nil
}
