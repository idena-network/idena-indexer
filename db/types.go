package db

import (
	"math/big"
)

type Data struct {
	Epoch          uint64
	ValidationTime big.Int
	Block          Block
	Identities     []EpochIdentity
	SubmittedFlips []Flip
	FlipKeys       []FlipKey
	FlipStats      []FlipStats
	Addresses      []Address
	FlipsData      []FlipData
}

type Block struct {
	ValidationFinished bool
	Height             uint64
	Hash               string
	Transactions       []Transaction
	Time               big.Int
}

type Transaction struct {
	Hash    string
	Type    string
	From    string
	To      string
	Amount  *big.Int
	Fee     *big.Int
	Payload []byte
}

type EpochIdentity struct {
	Address              string
	State                string
	ShortPoint           float32
	ShortFlips           uint32
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
}

type FlipKey struct {
	TxHash string
	Key    string
}

type FlipData struct {
	Cid    string
	TxHash string
	Data   []byte
}

type Address struct {
	Address  string
	NewState string
}
