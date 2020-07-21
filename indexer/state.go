package indexer

import (
	"math/big"
)

type indexerState struct {
	lastIndexedHeight uint64
	totalBalance      *big.Int
	totalStake        *big.Int
}
