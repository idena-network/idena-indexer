package db

const (
	saveMemPoolDataQuery = "saveMemPoolData.sql"
)

func (a *postgresAccessor) SaveMemPoolData(data *MemPoolData) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	_, err := a.db.Exec(a.getQuery(saveMemPoolDataQuery), data)
	return getResultError(err)
}
