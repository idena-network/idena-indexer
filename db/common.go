package db

import (
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

func ReadQueries(scriptsDirPath string) map[string]string {
	queries := make(map[string]string)
	readDirQueries(scriptsDirPath, queries, "")
	var namesToDelete []string
	var queriesToAddToInit []string
	for name, query := range queries {
		if !strings.HasPrefix(name, "init/") || strings.HasSuffix(name, initQuery) {
			continue
		}
		namesToDelete = append(namesToDelete, name)
		queriesToAddToInit = append(queriesToAddToInit, query)
	}
	for _, namesToDelete := range namesToDelete {
		delete(queries, namesToDelete)
	}
	for _, queriesToAddToInit := range queriesToAddToInit {
		queries[initQuery] += queriesToAddToInit
	}
	return queries
}

func readDirQueries(dirPath string, queriesByName map[string]string, namePrefix string) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			readDirQueries(path.Join(dirPath, file.Name()), queriesByName, fmt.Sprintf("%v%v/", namePrefix, file.Name()))
			continue
		}
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		bytes, err := ioutil.ReadFile(filepath.Join(dirPath, file.Name()))
		if err != nil {
			panic(err)
		}
		queryName := fmt.Sprintf("%v%v", namePrefix, file.Name())
		query := string(bytes)
		queriesByName[queryName] = query
		log.Debug(fmt.Sprintf("Read query %s from %s", queryName, dirPath))
	}
}
