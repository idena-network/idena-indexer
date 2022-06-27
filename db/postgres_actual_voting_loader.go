package db

import "github.com/idena-network/idena-go/common"

const (
	selectActualOracleVotingsQuery = "selectActualOracleVotings.sql"
)

func (a *postgresAccessor) ActualOracleVotings() ([]common.Address, error) {
	rows, err := a.db.Query(a.getQuery(selectActualOracleVotingsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []common.Address
	for rows.Next() {
		var item string
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, common.HexToAddress(item))
	}
	return res, nil
}
