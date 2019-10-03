package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
	"strconv"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
	log     log.Logger
}

const (
	addressQuery             = "address.sql"
	transactionQuery         = "transaction.sql"
	currentEpochQuery        = "currentEpoch.sql"
	isAddressQuery           = "isAddress.sql"
	isBlockHashQuery         = "isBlockHash.sql"
	isBlockHeightQuery       = "isBlockHeight.sql"
	isEpochQuery             = "isEpoch.sql"
	isFlipQuery              = "isFlip.sql"
	isTxQuery                = "isTx.sql"
	coinsBurntAndMintedQuery = "coinsBurntAndMinted.sql"
	coinsTotalQuery          = "coinsTotal.sql"
)

var NoDataFound = errors.New("no data found")

func (a *postgresAccessor) Search(value string) ([]types.Entity, error) {
	var isNum bool
	if _, err := strconv.ParseUint(value, 10, 64); err == nil {
		isNum = true
	}
	var res []types.Entity
	if exists, err := a.isEntity(value, isAddressQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name: "Identity",
			Ref:  fmt.Sprintf("/api/Identity/%s", value),
		})
		res = append(res, types.Entity{
			Name: "Address",
			Ref:  fmt.Sprintf("/api/Address/%s", value),
		})
	}

	if exists, err := a.isEntity(value, isBlockHashQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name: "Block",
			Ref:  fmt.Sprintf("/api/Block/%s", value),
		})
	} else if isNum {
		if exists, err := a.isEntity(value, isBlockHeightQuery); err != nil {
			return nil, err
		} else if exists {
			res = append(res, types.Entity{
				Name: "Block",
				Ref:  fmt.Sprintf("/api/Block/%s", value),
			})
		}
	}

	if isNum {
		if exists, err := a.isEntity(value, isEpochQuery); err != nil {
			return nil, err
		} else if exists {
			res = append(res, types.Entity{
				Name: "Epoch",
				Ref:  fmt.Sprintf("/api/Epoch/%s", value),
			})
		}
	}

	if exists, err := a.isEntity(value, isFlipQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name: "Flip",
			Ref:  fmt.Sprintf("/api/Flip/%s", value),
		})
	}

	if exists, err := a.isEntity(value, isTxQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name: "Transaction",
			Ref:  fmt.Sprintf("/api/Transaction/%s", value),
		})
	}

	return res, nil
}

func (a *postgresAccessor) Coins() (types.AllCoins, error) {
	res := types.AllCoins{}
	err := a.db.QueryRow(a.getQuery(coinsTotalQuery)).Scan(&res.TotalBalance, &res.TotalStake)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.AllCoins{}, err
	}
	err = a.db.QueryRow(a.getQuery(coinsBurntAndMintedQuery)).
		Scan(&res.Burnt,
			&res.Minted)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.AllCoins{}, err
	}
	return res, nil
}

func (a *postgresAccessor) isEntity(value, queryName string) (bool, error) {
	var exists bool
	err := a.db.QueryRow(a.getQuery(queryName), value).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

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
		&res.Hash, &res.Type, &timestamp, &res.From, &res.To, &res.Amount, &res.Tips, &res.MaxFee, &res.Fee, &res.Size)
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
