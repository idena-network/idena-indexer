package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
)

type Accessor interface {
	GetFlipContent(cid string) (db.FlipContent, error)
	GetProposerVrfScore(height uint64) (float64, error)
	GetMemPoolFlipKeys(epoch uint16) ([]*db.MemPoolFlipKey, error)
	GetLastHeight() (uint64, error)
	Destroy()
}

func NewPostgresAccessor(connStr string, scriptsDirPath string) Accessor {
	dbAccessor, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	return &postgresAccessor{
		db:      dbAccessor,
		queries: db.ReadQueries(scriptsDirPath),
	}
}

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
}

const (
	flipContentQuery      = "flipContent.sql"
	flipPicsQuery         = "flipPics.sql"
	flipPicOrdersQuery    = "flipPicOrders.sql"
	flipIconQuery         = "flipIcon.sql"
	proposerVrfScoreQuery = "proposerVrfScore.sql"
	memPoolFlipKeysQuery  = "memPoolFlipKeys.sql"
	maxHeightQuery        = "maxHeight.sql"
)

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) GetLastHeight() (uint64, error) {
	var maxHeight int64
	err := a.db.QueryRow(a.getQuery(maxHeightQuery)).Scan(&maxHeight)
	if err != nil {
		return 0, err
	}
	return uint64(maxHeight), nil
}

func (a *postgresAccessor) GetFlipContent(cid string) (db.FlipContent, error) {
	var dataId int64
	err := a.db.QueryRow(a.getQuery(flipContentQuery), cid).Scan(&dataId)
	if err != nil {
		return db.FlipContent{}, err
	}
	res := db.FlipContent{
		Cid: cid,
	}
	if res.Pics, err = a.flipPics(dataId); err != nil {
		return db.FlipContent{}, err
	}
	if res.Orders, err = a.flipPicOrders(dataId); err != nil {
		return db.FlipContent{}, err
	}
	if res.Icon, err = a.flipIcon(dataId); err != nil {
		return db.FlipContent{}, err
	}
	return res, nil
}

func (a *postgresAccessor) flipPics(dataId int64) ([][]byte, error) {
	rows, err := a.db.Query(a.getQuery(flipPicsQuery), dataId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res [][]byte
	for rows.Next() {
		item := hexutil.Bytes{}
		err := rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) flipPicOrders(dataId int64) ([][]byte, error) {
	rows, err := a.db.Query(a.getQuery(flipPicOrdersQuery), dataId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([][]byte, 2)
	for rows.Next() {
		var answerIndex, flipPicIndex uint16
		err := rows.Scan(&answerIndex, &flipPicIndex)
		if err != nil {
			return nil, err
		}
		switch answerIndex {
		case 0, 1:
			res[answerIndex] = append(res[answerIndex], byte(flipPicIndex))
		default:
			return nil, errors.Errorf("Unknown answer index, flipDataId: %v, index: %v", dataId, answerIndex)
		}
	}
	return res, nil
}

func (a *postgresAccessor) flipIcon(dataId int64) ([]byte, error) {
	res := hexutil.Bytes{}
	err := a.db.QueryRow(a.getQuery(flipIconQuery), dataId).Scan(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *postgresAccessor) GetProposerVrfScore(height uint64) (float64, error) {
	var res float64
	err := a.db.QueryRow(a.getQuery(proposerVrfScoreQuery), height).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (a *postgresAccessor) GetMemPoolFlipKeys(epoch uint16) ([]*db.MemPoolFlipKey, error) {
	rows, err := a.db.Query(a.getQuery(memPoolFlipKeysQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*db.MemPoolFlipKey
	for rows.Next() {
		item := &db.MemPoolFlipKey{}
		err := rows.Scan(&item.Address, &item.Key)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}
