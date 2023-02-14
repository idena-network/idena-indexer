package indexer

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/vm"
	"github.com/idena-network/idena-indexer/log"
	"github.com/shopspring/decimal"
)

func detectOracleVotingsToProlong(detector OracleVotingToProlongDetector, contracts map[common.Address]struct{}, appState *appstate.AppState, head *types.Header, config *config.Config) []common.Address {
	log.Debug(fmt.Sprintf("contracts to detect votings to prolong: %v", len(contracts)))
	var toProlong []common.Address
	for activeVoting := range contracts {
		if !detector.CanBeProlonged(activeVoting, appState, head, config) {
			continue
		}
		toProlong = append(toProlong, activeVoting)
	}
	log.Debug(fmt.Sprintf("oracle votings ready to prolong: %v", len(toProlong)))
	return toProlong
}

type OracleVotingToProlongDetector interface {
	CanBeProlonged(contractAddress common.Address, appState *appstate.AppState, head *types.Header, config *config.Config) bool
}

func NewOracleVotingToProlongDetector() OracleVotingToProlongDetector {
	return &oracleVotingToProlongDetectorImpl{}
}

type oracleVotingToProlongDetectorImpl struct {
}

func (d *oracleVotingToProlongDetectorImpl) CanBeProlonged(contractAddress common.Address, appState *appstate.AppState, head *types.Header, config *config.Config) bool {
	vm := vm.NewVmImpl(appState, nil, head, nil, config)
	from := common.Address{}
	payload, _ := attachments.CreateCallContractAttachment("prolongVoting").ToBytes()
	fakeNonce := uint32(1)
	fakeEpoch := uint16(1)
	tx := blockchain.BuildTxWithFeeEstimating(appState, from, &contractAddress, types.CallContractTx, decimal.Zero, decimal.Zero, decimal.Zero, fakeNonce, fakeEpoch, payload)
	receipt := vm.Run(tx, nil, -1)
	return receipt.Success
}
