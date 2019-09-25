package db

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/import/words"
	"github.com/idena-network/idena-indexer/log"
	"time"
)

func NewPostgresAccessor(connStr string, scriptsDirPath string, wordsLoader words.Loader) Accessor {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	a := &postgresAccessor{
		db:      db,
		queries: ReadQueries(scriptsDirPath),
	}
	for {
		if err := a.init(wordsLoader); err != nil {
			log.Error("Unable to initialize postgres connection", "err", err)
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	return a
}

func (a *postgresAccessor) init(wordsLoader words.Loader) error {
	if err := a.db.Ping(); err != nil {
		return err
	}
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(a.getQuery(initQuery)); err != nil {
		return err
	}

	if err := initWords(tx, wordsLoader); err != nil {
		return err
	}

	return tx.Commit()
}

func initWords(tx *sql.Tx, loader words.Loader) error {
	var count int64
	err := tx.QueryRow("select count(*) from words_dictionary").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	importedWords, err := loader.LoadWords()
	if err != nil {
		return err
	}
	for i, word := range importedWords {
		if _, err := tx.Exec("insert into words_dictionary (id, name, description) values ($1, $2, $3)",
			i,
			word.Name,
			word.Desc); err != nil {
			return err
		}
	}
	return nil
}
