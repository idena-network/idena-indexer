package postgres

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	addressPenaltiesCountQuery = "addressPenaltiesCount.sql"
	addressPenaltiesQuery      = "addressPenalties.sql"
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
