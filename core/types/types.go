package types

import (
	"github.com/shopspring/decimal"
	"time"
)

type OnlineIdentity struct {
	Address        string          `json:"address"`
	LastActivity   *time.Time      `json:"lastActivity"`
	Penalty        decimal.Decimal `json:"penalty"`
	PenaltySeconds uint16          `json:"penaltySeconds"`
	Online         bool            `json:"online"`
	Delegetee      *OnlineIdentity `json:"delegatee,omitempty"`
}

type Pool struct {
	TotalStake          decimal.Decimal `json:"totalStake" swaggertype:"string"`
	TotalValidatedStake decimal.Decimal `json:"totalValidatedStake" swaggertype:"string"`
}

type Validator struct {
	Address        string          `json:"address"`
	Size           uint32          `json:"size"`
	Online         bool            `json:"online"`
	LastActivity   *time.Time      `json:"lastActivity"`
	Penalty        decimal.Decimal `json:"penalty"`
	PenaltySeconds uint16          `json:"penaltySeconds"`
	IsPool         bool            `json:"isPool"`
}

type UpgradeVotes struct {
	Upgrade uint32 `json:"upgrade"`
	Votes   uint64 `json:"votes"`
}

type TransactionSummary struct {
	Hash   string           `json:"hash"`
	Type   string           `json:"type,omitempty" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx,DeployContract,CallContract,TerminateContract,DelegateTx,UndelegateTx,KillDelegatorTx,StoreToIpfsTx"`
	From   string           `json:"from,omitempty"`
	To     string           `json:"to,omitempty"`
	Amount *decimal.Decimal `json:"amount,omitempty"`
	Tips   *decimal.Decimal `json:"tips,omitempty"`
	MaxFee *decimal.Decimal `json:"maxFee,omitempty"`
	Fee    *decimal.Decimal `json:"fee,omitempty"`
	Size   uint32           `json:"size,omitempty"`
	Nonce  uint32           `json:"nonce,omitempty"`
	// Deprecated
	Transfer *decimal.Decimal `json:"transfer,omitempty"`
	Data     interface{}      `json:"data,omitempty"`

	TxReceipt *TxReceipt `json:"txReceipt,omitempty"`
}

type TransactionDetail struct {
	Epoch       uint64          `json:"epoch"`
	BlockHeight uint64          `json:"blockHeight"`
	BlockHash   string          `json:"blockHash"`
	Hash        string          `json:"hash"`
	Type        string          `json:"type" enums:"SendTx,ActivationTx,InviteTx,KillTx,SubmitFlipTx,SubmitAnswersHashTx,SubmitShortAnswersTx,SubmitLongAnswersTx,EvidenceTx,OnlineStatusTx,KillInviteeTx,ChangeGodAddressTx,BurnTx,ChangeProfileTx,DeleteFlipTx,DeployContract,CallContract,TerminateContract,DelegateTx,UndelegateTx,KillDelegatorTx,StoreToIpfsTx"`
	From        string          `json:"from"`
	To          string          `json:"to,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
	Tips        decimal.Decimal `json:"tips"`
	MaxFee      decimal.Decimal `json:"maxFee"`
	Fee         decimal.Decimal `json:"fee"`
	Size        uint32          `json:"size"`
	Nonce       uint32          `json:"nonce,omitempty"`
	// Deprecated
	Transfer *decimal.Decimal `json:"transfer,omitempty"`
	Data     interface{}      `json:"data,omitempty"`

	TxReceipt *TxReceipt `json:"txReceipt,omitempty"`
}

type TxReceipt struct {
	Success  bool            `json:"success"`
	GasUsed  uint64          `json:"gasUsed"`
	GasCost  decimal.Decimal `json:"gasCost"`
	Method   string          `json:"method,omitempty"`
	ErrorMsg string          `json:"errorMsg,omitempty"`
}

type Multisig struct {
	Signers []MultisigSigner `json:"signers,omitempty"`
}

type MultisigSigner struct {
	Address     string          `json:"address"`
	DescAddress string          `json:"destAddress"`
	Amount      decimal.Decimal `json:"amount"`
}

type Staking struct {
	Weight             float64 `json:"weight"`
	MinersWeight       float64 `json:"minersWeight"`
	AverageMinerWeight float64 `json:"averageMinerWeight"`
	MaxMinerWeight     float64 `json:"maxMinerWeight"`
}
