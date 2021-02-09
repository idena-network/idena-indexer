package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	contractQuery                 = "contract.sql"
	timeLockContractQuery         = "timeLockContract.sql"
	oracleVotingContractQuery     = "oracleVotingContract.sql"
	contractTxBalanceUpdatesQuery = "contractTxBalanceUpdates.sql"
)

func (a *postgresAccessor) Contract(address string) (types.Contract, error) {
	res := types.Contract{}
	var terminationTxTime sql.NullInt64
	var terminationTxHash sql.NullString
	var deployTxTimestamp int64
	err := a.db.QueryRow(a.getQuery(contractQuery), address).Scan(
		&res.Type,
		&res.Author,
		&res.DeployTx.Hash,
		&deployTxTimestamp,
		&terminationTxHash,
		&terminationTxTime,
	)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Contract{}, err
	}
	res.Address = address
	res.DeployTx.Timestamp = timestampToTimeUTC(deployTxTimestamp)
	if terminationTxHash.Valid {
		res.TerminationTx = &types.TransactionSummary{
			Hash:      terminationTxHash.String,
			Timestamp: timestampToTimeUTC(terminationTxTime.Int64),
		}
	}
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

func (a *postgresAccessor) TimeLockContract(address string) (types.TimeLockContract, error) {
	res := types.TimeLockContract{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(timeLockContractQuery), address).Scan(&timestamp)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.TimeLockContract{}, err
	}
	res.Timestamp = timestampToTimeUTC(timestamp)
	return res, nil
}

func (a *postgresAccessor) readTimeLockContracts(rows *sql.Rows) ([]types.TimeLockContract, *string, error) {
	var res []types.TimeLockContract
	for rows.Next() {
		item := types.TimeLockContract{}
		var timestamp int64
		if err := rows.Scan(&timestamp); err != nil {
			return nil, nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
	}
	return res, nil, nil
}

func (a *postgresAccessor) OracleVotingContract(address, oracle string) (types.OracleVotingContract, error) {
	rows, err := a.db.Query(a.getQuery(oracleVotingContractQuery), address, oracle)
	if err != nil {
		return types.OracleVotingContract{}, err
	}
	defer rows.Close()
	contracts, _, err := a.readOracleVotingContracts(rows)
	if err != nil {
		return types.OracleVotingContract{}, err
	}
	if len(contracts) == 0 {
		return types.OracleVotingContract{}, NoDataFound
	}
	return contracts[0], nil
}
