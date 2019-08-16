package db

import "github.com/idena-network/idena-indexer/report/types"

type Accessor interface {
	EpochsCount() (uint64, error)
	FlipCids(epoch uint64) ([]string, error)
	FlipContent(cid string) (types.FlipContent, error)
	Destroy()
}
