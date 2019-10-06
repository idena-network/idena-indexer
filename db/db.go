package db

type Accessor interface {
	GetLastHeight() (uint64, error)
	GetCurrentFlips(address string) ([]Flip, error)
	GetCurrentFlipsWithoutData(limit uint32) ([]AddressFlipCid, error)
	Save(data *Data) error
	SaveRestoredData(data *RestoredData) error
	Destroy()
	ResetTo(height uint64) error
}
