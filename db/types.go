package db

import (
	"github.com/idena-network/idena-go/common"
	"github.com/shopspring/decimal"
	"math/big"
)

type BurntCoinsReason = uint8

const (
	PenaltyBurntCoins BurntCoinsReason = 0x0
	InviteBurntCoins  BurntCoinsReason = 0x1
	FeeBurntCoins     BurntCoinsReason = 0x2
	KilledBurntCoins  BurntCoinsReason = 0x4
	BurnTxBurntCoins  BurntCoinsReason = 0x5
)

type RestoredData struct {
	Balances  []Balance
	Birthdays []Birthday
}

type Data struct {
	Epoch                  uint64
	ValidationTime         big.Int
	Block                  Block
	ActivationTxTransfers  []ActivationTxTransfer
	KillTxTransfers        []KillTxTransfer
	KillInviteeTxTransfers []KillInviteeTxTransfer
	ActivationTxs          []ActivationTx
	KillInviteeTxs         []KillInviteeTx
	BecomeOnlineTxs        []string
	BecomeOfflineTxs       []string
	Identities             []EpochIdentity
	SubmittedFlips         []Flip
	DeletedFlips           []DeletedFlip
	FlipKeys               []FlipKey
	FlipsWords             []FlipWords
	FlipStats              []FlipStats
	Addresses              []Address
	BalanceUpdates         []Balance
	Birthdays              []Birthday
	MemPoolFlipKeys        []*MemPoolFlipKey
	Coins                  Coins
	SaveEpochSummary       bool
	Penalty                *Penalty
	BurntPenalties         []Penalty
	EpochRewards           *EpochRewards
	MiningRewards          []*MiningReward
	BurntCoinsPerAddr      map[common.Address][]*BurntCoins
	FailedValidation       bool
}

type EpochRewards struct {
	BadAuthors        []*BadAuthor
	GoodAuthors       []*ValidationResult
	Total             *TotalRewards
	ValidationRewards []*Reward
	FundRewards       []*Reward
	AgesByAddress     map[string]uint16
	RewardedFlipCids  []string
}

type TotalRewards struct {
	Total             decimal.Decimal
	Validation        decimal.Decimal
	Flips             decimal.Decimal
	Invitations       decimal.Decimal
	FoundationPayouts decimal.Decimal
	ZeroWalletFund    decimal.Decimal
	ValidationShare   decimal.Decimal
	FlipsShare        decimal.Decimal
	InvitationsShare  decimal.Decimal
}

type Reward struct {
	Address string
	Balance decimal.Decimal
	Stake   decimal.Decimal
	Type    byte
}

type BadAuthor struct {
	Address string
	Reason  byte
}

type MiningReward struct {
	Address  string
	Balance  decimal.Decimal
	Stake    decimal.Decimal
	Proposer bool
}

type ValidationResult struct {
	Address           string
	StrongFlips       int
	WeakFlips         int
	SuccessfulInvites int
}

type Block struct {
	ValidationFinished   bool
	Height               uint64
	Hash                 string
	Transactions         []Transaction
	Time                 big.Int
	Proposer             string
	Flags                []string
	IsEmpty              bool
	BodySize             int
	FullSize             int
	VrfProposerThreshold float64
	ValidatorsCount      int
	ProposerVrfScore     float64
	FeeRate              decimal.Decimal
}

type Transaction struct {
	Hash    string
	Type    uint16
	From    string
	To      string
	Amount  decimal.Decimal
	Tips    decimal.Decimal
	MaxFee  decimal.Decimal
	Fee     decimal.Decimal
	Payload []byte
	Size    int
}

type ActivationTxTransfer struct {
	TxHash          string
	BalanceTransfer decimal.Decimal
}

type KillTxTransfer struct {
	TxHash        string
	StakeTransfer decimal.Decimal
}

type KillInviteeTxTransfer struct {
	TxHash        string
	StakeTransfer decimal.Decimal
}

type ActivationTx struct {
	TxHash       string
	InviteTxHash string
}

type KillInviteeTx struct {
	TxHash       string
	InviteTxHash string
}

type EpochIdentity struct {
	Address              string
	State                uint8
	ShortPoint           float32
	ShortFlips           uint32
	TotalShortPoint      float32
	TotalShortFlips      uint32
	LongPoint            float32
	LongFlips            uint32
	Approved             bool
	Missed               bool
	RequiredFlips        uint8
	AvailableFlips       uint8
	MadeFlips            uint8
	ShortFlipCidsToSolve []string
	LongFlipCidsToSolve  []string
}

type Flip struct {
	TxId   uint64
	TxHash string
	Cid    string
	Pair   uint8
}

type DeletedFlip struct {
	TxHash string
	Cid    string
}

type FlipStats struct {
	Cid          string
	ShortAnswers []Answer
	LongAnswers  []Answer
	Status       byte
	Answer       byte
	WrongWords   bool
}

type Answer struct {
	Address    string
	Answer     byte
	WrongWords bool
	Point      float32
}

type FlipKey struct {
	TxHash string
	Key    string
}

type MemPoolFlipKey struct {
	Address string
	Key     string
}

type FlipWords struct {
	FlipTxId uint64
	TxHash   string
	Word1    uint16
	Word2    uint16
}

type FlipSizeUpdate struct {
	Cid  string
	Size uint32
}

type FlipContent struct {
	Cid    string
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
	PrevState uint8
	NewState  uint8
	TxHash    string
}

type Balance struct {
	Address string
	TxHash  string
	Balance decimal.Decimal
	Stake   decimal.Decimal
}

type Coins struct {
	Minted       decimal.Decimal
	Burnt        decimal.Decimal
	TotalBalance decimal.Decimal
	TotalStake   decimal.Decimal
}

type AddressFlipCid struct {
	FlipId  uint64
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

type BurntCoins struct {
	Amount decimal.Decimal
	Reason BurntCoinsReason
	TxHash string
}

type MemPoolData struct {
	FlipKeyTimestamps       []*MemPoolActionTimestamp
	AnswersHashTxTimestamps []*MemPoolActionTimestamp
}

type MemPoolActionTimestamp struct {
	Address string
	Epoch   uint64
	Time    *big.Int
}

type FlipToLoadContent struct {
	Cid      string
	Key      string
	Attempts int
}

type FailedFlipContent struct {
	Cid                  string
	AttemptsLimitReached bool
	NextAttemptTimestamp *big.Int
}
