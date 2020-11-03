package api

import (
	"github.com/shopspring/decimal"
	"time"
)

type OnlineIdentity struct {
	Address      string          `json:"address"`
	LastActivity *time.Time      `json:"lastActivity"`
	Penalty      decimal.Decimal `json:"penalty"`
	Online       bool            `json:"online"`
}

type UpgradeVotes struct {
	Upgrade uint32 `json:"upgrade"`
	Votes   uint64 `json:"votes"`
}
