package types

import (
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/shopspring/decimal"
	"time"
)

type Entity struct {
	NameOld  string `json:"Name" swaggerignore:"true"`  // todo deprecated
	ValueOld string `json:"Value" swaggerignore:"true"` // todo deprecated
	RefOld   string `json:"Ref" swaggerignore:"true"`   // todo deprecated
	Name     string `json:"name" enums:"Address,Identity,Epoch,Block,Transaction,Flip"`
	Value    string `json:"value"`
	Ref      string `json:"ref"`
} // @Name Entity

type EpochSummary struct {
	Epoch             uint64         `json:"epoch"`
	ValidationTime    time.Time      `json:"validationTime" example:"2020-01-01T00:00:00Z"`
	ValidatedCount    uint32         `json:"validatedCount"`
	BlockCount        uint32         `json:"blockCount"`
	EmptyBlockCount   uint32         `json:"emptyBlockCount"`
	TxCount           uint32         `json:"txCount"`
	InviteCount       uint32         `json:"inviteCount"`
	FlipCount         uint32         `json:"flipCount"`
	Coins             AllCoins       `json:"coins"`
	Rewards           RewardsSummary `json:"rewards"`
	MinScoreForInvite float32        `json:"minScoreForInvite"`
} // @Name EpochSummary

type AllCoins struct {
	Minted       decimal.Decimal `json:"minted" swaggertype:"string"`
	Burnt        decimal.Decimal `json:"burnt" swaggertype:"string"`
	TotalBalance decimal.Decimal `json:"totalBalance" swaggertype:"string"`
	TotalStake   decimal.Decimal `json:"totalStake" swaggertype:"string"`
} // @Name Coins

type RewardsSummary struct {
	Epoch             uint64          `json:"epoch,omitempty"`
	Total             decimal.Decimal `json:"total" swaggertype:"string"`
	Validation        decimal.Decimal `json:"validation" swaggertype:"string"`
	Flips             decimal.Decimal `json:"flips" swaggertype:"string"`
	Invitations       decimal.Decimal `json:"invitations" swaggertype:"string"`
	FoundationPayouts decimal.Decimal `json:"foundationPayouts" swaggertype:"string"`
	ZeroWalletFund    decimal.Decimal `json:"zeroWalletFund" swaggertype:"string"`
	ValidationShare   decimal.Decimal `json:"validationShare" swaggertype:"string"`
	FlipsShare        decimal.Decimal `json:"flipsShare" swaggertype:"string"`
	InvitationsShare  decimal.Decimal `json:"invitationsShare" swaggertype:"string"`
} // @Name RewardsSummary

type EpochDetail struct {
	Epoch                      uint64    `json:"epoch"`
	ValidationTime             time.Time `json:"validationTime" example:"2020-01-01T00:00:00Z"`
	ValidationFirstBlockHeight uint64    `json:"validationFirstBlockHeight"`
	MinScoreForInvite          float32   `json:"minScoreForInvite"`
} // @Name Epoch

type BlockSummary struct {
	Height               uint64          `json:"height"`
	Hash                 string          `json:"hash"`
	Timestamp            time.Time       `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	TxCount              uint16          `json:"txCount"`
	IsEmpty              bool            `json:"isEmpty"`
	Coins                AllCoins        `json:"coins"`
	BodySize             uint32          `json:"bodySize"`
	FullSize             uint32          `json:"fullSize"`
	VrfProposerThreshold float64         `json:"vrfProposerThreshold"`
	Proposer             string          `json:"proposer"`
	ProposerVrfScore     float64         `json:"proposerVrfScore,omitempty"`
	FeeRate              decimal.Decimal `json:"feeRate" swaggertype:"string"`
	Flags                []string        `json:"flags" enums:"IdentityUpdate,FlipLotteryStarted,ShortSessionStarted,LongSessionStarted,AfterLongSessionStarted,ValidationFinished,Snapshot,OfflinePropose,OfflineCommit"`
} // @Name BlockSummary

type BlockDetail struct {
	Epoch                uint64          `json:"epoch"`
	Height               uint64          `json:"height"`
	Hash                 string          `json:"hash"`
	Timestamp            time.Time       `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	TxCount              uint16          `json:"txCount"`
	ValidatorsCount      uint16          `json:"validatorsCount"`
	IsEmpty              bool            `json:"isEmpty"`
	BodySize             uint32          `json:"bodySize"`
	FullSize             uint32          `json:"fullSize"`
	VrfProposerThreshold float64         `json:"vrfProposerThreshold"`
	Proposer             string          `json:"proposer"`
	ProposerVrfScore     float64         `json:"proposerVrfScore,omitempty"`
	FeeRate              decimal.Decimal `json:"feeRate" swaggertype:"string"`
	Flags                []string        `json:"flags" enums:"IdentityUpdate,FlipLotteryStarted,ShortSessionStarted,LongSessionStarted,AfterLongSessionStarted,ValidationFinished,Snapshot,OfflinePropose,OfflineCommit"`
} // @Name Block

type FlipSummary struct {
	Cid            string `json:"cid"`
	Author         string `json:"author"`
	Epoch          uint64 `json:"epoch"`
	ShortRespCount uint32 `json:"shortRespCount"`
	LongRespCount  uint32 `json:"longRespCount"`
	Status         string `json:"status" enums:",NotQualified,Qualified,WeaklyQualified,QualifiedByNone"`
	Answer         string `json:"answer" enums:",None,Left,Right"`
	// Deprecated
	WrongWords      bool          `json:"wrongWords"`
	WrongWordsVotes uint32        `json:"wrongWordsVotes"`
	Timestamp       time.Time     `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Size            uint32        `json:"size"`
	Icon            hexutil.Bytes `json:"icon,omitempty"`
	Words           *FlipWords    `json:"words"`
	WithPrivatePart bool          `json:"withPrivatePart"`
	Grade           byte          `json:"grade"`
} // @Name FlipSummary

type FlipWords struct {
	Word1 FlipWord `json:"word1"`
	Word2 FlipWord `json:"word2"`
} // @Name FlipWords

func (w FlipWords) IsEmpty() bool {
	return w.Word1.isEmpty() && w.Word2.isEmpty()
}

type FlipWord struct {
	Index uint16 `json:"index"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
} // @Name FlipWord

func (w FlipWord) isEmpty() bool {
	return w.Index == 0 && len(w.Name) == 0 && len(w.Desc) == 0
}

// mock type for swagger
type FlipAnswerCount struct {
	Value string `json:"value" enums:",None,Left,Right"`
	Count uint32 `json:"count"`
} // @Name FlipAnswerCount

// mock type for swagger
type FlipStateCount struct {
	Value string `json:"value" enums:",NotQualified,Qualified,WeaklyQualified,QualifiedByNone"`
	Count uint32 `json:"count"`
} // @Name FlipStateCount

// mock type for swagger
type IdentityStateCount struct {
	Value string `json:"value" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	Count uint32 `json:"count"`
} // @Name IdentityStateCount

// mock type for swagger
type SavedInviteRewardCount struct {
	Value string `json:"value" enums:"SavedInvite,SavedInviteWin"`
	Count uint32 `json:"count"`
} // @Name SavedInviteRewardCount

type NullableBoolValueCount struct {
	Value *bool  `json:"value"`
	Count uint32 `json:"count"`
} // @Name NullableBoolValueCount

type EpochIdentity struct {
	Address               string                 `json:"address,omitempty"`
	Epoch                 uint64                 `json:"epoch,omitempty"`
	PrevState             string                 `json:"prevState" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	State                 string                 `json:"state" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	ShortAnswers          IdentityAnswersSummary `json:"shortAnswers"`
	TotalShortAnswers     IdentityAnswersSummary `json:"totalShortAnswers"`
	LongAnswers           IdentityAnswersSummary `json:"longAnswers"`
	ShortAnswersCount     uint32                 `json:"shortAnswersCount"`
	LongAnswersCount      uint32                 `json:"longAnswersCount"`
	Approved              bool                   `json:"approved"`
	Missed                bool                   `json:"missed"`
	RequiredFlips         uint8                  `json:"requiredFlips"`
	MadeFlips             uint8                  `json:"madeFlips"`
	AvailableFlips        uint8                  `json:"availableFlips"`
	TotalValidationReward decimal.Decimal        `json:"totalValidationReward" swaggertype:"string"`
	BirthEpoch            uint64                 `json:"birthEpoch"`
} // @Name EpochIdentity

type TransactionSummary struct {
	Hash      string          `json:"hash"`
	Type      string          `json:"type" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx"`
	Timestamp time.Time       `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	From      string          `json:"from"`
	To        string          `json:"to,omitempty"`
	Amount    decimal.Decimal `json:"amount" swaggertype:"string"`
	Tips      decimal.Decimal `json:"tips" swaggertype:"string"`
	MaxFee    decimal.Decimal `json:"maxFee" swaggertype:"string"`
	Fee       decimal.Decimal `json:"fee" swaggertype:"string"`
	Size      uint32          `json:"size"`
	// Deprecated
	Transfer *decimal.Decimal `json:"transfer,omitempty" swaggerignore:"true"`
	Data     interface{}      `json:"data,omitempty"`
} // @Name TransactionSummary

// mock type for swagger
type TransactionSpecificData struct {
	Transfer     *decimal.Decimal `json:"transfer,omitempty" swaggertype:"string"`
	BecomeOnline bool             `json:"becomeOnline"`
} // @Name TransactionSpecificData

type TransactionDetail struct {
	Epoch       uint64          `json:"epoch"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Hash        string          `json:"hash"`
	Type        string          `json:"type" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx"`
	Timestamp   time.Time       `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	From        string          `json:"from"`
	To          string          `json:"to,omitempty"`
	Amount      decimal.Decimal `json:"amount" swaggertype:"string"`
	Tips        decimal.Decimal `json:"tips" swaggertype:"string"`
	MaxFee      decimal.Decimal `json:"maxFee" swaggertype:"string"`
	Fee         decimal.Decimal `json:"fee" swaggertype:"string"`
	Size        uint32          `json:"size"`
	// Deprecated
	Transfer *decimal.Decimal `json:"transfer,omitempty" swaggertype:"string"`
	Data     interface{}      `json:"data,omitempty"`
} // @Name Transaction

type ActivationTxSpecificData struct {
	Transfer *string `json:"transfer,omitempty"`
}

type KillTxSpecificData = ActivationTxSpecificData

type KillInviteeTxSpecificData = ActivationTxSpecificData

type OnlineStatusTxSpecificData struct {
	// Deprecated
	BecomeOnlineOld bool `json:"BecomeOnline"`
	BecomeOnline    bool `json:"becomeOnline"`
}

type Invite struct {
	Hash                 string     `json:"hash"`
	Author               string     `json:"author"`
	Timestamp            time.Time  `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Epoch                uint64     `json:"epoch"`
	ActivationHash       string     `json:"activationHash"`
	ActivationAuthor     string     `json:"activationAuthor"`
	ActivationTimestamp  *time.Time `json:"activationTimestamp" example:"2020-01-01T00:00:00Z"`
	State                string     `json:"state" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	KillInviteeHash      string     `json:"killInviteeHash,omitempty"`
	KillInviteeTimestamp *time.Time `json:"killInviteeTimestamp,omitempty" example:"2020-01-01T00:00:00Z"`
	KillInviteeEpoch     uint64     `json:"killInviteeEpoch,omitempty"`
} // @Name Invite

type Flip struct {
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Size      uint32    `json:"size"`
	Status    string    `json:"status" enums:",NotQualified,Qualified,WeaklyQualified,QualifiedByNone"`
	Answer    string    `json:"answer" enums:",None,Left,Right"`
	// Deprecated
	WrongWords      bool       `json:"wrongWords"`
	WrongWordsVotes uint32     `json:"wrongWordsVotes"`
	TxHash          string     `json:"txHash"`
	BlockHash       string     `json:"blockHash"`
	BlockHeight     uint64     `json:"blockHeight"`
	Epoch           uint64     `json:"epoch"`
	Words           *FlipWords `json:"words"`
	WithPrivatePart bool       `json:"withPrivatePart"`
	Grade           byte       `json:"grade"`
} // @Name Flip

type FlipContent struct {
	LeftOrder  []uint16        `json:"leftOrder"`
	RightOrder []uint16        `json:"rightOrder"`
	Pics       []hexutil.Bytes `json:"pics" swaggertype:"array"`
	// Deprecated
	LeftOrderOld []uint16 `json:"LeftOrder" swaggerignore:"true"`
	// Deprecated
	RightOrderOld []uint16 `json:"RightOrder" swaggerignore:"true"`
	// Deprecated
	PicsOld []hexutil.Bytes `json:"Pics" swaggerignore:"true"`
} // @Name FlipContent

type Answer struct {
	Cid        string `json:"cid,omitempty"`
	Address    string `json:"address,omitempty"`
	RespAnswer string `json:"respAnswer" enums:"None,Left,Right"`
	// Deprecated
	RespWrongWords bool   `json:"respWrongWords"`
	FlipAnswer     string `json:"flipAnswer" enums:"None,Left,Right"`
	// Deprecated
	FlipWrongWords bool    `json:"flipWrongWords"`
	FlipStatus     string  `json:"flipStatus" enums:"NotQualified,Qualified,WeaklyQualified,QualifiedByNone"`
	Point          float32 `json:"point"`
	RespGrade      byte    `json:"respGrade"`
	FlipGrade      byte    `json:"flipGrade"`
} // @Name Answer

type Identity struct {
	Address           string                 `json:"address"`
	State             string                 `json:"state"`
	TotalShortAnswers IdentityAnswersSummary `json:"totalShortAnswers"`
} // @Name Identity

type IdentityFlipsSummary struct {
	States  []StrValueCount `json:"states"`
	Answers []StrValueCount `json:"answers"`
}

type IdentityAnswersSummary struct {
	Point      float32 `json:"point"`
	FlipsCount uint32  `json:"flipsCount"`
} // @Name IdentityAnswersSummary

type InvitesSummary struct {
	AllCount  uint64 `json:"allCount"`
	UsedCount uint64 `json:"usedCount"`
} // @Name InvitesSummary

type Address struct {
	Address            string          `json:"address"`
	Balance            decimal.Decimal `json:"balance" swaggertype:"string"`
	Stake              decimal.Decimal `json:"stake" swaggertype:"string"`
	TxCount            uint32          `json:"txCount"`
	FlipsCount         uint32          `json:"flipsCount"`
	ReportedFlipsCount uint32          `json:"reportedFlipsCount"`
} // @Name Address

type Balance struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance" swaggertype:"string"`
	Stake   decimal.Decimal `json:"stake" swaggertype:"string"`
} // @Name Balance

type TotalMiningReward struct {
	Address        string          `json:"address,omitempty"`
	Balance        decimal.Decimal `json:"balance"`
	Stake          decimal.Decimal `json:"stake"`
	Proposer       uint64          `json:"proposer"`
	FinalCommittee uint64          `json:"finalCommittee"`
}

type Penalty struct {
	Address     string          `json:"address"`
	Penalty     decimal.Decimal `json:"penalty" swaggertype:"string"`
	Paid        decimal.Decimal `json:"paid" swaggertype:"string"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Timestamp   time.Time       `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	Epoch       uint64          `json:"epoch"`
} // @Name Penalty

type Reward struct {
	Address     string          `json:"address,omitempty"`
	Epoch       uint64          `json:"epoch,omitempty"`
	BlockHeight uint64          `json:"blockHeight,omitempty"`
	Balance     decimal.Decimal `json:"balance" swaggertype:"string"`
	Stake       decimal.Decimal `json:"stake" swaggertype:"string"`
	Type        string          `json:"type" enums:"Validation,Flips,Invitations,Invitations2,Invitations3,SavedInvite,SavedInviteWin"`
} // @Name Reward

type Rewards struct {
	Address   string   `json:"address,omitempty"`
	Epoch     uint64   `json:"epoch,omitempty"`
	PrevState string   `json:"prevState" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	State     string   `json:"state" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	Age       uint16   `json:"age"`
	Rewards   []Reward `json:"rewards"`
} // @Name Rewards

type FundPayment struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance" swaggertype:"string"`
	Type    string          `json:"type" enums:"FoundationPayouts,ZeroWalletFund"`
} // @Name FundPayment

type AddressState struct {
	State        string    `json:"state"`
	Epoch        uint64    `json:"epoch"`
	BlockHeight  uint64    `json:"blockHeight"`
	BlockHash    string    `json:"blockHash"`
	TxHash       string    `json:"txHash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	IsValidation bool      `json:"isValidation"`
}

type AddressBurntCoins struct {
	Address string          `json:"address,omitempty"`
	Amount  decimal.Decimal `json:"amount"`
}

type BadAuthor struct {
	Epoch      uint64 `json:"epoch,omitempty"`
	Address    string `json:"address,omitempty"`
	WrongWords bool   `json:"wrongWords"`
	Reason     string `json:"reason" enums:"NoQualifiedFlips,QualifiedByNone,WrongWords"`
	PrevState  string `json:"prevState" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
	State      string `json:"state" enums:"Undefined,Invite,Candidate,Verified,Suspended,Killed,Zombie,Newbie,Human"`
} // @Name BadAuthor

type AdjacentStrValues struct {
	Prev AdjacentStrValue `json:"prev"`
	Next AdjacentStrValue `json:"next"`
}

type AdjacentStrValue struct {
	Value  string `json:"value"`
	Cycled bool   `json:"cycled"`
}

type FlipWithRewardFlag struct {
	FlipSummary
	Rewarded bool `json:"rewarded"`
} // @Name RewardedFlip

type ReportedFlipReward struct {
	Cid     string          `json:"cid"`
	Balance decimal.Decimal `json:"balance" swaggertype:"string"`
	Stake   decimal.Decimal `json:"stake" swaggertype:"string"`
} // @Name ReportedFlipReward

type InviteWithRewardFlag struct {
	Invite
	RewardType string `json:"rewardType,omitempty" enums:",Invitations,Invitations2,Invitations3"`
} // @Name RewardedInvite

type EpochInvites struct {
	Epoch   uint64 `json:"epoch"`
	Invites uint8  `json:"invites"`
} // @Name EpochInvites

type BalanceUpdate struct {
	BalanceOld  decimal.Decimal `json:"balanceOld"`
	StakeOld    decimal.Decimal `json:"stakeOld"`
	PenaltyOld  decimal.Decimal `json:"penaltyOld"`
	BalanceNew  decimal.Decimal `json:"balanceNew"`
	StakeNew    decimal.Decimal `json:"stakeNew"`
	PenaltyNew  decimal.Decimal `json:"penaltyNew"`
	Reason      string          `json:"reason"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Timestamp   time.Time       `json:"timestamp"`
	Data        interface{}     `json:"data,omitempty"`
}

type TransactionBalanceUpdate struct {
	TxHash string `json:"txHash"`
}

type CommitteeRewardBalanceUpdate struct {
	LastBlockHeight    uint64          `json:"lastBlockHeight"`
	LastBlockHash      string          `json:"lastBlockHash"`
	LastBlockTimestamp time.Time       `json:"lastBlockTimestamp"`
	RewardShare        decimal.Decimal `json:"rewardShare"`
	BlocksCount        uint32          `json:"blocksCount"`
}

type StrValueCount struct {
	Value string `json:"value"`
	Count uint32 `json:"count"`
}
