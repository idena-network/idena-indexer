package db

import "database/sql"

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
	return cnt(a.db, query)
}

func (a *postgresAccessor) BurntCoinsInconsistencyCnt() (int, error) {
	const query1 = `SELECT count(*)
FROM coins c
         LEFT JOIN (SELECT block_height, sum(amount) total_amount FROM burnt_coins GROUP BY block_height) bc
                   ON bc.block_height = c.block_height
WHERE c.burnt <> coalesce(bc.total_amount, 0)`

	res1, err := cnt(a.db, query1)
	if err != nil {
		return 0, err
	}

	const query2 = `SELECT count(*)
FROM coins c
         LEFT JOIN
     (SELECT block_height,
             sum(coalesce(balance_new, 0) + coalesce(stake_new, 0) - coalesce(balance_old, 0) -
                 coalesce(stake_old, 0)) diff
      FROM balance_updates
      WHERE reason <> 3
      GROUP BY block_height) bu ON bu.block_height = c.block_height
         LEFT JOIN (SELECT sum(balance + stake) total, block_height
                    FROM mining_rewards
                    WHERE not proposer
                    GROUP BY block_height) comm_rew ON comm_rew.block_height = c.block_height
WHERE c.minted - c.burnt - coalesce(comm_rew.total, 0) <> coalesce(bu.diff, 0)
  AND c.block_height > 1`
	res2, err := cnt(a.db, query2)
	if err != nil {
		return 0, err
	}
	return res1 + res2, nil
}

func cnt(db *sql.DB, query string) (int, error) {
	var res int
	err := db.QueryRow(query).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}
