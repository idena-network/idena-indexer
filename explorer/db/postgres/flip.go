package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

const (
	flipQuery             = "flip.sql"
	flipAnswersCountQuery = "flipAnswersCount.sql"
	flipAnswersQuery      = "flipAnswers.sql"
	flipContentQuery      = "flipContent.sql"
	flipPicsQuery         = "flipPics.sql"
	flipPicOrdersQuery    = "flipPicOrders.sql"
)

func (a *postgresAccessor) Flip(hash string) (types.Flip, error) {
	flip := types.Flip{}
	words := types.FlipWords{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(flipQuery), hash).
		Scan(&flip.Author,
			&flip.Size,
			&timestamp,
			&flip.Answer,
			&flip.WrongWords,
			&flip.WrongWordsVotes,
			&flip.Status,
			&flip.TxHash,
			&flip.BlockHash,
			&flip.BlockHeight,
			&flip.Epoch,
			&words.Word1.Index,
			&words.Word1.Name,
			&words.Word1.Desc,
			&words.Word2.Index,
			&words.Word2.Name,
			&words.Word2.Desc)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Flip{}, err
	}
	flip.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
	if !words.IsEmpty() {
		flip.Words = &words
	}
	return flip, nil
}

func (a *postgresAccessor) FlipContent(hash string) (types.FlipContent, error) {
	var dataId int64
	err := a.db.QueryRow(a.getQuery(flipContentQuery), hash).Scan(&dataId)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.FlipContent{}, err
	}
	res := types.FlipContent{}
	if res.Pics, err = a.flipPics(dataId); err != nil {
		return types.FlipContent{}, err
	}
	if res.LeftOrder, res.RightOrder, err = a.flipPicOrders(dataId); err != nil {
		return types.FlipContent{}, err
	}
	return res, nil
}

func (a *postgresAccessor) flipPics(dataId int64) ([]hexutil.Bytes, error) {
	rows, err := a.db.Query(a.getQuery(flipPicsQuery), dataId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []hexutil.Bytes
	for rows.Next() {
		item := hexutil.Bytes{}
		err := rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) flipPicOrders(dataId int64) ([]uint16, []uint16, error) {
	rows, err := a.db.Query(a.getQuery(flipPicOrdersQuery), dataId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var leftOrder, rightOrder []uint16
	for rows.Next() {
		var answerIndex, flipPicIndex uint16
		err := rows.Scan(&answerIndex, &flipPicIndex)
		if err != nil {
			return nil, nil, err
		}
		switch answerIndex {
		case 0:
			leftOrder = append(leftOrder, flipPicIndex)
		case 1:
			rightOrder = append(rightOrder, flipPicIndex)
		default:
			log.Warn("Unknown answer index", "flipDataId", dataId, "index", answerIndex)
		}
	}
	return leftOrder, rightOrder, nil
}

func (a *postgresAccessor) FlipAnswersCount(hash string, isShort bool) (uint64, error) {
	return a.count(flipAnswersCountQuery, hash, isShort)
}

func (a *postgresAccessor) FlipAnswers(hash string, isShort bool, startIndex uint64, count uint64) ([]types.Answer, error) {
	rows, err := a.db.Query(a.getQuery(flipAnswersQuery), hash, isShort, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readAnswers(rows)
}
