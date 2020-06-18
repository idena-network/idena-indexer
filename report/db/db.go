package db

type Accessor interface {
	EpochsCount() (uint64, error)
	FlipCids(epoch uint64) ([]string, error)
	FlipContent(cid string) (FlipContent, error)
	Destroy()
}
