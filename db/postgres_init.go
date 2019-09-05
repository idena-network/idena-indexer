package db

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/log"
	"time"
)

func NewPostgresAccessor(connStr string, scriptsDirPath string) Accessor {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	a := &postgresAccessor{
		db:      db,
		queries: ReadQueries(scriptsDirPath),
	}
	for {
		if err := a.init(); err != nil {
			log.Error("Unable to initialize postgres connection", "err", err)
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	return a
}

func (a *postgresAccessor) init() error {
	if err := a.db.Ping(); err != nil {
		return err
	}
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(a.getQuery(initQuery))
	if err != nil {
		return err
	}

	return tx.Commit()
}
