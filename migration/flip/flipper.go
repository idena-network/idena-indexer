package flip

import (
	indexerDb "github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/migration/flip/db"
	"github.com/pkg/errors"
)

type SecondaryFlipStorage struct {
	db              db.Accessor
	lastBlockHeight uint64
}

func NewSecondaryFlipStorage(db db.Accessor) *SecondaryFlipStorage {
	lastBlockHeight, err := db.GetLastHeight()
	if err != nil {
		panic(errors.Wrap(err, "Unable to get last block height from secondary flip storage db"))
	}
	return &SecondaryFlipStorage{
		db:              db,
		lastBlockHeight: lastBlockHeight,
	}
}

func (s *SecondaryFlipStorage) GetFlipContent(cid string) (content indexerDb.FlipContent, err error) {
	return s.db.GetFlipContent(cid)
}

func (s *SecondaryFlipStorage) GetLastBlockHeight() uint64 {
	return s.lastBlockHeight
}

func (s *SecondaryFlipStorage) Destroy() {
	if s.db != nil {
		s.db.Destroy()
		s.db = nil
	}
}
