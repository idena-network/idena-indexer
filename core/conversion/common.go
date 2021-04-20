package conversion

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
)

var (
	txTypeNames = map[uint16]string{
		types.SendTx:               "SendTx",
		types.ActivationTx:         "ActivationTx",
		types.InviteTx:             "InviteTx",
		types.KillTx:               "KillTx",
		types.SubmitFlipTx:         "SubmitFlipTx",
		types.SubmitAnswersHashTx:  "SubmitAnswersHashTx",
		types.SubmitShortAnswersTx: "SubmitShortAnswersTx",
		types.SubmitLongAnswersTx:  "SubmitLongAnswersTx",
		types.EvidenceTx:           "EvidenceTx",
		types.OnlineStatusTx:       "OnlineStatusTx",
		types.KillInviteeTx:        "KillInviteeTx",
		types.ChangeGodAddressTx:   "ChangeGodAddressTx",
		types.BurnTx:               "BurnTx",
		types.ChangeProfileTx:      "ChangeProfileTx",
		types.DeleteFlipTx:         "DeleteFlipTx",
		types.DeployContractTx:     "DeployContract",
		types.CallContractTx:       "CallContract",
		types.TerminateContractTx:  "TerminateContract",
		types.DelegateTx:           "DelegateTx",
		types.UndelegateTx:         "UndelegateTx",
		types.KillDelegatorTx:      "KillDelegatorTx",
		types.StoreToIpfsTx:        "StoreToIpfsTx",
	}
)

func ConvertAddress(address common.Address) string {
	return address.Hex()
}

func BytesToAddr(bytes []byte) common.Address {
	addr := common.Address{}
	addr.SetBytes(bytes[1:])
	return addr
}

func ConvertHash(hash common.Hash) string {
	return hash.Hex()
}

func ConvertTxType(txType uint16) string {
	return txTypeNames[txType]
}
