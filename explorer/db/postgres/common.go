package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

func (a *postgresAccessor) count(queryName string, args ...interface{}) (uint64, error) {
	var res uint64
	err := a.db.QueryRow(a.getQuery(queryName), args...).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (a *postgresAccessor) readInvites(rows *sql.Rows) ([]types.Invite, error) {
	defer rows.Close()
	var res []types.Invite
	for rows.Next() {
		item := types.Invite{}
		var timestamp int64
		var activationTimestamp int64
		if err := rows.Scan(&item.Hash, &item.Author, &timestamp, &item.ActivationHash, &item.ActivationAuthor, &activationTimestamp); err != nil {
			return nil, err
		}
		item.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		if activationTimestamp > 0 {
			at := common.TimestampToTime(big.NewInt(activationTimestamp))
			item.ActivationTimestamp = &at
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) readTxs(rows *sql.Rows) ([]types.TransactionSummary, error) {
	defer rows.Close()
	var res []types.TransactionSummary
	for rows.Next() {
		item := types.TransactionSummary{}
		var timestamp int64
		if err := rows.Scan(&item.Hash,
			&item.Type,
			&timestamp,
			&item.From,
			&item.To,
			&item.Amount,
			&item.Fee,
			&item.Size); err != nil {
			return nil, err
		}
		item.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) strValueCounts(queryName string, args ...interface{}) ([]types.StrValueCount, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	return a.readStrValueCounts(rows)
}

func (a *postgresAccessor) flips(queryName string, args ...interface{}) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.FlipSummary
	for rows.Next() {
		item := types.FlipSummary{}
		var timestamp int64
		words := types.FlipWords{}
		err := rows.Scan(&item.Cid,
			&item.Size,
			&item.Author,
			&item.Epoch,
			&item.Status,
			&item.Answer,
			&item.ShortRespCount,
			&item.LongRespCount,
			&timestamp,
			&item.Icon,
			&words.Word1.Index,
			&words.Word1.Name,
			&words.Word1.Desc,
			&words.Word2.Index,
			&words.Word2.Name,
			&words.Word2.Desc)
		if err != nil {
			return nil, err
		}
		item.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		if !words.IsEmpty() {
			item.Words = &words
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) readEpochIdentitySummaries(rows *sql.Rows) ([]types.EpochIdentitySummary, error) {
	defer rows.Close()
	var res []types.EpochIdentitySummary
	for rows.Next() {
		item := types.EpochIdentitySummary{}
		err := rows.Scan(&item.Address,
			&item.Epoch,
			&item.State,
			&item.PrevState,
			&item.Approved,
			&item.Missed,
			&item.ShortAnswers.Point,
			&item.ShortAnswers.FlipsCount,
			&item.TotalShortAnswers.Point,
			&item.TotalShortAnswers.FlipsCount,
			&item.LongAnswers.Point,
			&item.LongAnswers.FlipsCount,
			&item.RequiredFlips,
			&item.MadeFlips)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) readAnswers(rows *sql.Rows) ([]types.Answer, error) {
	defer rows.Close()
	var res []types.Answer
	for rows.Next() {
		item := types.Answer{}
		err := rows.Scan(&item.Cid, &item.Address, &item.RespAnswer, &item.FlipAnswer, &item.FlipStatus, &item.Point)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) coins(queryName string, args ...interface{}) (types.AllCoins, error) {
	res := types.AllCoins{
		Balance: types.Coins{},
		Stake:   types.Coins{},
	}
	err := a.db.QueryRow(a.getQuery(queryName), args...).
		Scan(&res.Balance.Burnt,
			&res.Balance.Minted,
			&res.Balance.Total,
			&res.Stake.Burnt,
			&res.Stake.Minted,
			&res.Stake.Total)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.AllCoins{}, err
	}
	return res, nil
}

func (a *postgresAccessor) rewards(queryName string, args ...interface{}) ([]types.Reward, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Reward
	for rows.Next() {
		item := types.Reward{}
		if err := rows.Scan(&item.Address, &item.Epoch, &item.Balance, &item.Stake, &item.Type); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
