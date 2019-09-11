package db

import (
	"database/sql"
	"fmt"
	indexerDb "github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"strings"
)

type postgresAccessor struct {
	db        *sql.DB
	oldSchema string
	queries   map[string]string
}

func NewPostgresAccessor(connStr string, oldSchema string, scriptsDirPath string) Accessor {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return &postgresAccessor{
		db:        db,
		oldSchema: oldSchema,
		queries:   indexerDb.ReadQueries(scriptsDirPath),
	}
}

const (
	migrateQuery = "migrate.sql"
)

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) MigrateTo(height uint64) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	migrationScript := strings.ReplaceAll(a.getQuery(migrateQuery), "OLD_SCHEMA_TAG", a.oldSchema)
	queries := strings.Split(migrationScript, ";")
	for _, query := range queries {
		if len(strings.TrimSpace(query)) == 0 {
			continue
		}
		var err error
		if strings.Contains(query, "$1") {
			_, err = tx.Exec(query, height)
		} else {
			_, err = tx.Exec(query)
		}
		if err != nil {
			return errors.Wrapf(err, "unable to migrate data from schema %s, query: %s", a.oldSchema, query)
		}
	}
	return tx.Commit()
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}
