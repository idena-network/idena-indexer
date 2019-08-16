package postgres

import "github.com/idena-network/idena-indexer/explorer/types"

const (
	balancesCountQuery = "balancesCount.sql"
	balancesQuery      = "balances.sql"
)

func (a *postgresAccessor) BalancesCount() (uint64, error) {
	return a.count(balancesCountQuery)
}

func (a *postgresAccessor) Balances(startIndex uint64, count uint64) ([]types.Balance, error) {
	rows, err := a.db.Query(a.getQuery(balancesQuery), startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Balance
	for rows.Next() {
		item := types.Balance{}
		err = rows.Scan(&item.Address,
			&item.Balance,
			&item.Stake)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
