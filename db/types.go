package db

import (
	"math/big"
)

type Data struct {
	Epoch          uint64
	Block          Block
	Identities     []EpochIdentity
	SubmittedFlips []Flip
	FlipKeys       []FlipKey
	FlipStats      []FlipStats
	//Answers        []*Answer
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
	Address    string
	State      string
	ShortPoint float32
	ShortFlips uint32
	LongPoint  float32
	LongFlips  uint32
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
