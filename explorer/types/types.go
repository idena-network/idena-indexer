package types

import (
	"math/big"
	"time"
)

type EpochSummary struct {
	Epoch         uint64 `json:"epoch"`
	VerifiedCount uint32 `json:"verifiedCount"`
	BlockCount    uint32 `json:"blockCount"`
	FlipCount     uint32 `json:"flipCount"`
}

type EpochDetail struct {
	Epoch                    uint64 `json:"epoch"`
	VerifiedCount            uint32 `json:"verifiedCount"`
	BlockCount               uint32 `json:"blockCount"`
	FlipCount                uint32 `json:"flipCount"`
	FlipsWithKeyCount        uint32 `json:"flipsWithKeyCount"`
	QualifiedFlipCount       uint32 `json:"qualifiedFlipCount"`
	WeaklyQualifiedFlipCount uint32 `json:"weaklyQualifiedFlipCount"`
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

type Flip struct {
	// todo images
	Status       string   `json:"status"`
	Answer       string   `json:"answer"`
	ShortAnswers []Answer `json:"shortAnswers"`
	LongAnswers  []Answer `json:"longAnswers"`
}

type Answer struct {
	Address string `json:"address"`
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

	Epochs          []IdentityEpoch `json:"epochs"`
	Txs             []Transaction   `json:"txs"`
	CurrentFlipCids []string        `json:"currentFlipCids"`
	Invites         []Invite        `json:"invites"`
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
	State       string  `json:"string"`
}
