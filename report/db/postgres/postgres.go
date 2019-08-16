package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/report/types"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
	logger  log.Logger
}

const (
	epochsCountQuery   = "epochsCount.sql"
	epochFlipCidsQuery = "epochFlipCids.sql"
	flipContentQuery   = "flipContent.sql"
	flipPicsQuery      = "flipPics.sql"
	flipPicOrdersQuery = "flipPicOrders.sql"
)

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) EpochsCount() (uint64, error) {
	var res uint64
	err := a.db.QueryRow(a.getQuery(epochsCountQuery)).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (a *postgresAccessor) FlipCids(epoch uint64) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipCidsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) FlipContent(cid string) (types.FlipContent, error) {
	var dataId int64
	err := a.db.QueryRow(a.getQuery(flipContentQuery), cid).Scan(&dataId)
	if err == sql.ErrNoRows {
		err = errors.New("no flip content in db")
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

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		a.logger.Error("Unable to close db: %v", err)
	}
}
