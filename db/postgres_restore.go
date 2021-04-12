package db

import (
	"database/sql"
	"github.com/pkg/errors"
)

const saveRestoredDataQuery = "saveRestoredData.sql"

func (a *postgresAccessor) SaveRestoredData(data *RestoredData) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := a.saveBalances(tx, 0, data.Balances, nil, nil); err != nil {
		return err
	}

	if err := a.saveBirthdays(tx, data.Birthdays); err != nil {
		return err
	}

	if err := a.saveRestoredData(tx, data); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *postgresAccessor) saveRestoredData(tx *sql.Tx, data *RestoredData) error {
	postgresData := getRestoredData(data)
	if _, err := tx.Exec(a.getQuery(saveRestoredDataQuery), postgresData); err != nil {
		return errors.Wrap(err, "unable to save restored data")
	}
	return nil
}
