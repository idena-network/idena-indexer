package indexer

import (
	"math/big"
)

type indexerState struct {
	lastHeight   uint64
	totalBalance *big.Int
	totalStake   *big.Int
}
