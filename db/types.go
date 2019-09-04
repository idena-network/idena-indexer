package db

import (
	"github.com/shopspring/decimal"
	"math/big"
)

type Data struct {
	Epoch               uint64
	ValidationTime      big.Int
	Block               Block
	PrevBlockValidators []string
	Identities          []EpochIdentity
	SubmittedFlips      []Flip
	FlipKeys            []FlipKey
	FlipStats           []FlipStats
	Addresses           []Address
	FlipsData           []FlipData
	FlipSizeUpdates     []FlipSizeUpdate
	BalanceUpdates      []Balance
	BalanceCoins        Coins
	StakeCoins          Coins
	SaveEpochSummary    bool
}

type Block struct {
	ValidationFinished bool
	Height             uint64
	Hash               string
	Transactions       []Transaction
	Time               big.Int
	Proposer           string
	Flags              []string
	IsEmpty            bool
}

type Transaction struct {
	Hash    string
	Type    string
	From    string
	To      string
	Amount  decimal.Decimal
	Fee     decimal.Decimal
	Payload []byte
}

type EpochIdentity struct {
	Address              string
	State                string
	ShortPoint           float32
	ShortFlips           uint32
	TotalShortPoint      float32
	TotalShortFlips      uint32
	LongPoint            float32
	LongFlips            uint32
	Approved             bool
	Missed               bool
	ShortFlipCidsToSolve []string
	LongFlipCidsToSolve  []string
}

type Flip struct {
	TxHash string
	Cid    string
	Size   uint32
}

type FlipStats struct {
	Cid          string
	ShortAnswers []Answer
	LongAnswers  []Answer
	Status       string
	Answer       string
}

type Answer struct {
	Address string
	Answer  string
	Point   float32
}

type FlipKey struct {
	TxHash string
	Key    string
}

type FlipData struct {
	Cid     string
	TxHash  string
	Content FlipContent
}

type FlipSizeUpdate struct {
	Cid  string
	Size uint32
}

type FlipContent struct {
	Pics   [][]byte
	Orders [][]byte
	Icon   []byte
}

type Address struct {
	Address      string
	IsTemporary  bool
	StateChanges []AddressStateChange
}

type AddressStateChange struct {
	PrevState string
	NewState  string
	TxHash    string
}

type Balance struct {
	Address string
	TxHash  string
	Balance decimal.Decimal
	Stake   decimal.Decimal
}

type Coins struct {
	Minted decimal.Decimal
	Burnt  decimal.Decimal
	Total  decimal.Decimal
}

type AddressFlipCid struct {
	Address string
	Cid     string
}
