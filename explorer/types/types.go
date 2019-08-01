package types

import (
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/shopspring/decimal"
	"time"
)

type EpochSummary struct {
	Epoch         uint64 `json:"epoch"`
	VerifiedCount uint32 `json:"verified"`
	BlockCount    uint32 `json:"blocks"`
	FlipCount     uint32 `json:"flips"`
}

type EpochDetail struct {
	Epoch            uint64                     `json:"epoch"`
	LatestValidation CompletedValidationSummary `json:"latestValidation"`
	NextValidation   NewValidationSummary       `json:"nextValidation"`
	BlockCount       uint32                     `json:"blockCount"`
}

type BlockSummary struct {
	Height    uint64    `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   uint16    `json:"txCount"`
}

type BlockDetail struct {
	Height          uint64    `json:"height"`
	Timestamp       time.Time `json:"timestamp"`
	TxCount         uint16    `json:"txCount"`
	ValidatorsCount uint16    `json:"validatorsCount"`
	Proposer        string    `json:"proposer"`
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

type EpochIdentitySummary struct {
	Address     string  `json:"address"`
	State       string  `json:"state"`
	PrevState   string  `json:"prevState"`
	RespScore   float32 `json:"respScore"`
	AuthorScore float32 `json:"authorScore"`
	Approved    bool    `json:"approved"`
	Missed      bool    `json:"missed"`
}

type EpochIdentity struct {
	ShortFlipsToSolve []string `json:"shortFlipToSolve"`
	LongFlipsToSolve  []string `json:"longFlipToSolve"`
	ShortAnswers      []Answer `json:"shortAnswers"`
	LongAnswers       []Answer `json:"longAnswers"`
}

type Flip struct {
	Status       string        `json:"status"`
	Answer       string        `json:"answer"`
	ShortAnswers []Answer      `json:"shortAnswers"`
	LongAnswers  []Answer      `json:"longAnswers"`
	Data         hexutil.Bytes `json:"hex,omitempty"`
}

type Answer struct {
	Cid     string `json:"cid,omitempty"`
	Address string `json:"address,omitempty"`
	Answer  string `json:"answer"`
}

type Identity struct {
	Address  string `json:"address"`
	Nickname string `json:"nickname,omitempty"` // todo
	Age      uint16 `json:"age,omitempty"`      // todo
	State    string `json:"state"`

	ShortAnswers IdentityAnswerSummary `json:"shortAnswers"`
	LongAnswers  IdentityAnswerSummary `json:"longAnswers"`

	SubmittedFlipCount       uint32  `json:"submittedFlipCount"`
	QualifiedFlipCount       uint32  `json:"qualifiedFlipCount"`
	WeaklyQualifiedFlipCount uint32  `json:"weaklyQualifiedFlipCount"`
	AuthorScore              float32 `json:"authorScore"`

	Epochs          []IdentityEpoch      `json:"epochs"`
	Txs             []TransactionSummary `json:"txs"`
	CurrentFlipCids []string             `json:"currentFlipCids"`
	Invites         []Invite             `json:"invites"`
}

type IdentityAnswerSummary struct {
	AnswerCount      uint32  `json:"answerCount"`
	RightAnswerCount uint32  `json:"rightAnswerCount"`
	RespScore        float32 `json:"respScore"`
}

type IdentityEpoch struct {
	Epoch       uint64  `json:"epoch"`
	RespScore   float32 `json:"respScore"`
	AuthorScore float32 `json:"authorScore"`
	State       string  `json:"state"`
	Approved    bool    `json:"approved"`
	Missed      bool    `json:"missed"`
}

type Summary struct {
	Identities       IdentitiesSummary          `json:"identities"`
	LatestValidation CompletedValidationSummary `json:"latestValidation,omitempty"`
	NextValidation   NewValidationSummary       `json:"nextValidation"`
}

type IdentitiesSummary struct {
	States []StateCount `json:"States"`
}

type StateCount struct {
	State string `json:"state"`
	Count uint32 `json:"count"`
}

type CompletedValidationSummary struct {
	Time                 time.Time         `json:"time"`
	FirstBlockHeight     uint64            `json:"firstBlockHeight"`
	Identities           IdentitiesSummary `json:"identities"`
	SubmittedFlips       uint32            `json:"submittedFlips"`
	SolvedFlips          uint32            `json:"solvedFlips"`
	FlipsWithKey         uint32            `json:"flipsWithKey"`
	QualifiedFlips       uint32            `json:"qualifiedFlips"`
	WeaklyQualifiedFlips uint32            `json:"weaklyQualifiedFlips"`
	NotQualifiedFlips    uint32            `json:"notQualifiedFlips"`
	InappropriateFlips   uint32            `json:"inappropriateFlips"`
}

type NewValidationSummary struct {
	Time    time.Time `json:"time"`
	Invites uint32    `json:"invites"`
	Flips   uint32    `json:"flips"`
}

type Address struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
	Stake   decimal.Decimal `json:"stake"`
	TxCount uint32          `json:"txCount"`
}
