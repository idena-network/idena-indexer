package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/explorer/service"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type postgresAccessor struct {
	db                          *sql.DB
	networkSizeLoader           service.NetworkSizeLoader
	estimatedOracleRewardsCache *estimatedOracleRewardsService
	queries                     map[string]string
	log                         log.Logger
}

const (
	transactionQuery          = "transaction.sql"
	transactionRawQuery       = "transactionRaw.sql"
	isAddressQuery            = "isAddress.sql"
	isBlockHashQuery          = "isBlockHash.sql"
	isBlockHeightQuery        = "isBlockHeight.sql"
	isEpochQuery              = "isEpoch.sql"
	isFlipQuery               = "isFlip.sql"
	isTxQuery                 = "isTx.sql"
	coinsBurntAndMintedQuery  = "coinsBurntAndMinted.sql"
	coinsTotalQuery           = "coinsTotal.sql"
	circulatingSupplyQuery    = "circulatingSupply.sql"
	activeAddressesCountQuery = "activeAddressesCount.sql"
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
			Name:     "Identity",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Identity/%s", value),
			NameOld:  "Identity",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Identity/%s", value),
		})
		res = append(res, types.Entity{
			Name:     "Address",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Address/%s", value),
			NameOld:  "Address",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Address/%s", value),
		})
	}

	if exists, err := a.isEntity(value, isBlockHashQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name:     "Block",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Block/%s", value),
			NameOld:  "Block",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Block/%s", value),
		})
	} else if isNum {
		if exists, err := a.isEntity(value, isBlockHeightQuery); err != nil {
			return nil, err
		} else if exists {
			res = append(res, types.Entity{
				Name:     "Block",
				Value:    value,
				Ref:      fmt.Sprintf("/api/Block/%s", value),
				NameOld:  "Block",
				ValueOld: value,
				RefOld:   fmt.Sprintf("/api/Block/%s", value),
			})
		}
	}

	if isNum {
		if exists, err := a.isEntity(value, isEpochQuery); err != nil {
			return nil, err
		} else if exists {
			res = append(res, types.Entity{
				Name:     "Epoch",
				Value:    value,
				Ref:      fmt.Sprintf("/api/Epoch/%s", value),
				NameOld:  "Epoch",
				ValueOld: value,
				RefOld:   fmt.Sprintf("/api/Epoch/%s", value),
			})
		}
	}

	if exists, err := a.isEntity(value, isFlipQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name:     "Flip",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Flip/%s", value),
			NameOld:  "Flip",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Flip/%s", value),
		})
	}

	if exists, err := a.isEntity(value, isTxQuery); err != nil {
		return nil, err
	} else if exists {
		res = append(res, types.Entity{
			Name:     "Transaction",
			Value:    value,
			Ref:      fmt.Sprintf("/api/Transaction/%s", value),
			NameOld:  "Transaction",
			ValueOld: value,
			RefOld:   fmt.Sprintf("/api/Transaction/%s", value),
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

func (a *postgresAccessor) CirculatingSupply(addressesToExclude []string) (decimal.Decimal, error) {
	var totalBalance, totalStake decimal.Decimal
	var err error
	if len(addressesToExclude) == 0 {
		err = a.db.QueryRow(a.getQuery(coinsTotalQuery)).Scan(&totalBalance, &totalStake)
	} else {
		err = a.db.QueryRow(a.getQuery(circulatingSupplyQuery), pq.Array(addressesToExclude)).Scan(&totalBalance)
	}
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return decimal.Decimal{}, err
	}
	return totalBalance, nil
}

func (a *postgresAccessor) ActiveAddressesCount(afterTime time.Time) (uint64, error) {
	return a.count(activeAddressesCountQuery, afterTime.Unix())
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

func (a *postgresAccessor) Transaction(hash string) (types.TransactionDetail, error) {
	res := types.TransactionDetail{}
	var timestamp int64
	var transfer NullDecimal
	var becomeOnline sql.NullBool
	err := a.db.QueryRow(a.getQuery(transactionQuery), hash).Scan(&res.Epoch, &res.BlockHeight, &res.BlockHash,
		&res.Hash, &res.Type, &timestamp, &res.From, &res.To, &res.Amount, &res.Tips, &res.MaxFee, &res.Fee, &res.Size,
		&transfer, &becomeOnline)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.TransactionDetail{}, err
	}
	res.Timestamp = timestampToTimeUTC(timestamp)
	if transfer.Valid {
		res.Transfer = &transfer.Decimal
	}
	res.Data = readTxSpecificData(res.Type, transfer, becomeOnline)
	return res, nil
}

func (a *postgresAccessor) TransactionRaw(hash string) (hexutil.Bytes, error) {
	var res hexutil.Bytes
	err := a.db.QueryRow(a.getQuery(transactionRawQuery), hash).Scan(&res)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return hexutil.Bytes{}, err
	}
	return res, nil
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		a.log.Error("Unable to close db: %v", err)
	}
}
