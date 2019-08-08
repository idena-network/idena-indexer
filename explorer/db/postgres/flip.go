package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	flipQuery             = "flip.sql"
	flipAnswersCountQuery = "flipAnswersCount.sql"
	flipAnswersQuery      = "flipAnswers.sql"
)

func (a *postgresAccessor) Flip(hash string) (types.Flip, error) {
	flip := types.Flip{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(flipQuery), hash).
		Scan(&flip.Author,
			&flip.Size,
			&timestamp,
			&flip.Answer,
			&flip.Status,
			&flip.Data,
			&flip.TxHash,
			&flip.BlockHash,
			&flip.BlockHeight,
			&flip.Epoch)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Flip{}, err
	}
	flip.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
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
	return a.readAnswers(rows)
}
