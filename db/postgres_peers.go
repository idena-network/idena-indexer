package db

import "time"

const savePeersCountQuery = "savePeersCount.sql"

func (a *postgresAccessor) SavePeersCount(count int, timestamp time.Time) error {
	t := timestamp.UTC().Unix()
	_, err := a.db.Exec(a.getQuery(savePeersCountQuery), t, count)
	return getResultError(err)
}
