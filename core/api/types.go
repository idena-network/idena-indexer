package api

import (
	"github.com/shopspring/decimal"
	"time"
)

type Activity struct {
	Address string    `json:"address"`
	Time    time.Time `json:"timestamp"`
}

type Penalty struct {
	Address string          `json:"address"`
	Penalty decimal.Decimal `json:"penalty"`
}
