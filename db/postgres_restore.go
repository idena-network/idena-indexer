package db

func (a *postgresAccessor) SaveRestoredData(data *RestoredData) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := a.saveBalances(tx, data.Balances); err != nil {
		return err
	}

	if err := a.saveBirthdays(tx, data.Birthdays); err != nil {
		return err
	}

	return tx.Commit()
}
