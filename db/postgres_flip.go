package db

import "github.com/pkg/errors"

const selectEpochFlipsWithoutSizeQuery = "selectEpochFlipsWithoutSize.sql"

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
