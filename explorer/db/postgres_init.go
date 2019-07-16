package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func NewPostgresAccessor(connStr string, scriptsDirPath string, logger log.Logger) Accessor {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	return &postgresAccessor{
		db:      db,
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
		log.Info(fmt.Sprintf("Read query %s", queryName))
	}
	return queries
}
