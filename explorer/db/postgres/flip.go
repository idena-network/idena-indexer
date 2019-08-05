package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	flipQuery             = "flip.sql"
	flipAnswersCountQuery = "flipAnswersCount.sql"
	flipAnswersQuery      = "flipAnswers.sql"
)

func (a *postgresAccessor) Flip(hash string) (types.Flip, error) {
	flip := types.Flip{}
	var id uint64
	err := a.db.QueryRow(a.getQuery(flipQuery), hash).Scan(&id, &flip.Answer, &flip.Status, &flip.Data)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Flip{}, err
	}
	return flip, nil
}

func (a *postgresAccessor) FlipAnswersCount(hash string, isShort bool) (uint64, error) {
	return a.count(flipAnswersCountQuery, hash, isShort)
}

func (a *postgresAccessor) FlipAnswers(hash string, isShort bool, startIndex uint64, count uint64) ([]types.Answer, error) {
	rows, err := a.db.Query(a.getQuery(flipAnswersQuery), hash, isShort, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Answer
	for rows.Next() {
		item := types.Answer{}
		err = rows.Scan(&item.Address, &item.RespAnswer, &item.FlipAnswer)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
