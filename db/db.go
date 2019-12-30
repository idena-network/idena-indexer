package db

type Accessor interface {
	GetLastHeight() (uint64, error)
	GetCurrentFlips(address string) ([]Flip, error)
	GetCurrentFlipsWithoutData(limit uint32) ([]AddressFlipCid, error)
	Save(data *Data) error
	SaveRestoredData(data *RestoredData) error
	SaveMemPoolData(data *MemPoolData) error
	SaveFlipSize(flipCid string, size int) error
	GetEpochFlipsWithoutSize(epoch uint64, limit int) (cids []string, err error)
	Destroy()
	ResetTo(height uint64) error
}
