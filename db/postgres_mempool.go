package db

import (
	"database/sql"
	"github.com/pkg/errors"
)

const (
	insertFlipKeyTimestampQuery            = "insertFlipKeyTimestamp.sql"
	insertAnswersHashTxTimestampQuery      = "insertAnswersHashTxTimestamp.sql"
	selectFlipKeyTimestampCountQuery       = "selectFlipKeyTimestampCount.sql"
	selectAnswersHashTxTimestampCountQuery = "selectAnswersHashTxTimestampCount.sql"
)

func (a *postgresAccessor) SaveMemPoolData(data *MemPoolData) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	tx, err := a.db.Begin()
	if err != nil {
		return errors.Wrap(err, "unable to begin tx to save mem pool data")
	}
	defer tx.Rollback()

	if err := a.saveMemPoolActionTimestamps(tx,
		selectFlipKeyTimestampCountQuery,
		insertFlipKeyTimestampQuery,
		data.FlipKeyTimestamps); err != nil {
		return err
	}

	if err := a.saveMemPoolActionTimestamps(tx,
		selectAnswersHashTxTimestampCountQuery,
		insertAnswersHashTxTimestampQuery,
		data.AnswersHashTxTimestamps); err != nil {
		return err
	}

	return errors.Wrap(tx.Commit(), "unable to commit mem pool data (%v)")
}

func (a *postgresAccessor) saveMemPoolActionTimestamps(tx *sql.Tx, selectQueryName, insertQueryName string, timestamps []*MemPoolActionTimestamp) error {
	for _, t := range timestamps {
		var count int
		if err := tx.QueryRow(a.getQuery(selectQueryName), t.Address, t.Epoch).Scan(&count); err != nil {
			return errors.Wrapf(err, "unable to get count (queryName: %s)", selectQueryName)
		}
		if count > 0 {
			// Do not save if record already exists
			continue
		}
		_, err := tx.Exec(a.getQuery(insertQueryName),
			t.Address,
			t.Epoch,
			t.Time.Int64())

		if err != nil {
			return errors.Wrapf(err, "unable to save action timestamp (%v, queryName: %s)", t, insertQueryName)
		}
	}
	return nil
}
