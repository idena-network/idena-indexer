package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
	log     log.Logger
}

const (
	epochFlipsWithKeyQuery = "epochFlipsWithKey.sql"
	addressQuery           = "address.sql"
	transactionQuery       = "transaction.sql"
	currentEpochQuery      = "currentEpoch.sql"
)

type flipWithKey struct {
	cid string
	key string
}

var NoDataFound = errors.New("no data found")

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) getCurrentEpoch() (int64, error) {
	var epoch int64
	err := a.db.QueryRow(a.getQuery(currentEpochQuery)).Scan(&epoch)
	if err != nil {
		return 0, err
	}
	return epoch, nil
}

func (a *postgresAccessor) identityStates(queryName string, args ...interface{}) ([]types.StrValueCount, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	return a.readStrValueCounts(rows)
}

func (a *postgresAccessor) epochFlipsWithKey(epoch uint64) ([]flipWithKey, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsWithKeyQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []flipWithKey
	for rows.Next() {
		item := flipWithKey{}
		err = rows.Scan(&item.cid, &item.key)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

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

func (a *postgresAccessor) Transaction(hash string) (types.TransactionDetail, error) {
	res := types.TransactionDetail{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(transactionQuery), hash).Scan(&res.Epoch, &res.BlockHeight, &res.BlockHash,
		&res.Hash, &res.Type, &timestamp, &res.From, &res.To, &res.Amount, &res.Fee)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.TransactionDetail{}, err
	}
	res.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
	return res, nil
}

func (a *postgresAccessor) readStrValueCounts(rows *sql.Rows) ([]types.StrValueCount, error) {
	defer rows.Close()
	var res []types.StrValueCount
	for rows.Next() {
		item := types.StrValueCount{}
		if err := rows.Scan(&item.Value, &item.Count); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		a.log.Error("Unable to close db: %v", err)
	}
}
