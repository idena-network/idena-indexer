package postgres

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/log"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

func NewPostgresAccessor(connStr string, scriptsDirPath string, logger log.Logger) db.Accessor {
	dbAccessor, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	dbAccessor.SetMaxOpenConns(25)
	dbAccessor.SetMaxIdleConns(25)
	dbAccessor.SetConnMaxLifetime(5 * time.Minute)

	return &postgresAccessor{
		db:      dbAccessor,
		queries: readQueries(scriptsDirPath, logger),
		log:     logger,
	}
}

func readQueries(scriptsDirPath string, log log.Logger) map[string]string {
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
		log.Debug(fmt.Sprintf("Read query %s from %s", queryName, scriptsDirPath))
	}
	return queries
}
