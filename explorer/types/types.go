package types

import (
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/shopspring/decimal"
	"time"
)

type Entity struct {
	Name string
	Ref  string
}

type EpochSummary struct {
	Epoch           uint64   `json:"epoch"`
	ValidatedCount  uint32   `json:"validatedCount"`
	BlockCount      uint32   `json:"blockCount"`
	EmptyBlockCount uint32   `json:"emptyBlockCount"`
	TxCount         uint32   `json:"txCount"`
	InviteCount     uint32   `json:"inviteCount"`
	FlipCount       uint32   `json:"flipCount"`
	Coins           AllCoins `json:"coins"`
}

type EpochDetail struct {
	Epoch                      uint64    `json:"epoch"`
	ValidationTime             time.Time `json:"validationTime"`
	ValidationFirstBlockHeight uint64    `json:"validationFirstBlockHeight"`
}

type BlockSummary struct {
	Height    uint64    `json:"height"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   uint16    `json:"txCount"`
	Proposer  string    `json:"proposer"`
	IsEmpty   bool      `json:"isEmpty"`
	Coins     AllCoins  `json:"coins"`
	Size      uint32    `json:"size"`
}

type BlockDetail struct {
	Height          uint64    `json:"height"`
	Hash            string    `json:"hash"`
	Timestamp       time.Time `json:"timestamp"`
	TxCount         uint16    `json:"txCount"`
	ValidatorsCount uint16    `json:"validatorsCount"`
	Proposer        string    `json:"proposer"`
	IsEmpty         bool      `json:"isEmpty"`
	Size            uint32    `json:"size"`
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
	Tips      decimal.Decimal `json:"tips"`
	MaxFee    decimal.Decimal `json:"maxFee"`
	Fee       decimal.Decimal `json:"fee"`
	Size      uint32          `json:"size"`
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
	Tips        decimal.Decimal `json:"tips"`
	MaxFee      decimal.Decimal `json:"maxFee"`
	Fee         decimal.Decimal `json:"fee"`
	Size        uint32          `json:"size"`
}

type FlipSummary struct {
	Cid            string        `json:"cid"`
	Author         string        `json:"author"`
	Epoch          uint64        `json:"epoch"`
	ShortRespCount uint32        `json:"shortRespCount"`
	LongRespCount  uint32        `json:"longRespCount"`
	Status         string        `json:"status"`
	Answer         string        `json:"answer"`
	WrongWords     bool          `json:"wrongWords"`
	Timestamp      time.Time     `json:"timestamp"`
	Size           uint32        `json:"size"`
	Icon           hexutil.Bytes `json:"icon,omitempty"`
	Words          *FlipWords    `json:"words"`
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
	PrevState         string                 `json:"prevState"`
	State             string                 `json:"state"`
	ShortAnswers      IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers       IdentityAnswersSummary `json:"longAnswers"`
	Approved          bool                   `json:"approved"`
	Missed            bool                   `json:"missed"`
	RequiredFlips     uint8                  `json:"requiredFlips"`
	MadeFlips         uint8                  `json:"madeFlips"`
}

type EpochIdentity struct {
	PrevState         string                 `json:"prevState"`
	State             string                 `json:"state"`
	ShortAnswers      IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers       IdentityAnswersSummary `json:"longAnswers"`
	Approved          bool                   `json:"approved"`
	Missed            bool                   `json:"missed"`
	RequiredFlips     uint8                  `json:"requiredFlips"`
	MadeFlips         uint8                  `json:"madeFlips"`
}

type Flip struct {
	Author      string     `json:"author"`
	Timestamp   time.Time  `json:"timestamp"`
	Size        uint32     `json:"size"`
	Status      string     `json:"status"`
	Answer      string     `json:"answer"`
	WrongWords  bool       `json:"wrongWords"`
	TxHash      string     `json:"txHash"`
	BlockHash   string     `json:"blockHash"`
	BlockHeight uint64     `json:"blockHeight"`
	Epoch       uint64     `json:"epoch"`
	Words       *FlipWords `json:"words"`
}

type FlipWords struct {
	Word1 FlipWord `json:"word1"`
	Word2 FlipWord `json:"word2"`
}

func (w FlipWords) IsEmpty() bool {
	return w.Word1.isEmpty() && w.Word2.isEmpty()
}

type FlipWord struct {
	Index uint16 `json:"index"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
}

func (w FlipWord) isEmpty() bool {
	return w.Index == 0 && len(w.Name) == 0 && len(w.Desc) == 0
}

type FlipContent struct {
	LeftOrder  []uint16
	RightOrder []uint16
	Pics       []hexutil.Bytes
}

type Answer struct {
	Cid            string  `json:"cid,omitempty"`
	Address        string  `json:"address,omitempty"`
	RespAnswer     string  `json:"respAnswer"`
	RespWrongWords bool    `json:"respWrongWords"`
	FlipAnswer     string  `json:"flipAnswer"`
	FlipWrongWords bool    `json:"flipWrongWords"`
	FlipStatus     string  `json:"flipStatus"`
	Point          float32 `json:"point"`
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

type AllCoins struct {
	Minted       decimal.Decimal `json:"minted"`
	Burnt        decimal.Decimal `json:"burnt"`
	TotalBalance decimal.Decimal `json:"totalBalance"`
	TotalStake   decimal.Decimal `json:"totalStake"`

	// todo tmp for backward compatibility
	Balance Coins `json:"balance"`
	Stake   Coins `json:"stake"`
}

type Coins struct {
	Minted decimal.Decimal `json:"minted"`
	Burnt  decimal.Decimal `json:"burnt"`
	Total  decimal.Decimal `json:"total"`
}

type Balance struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
	Stake   decimal.Decimal `json:"stake"`
}

type Penalty struct {
	Address     string          `json:"address"`
	Penalty     decimal.Decimal `json:"penalty"`
	Paid        decimal.Decimal `json:"paid"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Timestamp   time.Time       `json:"timestamp"`
	Epoch       uint64          `json:"epoch"`
}

type RewardsSummary struct {
	Epoch             uint64          `json:"epoch"`
	Total             decimal.Decimal `json:"total"`
	Validation        decimal.Decimal `json:"validation"`
	Flips             decimal.Decimal `json:"flips"`
	Invitations       decimal.Decimal `json:"invitations"`
	FoundationPayouts decimal.Decimal `json:"foundationPayouts"`
	ZeroWalletFund    decimal.Decimal `json:"zeroWalletFund"`
}

type AuthorValidationSummary struct {
	Address           string `json:"address"`
	StrongFlips       uint16 `json:"strongFlips"`
	WeakFlips         uint16 `json:"weakFlips"`
	SuccessfulInvites uint16 `json:"successfulInvites"`
}

type Reward struct {
	Address     string          `json:"address,omitempty"`
	Epoch       uint64          `json:"epoch,omitempty"`
	BlockHeight uint64          `json:"blockHeight,omitempty"`
	Balance     decimal.Decimal `json:"balance"`
	Stake       decimal.Decimal `json:"stake"`
	Type        string          `json:"type"`
}

type Rewards struct {
	Address   string   `json:"address,omitempty"`
	Epoch     uint64   `json:"epoch,omitempty"`
	PrevState string   `json:"prevState"`
	State     string   `json:"state"`
	Age       uint16   `json:"age"`
	Rewards   []Reward `json:"rewards"`
}

type BlockRewards struct {
	Height  uint64   `json:"height"`
	Epoch   uint64   `json:"epoch"`
	Rewards []Reward `json:"rewards"`
}

type FundPayment struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
	Type    string          `json:"type"`
}
