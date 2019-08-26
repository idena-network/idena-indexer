package db

import (
	"database/sql"
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
	if err := a.init(); err != nil {
		panic(err)
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
