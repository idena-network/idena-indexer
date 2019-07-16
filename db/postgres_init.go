package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func NewPostgresAccessor(connStr string, scriptsDirPath string) Accessor {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	a := &postgresAccessor{
		db:      db,
		queries: readQueries(scriptsDirPath),
	}
	if err := a.init(); err != nil {
		panic(err)
	}
	return a
}

func readQueries(scriptsDirPath string) map[string]string {
	files, err := ioutil.ReadDir(scriptsDirPath)
	if err != nil {
		panic(err)
	}
	queries := make(map[string]string)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		bytes, err := ioutil.ReadFile(filepath.Join(scriptsDirPath, file.Name()))
		if err != nil {
			panic(err)
		}
		queryName := file.Name()
		query := string(bytes)
		queries[queryName] = query
		log.Info(fmt.Sprintf("Read query %s", queryName))
	}
	return queries
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

	_, err = tx.Exec(a.queries[initQuery])
	if err != nil {
		return err
	}

	return tx.Commit()
}
