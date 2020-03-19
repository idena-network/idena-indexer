package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

const (
	activationTx   = "ActivationTx"
	killTx         = "KillTx"
	killInviteeTx  = "KillInviteeTx"
	onlineStatusTx = "OnlineStatusTx"
)

func timestampToTimeUTC(timestamp int64) time.Time {
	return common.TimestampToTime(big.NewInt(timestamp)).UTC()
}

type NullDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

func (n *NullDecimal) Scan(value interface{}) error {
	n.Valid = value != nil
	n.Decimal = decimal.Decimal{}
	if n.Valid {
		return n.Decimal.Scan(value)
	}
	return nil
}

func (a *postgresAccessor) count(queryName string, args ...interface{}) (uint64, error) {
	var res uint64
	err := a.db.QueryRow(a.getQuery(queryName), args...).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func readInvites(rows *sql.Rows) ([]types.Invite, error) {
	defer rows.Close()
	var res []types.Invite
	for rows.Next() {
		item := types.Invite{}
		var timestamp, activationTimestamp, killInviteeTimestamp int64
		if err := rows.Scan(
			&item.Hash,
			&item.Author,
			&timestamp,
			&item.Epoch,
			&item.ActivationHash,
			&item.ActivationAuthor,
			&activationTimestamp,
			&item.State,
			&item.KillInviteeHash,
			&killInviteeTimestamp,
			&item.KillInviteeEpoch,
		); err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		if activationTimestamp > 0 {
			t := timestampToTimeUTC(activationTimestamp)
			item.ActivationTimestamp = &t
		}
		if killInviteeTimestamp > 0 {
			t := timestampToTimeUTC(killInviteeTimestamp)
			item.KillInviteeTimestamp = &t
		}
		res = append(res, item)
	}
	return res, nil
}

func readTxs(rows *sql.Rows) ([]types.TransactionSummary, error) {
	defer rows.Close()
	var res []types.TransactionSummary
	for rows.Next() {
		item := types.TransactionSummary{}
		var timestamp int64
		var transfer NullDecimal
		var becomeOnline sql.NullBool
		if err := rows.Scan(
			&item.Hash,
			&item.Type,
			&timestamp,
			&item.From,
			&item.To,
			&item.Amount,
			&item.Tips,
			&item.MaxFee,
			&item.Fee,
			&item.Size,
			&transfer,
			&becomeOnline,
		); err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		if transfer.Valid {
			item.Transfer = &transfer.Decimal
		}
		item.Data = readTxSpecificData(item.Type, transfer, becomeOnline)
		res = append(res, item)
	}
	return res, nil
}

func readTxSpecificData(txType string, transfer NullDecimal, becomeOnline sql.NullBool) interface{} {
	var res interface{}
	switch txType {
	case activationTx:
		if data := readActivationTxSpecificData(transfer); data != nil {
			res = data
		}
	case killTx:
		if data := readKillTxSpecificData(transfer); data != nil {
			res = data
		}
	case killInviteeTx:
		if data := readKillInviteeTxSpecificData(transfer); data != nil {
			res = data
		}
	case onlineStatusTx:
		if data := readOnlineStatusTxSpecificData(becomeOnline); data != nil {
			res = data
		}
	}
	return res
}

func readActivationTxSpecificData(transfer NullDecimal) *types.ActivationTxSpecificData {
	if !transfer.Valid {
		return nil
	}
	return &types.ActivationTxSpecificData{
		Transfer: &transfer.Decimal,
	}
}

func readKillTxSpecificData(transfer NullDecimal) *types.KillTxSpecificData {
	return readActivationTxSpecificData(transfer)
}

func readKillInviteeTxSpecificData(transfer NullDecimal) *types.KillInviteeTxSpecificData {
	return readActivationTxSpecificData(transfer)
}

func readOnlineStatusTxSpecificData(becomeOnline sql.NullBool) *types.OnlineStatusTxSpecificData {
	if !becomeOnline.Valid {
		return nil
	}
	return &types.OnlineStatusTxSpecificData{
		BecomeOnline: becomeOnline.Bool,
	}
}

func (a *postgresAccessor) strValueCounts(queryName string, args ...interface{}) ([]types.StrValueCount, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	return readStrValueCounts(rows)
}

func readStrValueCounts(rows *sql.Rows) ([]types.StrValueCount, error) {
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

func (a *postgresAccessor) nullableBoolValueCounts(queryName string, args ...interface{}) ([]types.NullableBoolValueCount, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return nil, err
	}
	return readNullableBoolValueCounts(rows)
}

func readNullableBoolValueCounts(rows *sql.Rows) ([]types.NullableBoolValueCount, error) {
	defer rows.Close()
	var res []types.NullableBoolValueCount
	for rows.Next() {
		item := types.NullableBoolValueCount{}
		nullBool := sql.NullBool{}
		if err := rows.Scan(&nullBool, &item.Count); err != nil {
			return nil, err
		}
		if nullBool.Valid {
			item.Value = &nullBool.Bool
		}
		res = append(res, item)
	}
	return res, nil
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
			&item.WrongWords,
			&item.WrongWordsVotes,
			&item.ShortRespCount,
			&item.LongRespCount,
			&timestamp,
			&item.Icon,
			&words.Word1.Index,
			&words.Word1.Name,
			&words.Word1.Desc,
			&words.Word2.Index,
			&words.Word2.Name,
			&words.Word2.Desc,
			&item.WithPrivatePart)
		if err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		if !words.IsEmpty() {
			item.Words = &words
		}
		res = append(res, item)
	}
	return res, nil
}

func readEpochIdentitySummaries(rows *sql.Rows) ([]types.EpochIdentitySummary, error) {
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

func readAnswers(rows *sql.Rows) ([]types.Answer, error) {
	defer rows.Close()
	var res []types.Answer
	for rows.Next() {
		item := types.Answer{}
		err := rows.Scan(&item.Cid,
			&item.Address,
			&item.RespAnswer,
			&item.RespWrongWords,
			&item.FlipAnswer,
			&item.FlipWrongWords,
			&item.FlipStatus,
			&item.Point)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) coins(queryName string, args ...interface{}) (types.AllCoins, error) {
	res := types.AllCoins{}
	err := a.db.QueryRow(a.getQuery(queryName), args...).
		Scan(&res.Burnt,
			&res.Minted,
			&res.TotalBalance,
			&res.TotalStake)
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
		if err := rows.Scan(&item.Address, &item.Epoch, &item.BlockHeight, &item.Balance, &item.Stake, &item.Type); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func readBadAuthors(rows *sql.Rows) ([]types.BadAuthor, error) {
	defer rows.Close()
	var res []types.BadAuthor
	for rows.Next() {
		item := types.BadAuthor{}
		err := rows.Scan(
			&item.Address,
			&item.Epoch,
			&item.WrongWords,
			&item.PrevState,
			&item.State,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) adjacentStrValues(queryName string, value string) (types.AdjacentStrValues, error) {
	res := types.AdjacentStrValues{}
	err := a.db.QueryRow(a.getQuery(queryName), value).Scan(
		&res.Prev.Value,
		&res.Prev.Cycled,
		&res.Next.Value,
		&res.Next.Cycled,
	)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.AdjacentStrValues{}, err
	}
	return res, nil
}
