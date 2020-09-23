package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
)

const (
	flipQuery                   = "flip.sql"
	flipAnswersQuery            = "flipAnswers.sql"
	flipPicsQuery               = "flipPics.sql"
	flipPicOrdersQuery          = "flipPicOrders.sql"
	flipEpochAdjacentFlipsQuery = "flipEpochAdjacentFlips.sql"
)

func (a *postgresAccessor) Flip(hash string) (types.Flip, error) {
	flip := types.Flip{}
	words := types.FlipWords{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(flipQuery), hash).
		Scan(
			&flip.Author,
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
			&words.Word2.Desc,
			&flip.WithPrivatePart,
			&flip.Grade,
		)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Flip{}, err
	}
	flip.Timestamp = timestampToTimeUTC(timestamp)
	if !words.IsEmpty() {
		flip.Words = &words
	}
	return flip, nil
}

func (a *postgresAccessor) FlipContent(hash string) (types.FlipContent, error) {
	res := types.FlipContent{}
	if pics, err := a.flipPics(hash); err != nil {
		return types.FlipContent{}, err
	} else if len(pics) == 0 {
		return types.FlipContent{}, NoDataFound
	} else {
		res.Pics = pics
		res.PicsOld = pics
	}
	if leftOrder, rightOrder, err := a.flipPicOrders(hash); err != nil {
		return types.FlipContent{}, err
	} else {
		res.LeftOrder, res.RightOrder = leftOrder, rightOrder
		res.LeftOrderOld, res.RightOrderOld = leftOrder, rightOrder
	}
	return res, nil
}

func (a *postgresAccessor) flipPics(hash string) ([]hexutil.Bytes, error) {
	rows, err := a.db.Query(a.getQuery(flipPicsQuery), hash)
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

func (a *postgresAccessor) flipPicOrders(hash string) ([]uint16, []uint16, error) {
	rows, err := a.db.Query(a.getQuery(flipPicOrdersQuery), hash)
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
			log.Warn("Unknown answer index", "cid", hash, "index", answerIndex)
		}
	}
	return leftOrder, rightOrder, nil
}

func (a *postgresAccessor) FlipAnswers(hash string, isShort bool) ([]types.Answer, error) {
	rows, err := a.db.Query(a.getQuery(flipAnswersQuery), hash, isShort)
	if err != nil {
		return nil, err
	}
	return readAnswers(rows)
}

func (a *postgresAccessor) FlipEpochAdjacentFlips(hash string) (types.AdjacentStrValues, error) {
	return a.adjacentStrValues(flipEpochAdjacentFlipsQuery, hash)
}
