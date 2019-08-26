package db

import (
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func ReadQueries(scriptsDirPath string) map[string]string {
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
		log.Info(fmt.Sprintf("Read query %s from %s", queryName, scriptsDirPath))
	}
	return queries
}
