package db

import (
	"github.com/shopspring/decimal"
	"math/big"
)

type RestoredData struct {
	Balances  []Balance
	Birthdays []Birthday
}

type Data struct {
	Epoch            uint64
	ValidationTime   big.Int
	Block            Block
	Identities       []EpochIdentity
	SubmittedFlips   []Flip
	FlipKeys         []FlipKey
	FlipsWords       []FlipWords
	FlipStats        []FlipStats
	Addresses        []Address
	FlipsData        []FlipData
	FlipSizeUpdates  []FlipSizeUpdate
	BalanceUpdates   []Balance
	Birthdays        []Birthday
	BalanceCoins     Coins
	StakeCoins       Coins
	SaveEpochSummary bool
	Penalty          *Penalty
	BurntPenalties   []Penalty
	EpochRewards     *EpochRewards
	MiningRewards    []*Reward
	FailedValidation bool
}

type EpochRewards struct {
	BadAuthors        []string
	GoodAuthors       []*ValidationResult
	Total             *TotalRewards
	ValidationRewards []*Reward
	FundRewards       []*Reward
	AgesByAddress     map[string]uint16
}

type TotalRewards struct {
	Total             decimal.Decimal
	Validation        decimal.Decimal
	Flips             decimal.Decimal
	Invitations       decimal.Decimal
	FoundationPayouts decimal.Decimal
	ZeroWalletFund    decimal.Decimal
}

type Reward struct {
	Address string
	Balance decimal.Decimal
	Stake   decimal.Decimal
	Type    string
}

type ValidationResult struct {
	Address           string
	StrongFlips       int
	WeakFlips         int
	SuccessfulInvites int
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
	Size               int
	ValidatorsCount    int
}

type Transaction struct {
	Hash    string
	Type    string
	From    string
	To      string
	Amount  decimal.Decimal
	Fee     decimal.Decimal
	Payload []byte
	Size    int
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
	RequiredFlips        uint8
	MadeFlips            uint8
	ShortFlipCidsToSolve []string
	LongFlipCidsToSolve  []string
}

type Flip struct {
	Id     uint64
	TxHash string
	Cid    string
	Size   uint32
	Pair   uint8
}

type FlipStats struct {
	Cid          string
	ShortAnswers []Answer
	LongAnswers  []Answer
	Status       string
	Answer       string
	WrongWords   bool
}

type Answer struct {
	Address    string
	Answer     string
	WrongWords bool
	Point      float32
}

type FlipKey struct {
	TxHash string
	Key    string
}

type FlipWords struct {
	FlipId uint64
	TxHash string
	Word1  uint16
	Word2  uint16
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

type Penalty struct {
	Address string
	Penalty decimal.Decimal
}

type Birthday struct {
	Address    string
	BirthEpoch uint64
}
