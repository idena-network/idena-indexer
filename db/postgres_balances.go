package db

func (a *postgresAccessor) Balances() ([]Balance, error) {
	rows, err := a.db.Query("select a.address, b.balance from balances b join addresses a on a.id = b.address_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Balance
	for rows.Next() {
		var item Balance
		err = rows.Scan(&item.Address, &item.Balance)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (a *postgresAccessor) LatestBalanceUpdates() ([]Balance, error) {
	rows, err := a.db.Query("select distinct on (a.address) a.address, bu.balance_new from balance_updates bu join addresses a on a.id = bu.address_id order by a.address, bu.id desc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Balance
	for rows.Next() {
		var item Balance
		err = rows.Scan(&item.Address, &item.Balance)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (a *postgresAccessor) BalanceUpdateGapCnt() (int, error) {
	const query = `select count(*)
from balance_updates bu1
         left join balance_updates bu2 on bu1.address_id = bu2.address_id and bu2.id = (select max(id)
                                                                                        from balance_updates bu3
                                                                                        where bu3.address_id = bu1.address_id
                                                                                          and bu3.id < bu1.id)
where bu1.id > 1 and coalesce(bu2.balance_new, 0) <> coalesce(bu1.balance_old, 0)
   or coalesce(bu2.stake_new, 0) <> coalesce(bu1.stake_old, 0)
   or coalesce(bu2.penalty_seconds_new, 0) <> coalesce(bu1.penalty_seconds_old, 0)`

	var res int
	err := a.db.QueryRow(query).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}
