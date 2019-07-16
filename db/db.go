package db

type Accessor interface {
	GetLastHeight() (uint64, error)
	Save(data *Data) error
	Destroy()
}
