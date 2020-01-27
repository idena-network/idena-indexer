package runtime

import (
	indexerDb "github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/migration/runtime/db"
	"github.com/pkg/errors"
)

type SecondaryStorage struct {
	db              db.Accessor
	lastBlockHeight uint64
}

func NewSecondaryStorage(db db.Accessor) *SecondaryStorage {
	lastBlockHeight, err := db.GetLastHeight()
	if err != nil {
		panic(errors.Wrap(err, "Unable to get last block height from secondary storage db"))
	}
	return &SecondaryStorage{
		db:              db,
		lastBlockHeight: lastBlockHeight,
	}
}

func (s *SecondaryStorage) GetFlipContent(cid string) (indexerDb.FlipContent, error) {
	return s.db.GetFlipContent(cid)
}

func (s *SecondaryStorage) GetProposerVrfScore(blockHeight uint64) (float64, error) {
	return s.db.GetProposerVrfScore(blockHeight)
}

func (s *SecondaryStorage) GetLastBlockHeight() uint64 {
	return s.lastBlockHeight
}

func (s *SecondaryStorage) GetMemPoolFlipKeys(epoch uint16) ([]*indexerDb.MemPoolFlipKey, error) {
	return s.db.GetMemPoolFlipKeys(epoch)
}

func (s *SecondaryStorage) Destroy() {
	if s.db != nil {
		s.db.Destroy()
		s.db = nil
	}
}
