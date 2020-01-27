package db

import (
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"math/big"
)

const (
	selectEpochFlipsWithoutSizeQuery = "selectEpochFlipsWithoutSize.sql"
	selectFlipsToLoadContentQuery    = "selectFlipsToLoadContent.sql"
	saveFlipsContentQuery            = "saveFlipsContent.sql"
)

func (a *postgresAccessor) SaveFlipSize(flipCid string, size int) error {
	var flipId int64
	err := a.db.QueryRow(a.getQuery(updateFlipSizeQuery), size, flipCid).Scan(&flipId)
	return errors.Wrapf(err, "unable to save flip size (cid %s, size %d)", flipCid, size)
}

func (a *postgresAccessor) GetEpochFlipsWithoutSize(epoch uint64, limit int) (cids []string, err error) {
	rows, err := a.db.Query(a.getQuery(selectEpochFlipsWithoutSizeQuery), epoch, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var item string
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) GetFlipsToLoadContent(timestamp *big.Int, limit int) ([]*FlipToLoadContent, error) {
	rows, err := a.db.Query(a.getQuery(selectFlipsToLoadContentQuery), timestamp.Int64(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*FlipToLoadContent
	for rows.Next() {
		var item FlipToLoadContent
		err = rows.Scan(&item.Cid, &item.Key, &item.Attempts)
		if err != nil {
			return nil, err
		}
		res = append(res, &item)
	}
	return res, nil
}

func (a *postgresAccessor) SaveFlipsContent(failedFlips []*FailedFlipContent, flipsContent []*FlipContent) error {
	if len(failedFlips) == 0 && len(flipsContent) == 0 {
		return nil
	}
	_, err := a.db.Exec(a.getQuery(saveFlipsContentQuery),
		pq.Array(failedFlips),
		pq.Array(flipsContent))
	return errors.Wrap(err, "unable to save flips content")
}
