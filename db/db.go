package db

type Accessor interface {
	GetLastHeight() (uint64, error)
	GetCurrentFlipCids(address string) ([]string, error)
	GetCurrentFlipsWithoutData(limit uint32) ([]string, error)
	Save(data *Data) error
	Destroy()
	ResetTo(height uint64) error
}
