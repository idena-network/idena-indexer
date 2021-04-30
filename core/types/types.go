package types

import (
	"github.com/shopspring/decimal"
	"time"
)

type OnlineIdentity struct {
	Address      string          `json:"address"`
	LastActivity *time.Time      `json:"lastActivity"`
	Penalty      decimal.Decimal `json:"penalty"`
	Online       bool            `json:"online"`
	Delegetee    *OnlineIdentity `json:"delegatee,omitempty"`
}

type UpgradeVotes struct {
	Upgrade uint32 `json:"upgrade"`
	Votes   uint64 `json:"votes"`
}

type TransactionSummary struct {
	Hash      string           `json:"hash"`
	Type      string           `json:"type,omitempty" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx,DeployContract,CallContract,TerminateContract,DelegateTx,UndelegateTx,KillDelegatorTx"`
	Timestamp time.Time        `json:"timestamp" example:"2020-01-01T00:00:00Z"`
	From      string           `json:"from,omitempty"`
	To        string           `json:"to,omitempty"`
	Amount    *decimal.Decimal `json:"amount,omitempty" swaggertype:"string"`
	Tips      *decimal.Decimal `json:"tips,omitempty" swaggertype:"string"`
	MaxFee    *decimal.Decimal `json:"maxFee,omitempty" swaggertype:"string"`
	Fee       *decimal.Decimal `json:"fee,omitempty" swaggertype:"string"`
	Size      uint32           `json:"size,omitempty"`
	// Deprecated
	Transfer *decimal.Decimal `json:"transfer,omitempty" swaggerignore:"true"`
	Data     interface{}      `json:"data,omitempty"`

	TxReceipt *TxReceipt `json:"txReceipt,omitempty"`
}

type TransactionDetail struct {
	Epoch       uint64          `json:"epoch"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Hash        string          `json:"hash"`
	Type        string          `json:"type" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx,DeployContract,CallContract,TerminateContract,DelegateTx,UndelegateTx,KillDelegatorTx"`
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

	TxReceipt *TxReceipt `json:"txReceipt,omitempty"`
}

type TxReceipt struct {
	Success  bool            `json:"success"`
	GasUsed  uint64          `json:"gasUsed"`
	GasCost  decimal.Decimal `json:"gasCost" swaggertype:"string"`
	Method   string          `json:"method,omitempty"`
	ErrorMsg string          `json:"errorMsg,omitempty"`
}
