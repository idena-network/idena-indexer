package types

import (
	"math/big"
	"time"
)

type EpochSummary struct {
	Epoch         uint64 `json:"epoch"`
	VerifiedCount uint32 `json:"verifiedCount"`
	BlockCount    uint32 `json:"blockCount"`
	FlipsCount    uint32 `json:"flipsCount"`
}

type EpochDetail struct {
	Epoch                     uint64 `json:"epoch"`
	VerifiedCount             uint32 `json:"verifiedCount"`
	BlockCount                uint32 `json:"blockCount"`
	FlipsCount                uint32 `json:"flipsCount"`
	FlipsWithKeyCount         uint32 `json:"flipsWithKeyCount"`
	QualifiedFlipsCount       uint32 `json:"qualifiedFlipsCount"`
	WeaklyQualifiedFlipsCount uint32 `json:"weaklyQualifiedFlipsCount"`
}

type Block struct {
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   uint16    `json:"txCount"`
}

type Transaction struct {
	Hash      string    `json:"hash"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	From      string    `json:"from"`
	To        string    `json:"to,omitempty"`
	Amount    *big.Int  `json:"amount"`
	Fee       *big.Int  `json:"fee"`
}

type FlipSummary struct {
	Cid            string `json:"cid"`
	Author         string `json:"author"`
	ShortRespCount uint32 `json:"shortRespCount"`
	LongRespCount  uint32 `json:"longRespCount"`
	Status         string `json:"status"`
}

type Invite struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Status string `json:"status"`
}

type EpochIdentity struct {
	Address     string  `json:"address"`
	State       string  `json:"state"`
	RespScore   float32 `json:"respScore"`
	AuthorScore float32 `json:"authorScore"`
}
