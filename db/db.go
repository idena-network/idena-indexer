package db

import "github.com/shopspring/decimal"

type Accessor interface {
	GetLastHeight() (uint64, error)
	GetTotalCoins() (balance decimal.Decimal, stake decimal.Decimal, err error)
	GetCurrentFlipCids(address string) ([]string, error)
	GetCurrentFlipsWithoutData(limit uint32) ([]string, error)
	Save(data *Data) error
	Destroy()
	ResetTo(height uint64) error
}
