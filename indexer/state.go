package indexer

import (
	"github.com/shopspring/decimal"
)

type indexerState struct {
	lastHeight   uint64
	totalBalance decimal.Decimal
	totalStake   decimal.Decimal
}
