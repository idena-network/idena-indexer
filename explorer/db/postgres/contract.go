package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	contractQuery                 = "contract.sql"
	contractTxBalanceUpdatesQuery = "contractTxBalanceUpdates.sql"
)

func (a *postgresAccessor) Contract(address string) (types.Contract, error) {
	res := types.Contract{}
	err := a.db.QueryRow(a.getQuery(contractQuery), address).Scan(&res.Type, &res.Author)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Contract{}, err
	}
	res.Address = address
	return res, nil
}

func (a *postgresAccessor) ContractTxBalanceUpdates(contractAddress string, count uint64, continuationToken *string) ([]types.ContractTxBalanceUpdate, *string, error) {
	res, nextContinuationToken, err := a.page(contractTxBalanceUpdatesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
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
	}, count, continuationToken, contractAddress)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.ContractTxBalanceUpdate), nextContinuationToken, nil
}
