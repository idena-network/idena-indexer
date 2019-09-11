package db

type Accessor interface {
	MigrateTo(height uint64) error
	Destroy()
}
