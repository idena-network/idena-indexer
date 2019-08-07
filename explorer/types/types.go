package types

import (
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/shopspring/decimal"
	"time"
)

type EpochSummary struct {
	Epoch         uint64 `json:"epoch"`
	VerifiedCount uint32 `json:"verified"`
	BlockCount    uint32 `json:"blockCount"`
	FlipCount     uint32 `json:"flipCount"`
}

type EpochDetail struct {
	ValidationTime             time.Time `json:"validationTime"`
	ValidationFirstBlockHeight uint64    `json:"validationFirstBlockHeight"`
}

type BlockSummary struct {
	Height    uint64    `json:"height"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   uint16    `json:"txCount"`
	Proposer  string    `json:"proposer"`
}

type BlockDetail struct {
	Height          uint64    `json:"height"`
	Hash            string    `json:"hash"`
	Timestamp       time.Time `json:"timestamp"`
	TxCount         uint16    `json:"txCount"`
	ValidatorsCount uint16    `json:"validatorsCount"` // todo
	Proposer        string    `json:"proposer"`
}

type IdentityState struct {
	State       string    `json:"state"`
	Epoch       uint64    `json:"epoch"`
	BlockHeight uint64    `json:"blockHeight"`
	BlockHash   string    `json:"blockHash"`
	TxHash      string    `json:"txHash,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

type TransactionSummary struct {
	Hash      string          `json:"hash"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	From      string          `json:"from"`
	To        string          `json:"to,omitempty"`
	Amount    decimal.Decimal `json:"amount"`
	Fee       decimal.Decimal `json:"fee"`
}

type TransactionDetail struct {
	Epoch       uint64          `json:"epoch"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Hash        string          `json:"hash"`
	Type        string          `json:"type"`
	Timestamp   time.Time       `json:"timestamp"`
	From        string          `json:"from"`
	To          string          `json:"to,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
	Fee         decimal.Decimal `json:"fee"`
}

type FlipSummary struct {
	Cid            string    `json:"cid"`
	Author         string    `json:"author"`
	ShortRespCount uint32    `json:"shortRespCount"`
	LongRespCount  uint32    `json:"longRespCount"`
	Status         string    `json:"status"`
	Answer         string    `json:"answer"`
	Timestamp      time.Time `json:"timestamp"`
	Size           uint32    `json:"size"` // todo
}

type Invite struct {
	Hash                string     `json:"hash"`
	Author              string     `json:"author"`
	Timestamp           time.Time  `json:"timestamp"`
	ActivationHash      string     `json:"activationHash"`
	ActivationAuthor    string     `json:"activationAuthor"`
	ActivationTimestamp *time.Time `json:"activationTimestamp"`
}

type EpochIdentitySummary struct {
	Address           string                 `json:"address"`
	Epoch             uint64                 `json:"epoch"`
	State             string                 `json:"state"`
	PrevState         string                 `json:"prevState"`
	ShortAnswers      IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers       IdentityAnswersSummary `json:"longAnswers"`
	Approved          bool                   `json:"approved"`
	Missed            bool                   `json:"missed"`
}

type EpochIdentity struct {
	State             string                 `json:"state"`
	ShortAnswers      IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers       IdentityAnswersSummary `json:"longAnswers"`
	Approved          bool                   `json:"approved"`
	Missed            bool                   `json:"missed"`
}

type Flip struct {
	Status      string        `json:"status"`
	Answer      string        `json:"answer"`
	TxHash      string        `json:"txHash"`
	BlockHash   string        `json:"blockHash"`
	BlockHeight uint64        `json:"blockHeight"`
	Epoch       uint64        `json:"epoch"`
	Data        hexutil.Bytes `json:"hex,omitempty"`
}

type Answer struct {
	Cid        string `json:"cid,omitempty"`
	Address    string `json:"address,omitempty"`
	RespAnswer string `json:"respAnswer"`
	FlipAnswer string `json:"flipAnswer"`
}

type Identity struct {
	Address           string                 `json:"address"`
	State             string                 `json:"state"`
	ShortAnswers      IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers       IdentityAnswersSummary `json:"longAnswers"`
}

type IdentityFlipsSummary struct {
	States  []StrValueCount `json:"states"`
	Answers []StrValueCount `json:"answers"`
}

type IdentityAnswersSummary struct {
	Point      float32 `json:"point"`
	FlipsCount uint32  `json:"flipsCount"`
}

type StrValueCount struct {
	Value string `json:"value"`
	Count uint32 `json:"count"`
}

type InvitesSummary struct {
	AllCount  uint64 `json:"allCount"`
	UsedCount uint64 `json:"usedCount"`
}

type Address struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
	Stake   decimal.Decimal `json:"stake"`
	TxCount uint32          `json:"txCount"`
}
