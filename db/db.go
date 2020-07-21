package db

import "math/big"

type Accessor interface {
	GetLastHeight() (uint64, error)
	Save(data *Data) error
	SaveRestoredData(data *RestoredData) error
	SaveMemPoolData(data *MemPoolData) error

	SaveFlipSize(flipCid string, size int) error
	GetEpochFlipsWithoutSize(epoch uint64, limit int) (cids []string, err error)
	GetFlipsToLoadContent(timestamp *big.Int, limit int) ([]*FlipToLoadContent, error)
	SaveFlipsContent(failedFlips []*FailedFlipContent, flipsContent []*FlipContent) error

	Destroy()
	ResetTo(height uint64) error
}
