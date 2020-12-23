package service

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	db2 "github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
	"math/big"
)

type Contracts interface {
	OracleVotingContracts(authorAddress, oracleAddress string, states []string, all bool, count uint64, continuationToken *string) ([]types.OracleVotingContract, *string, error)
	OracleVotingContract(address, oracle string) (types.OracleVotingContract, error)
	AddressContractTxBalanceUpdates(address string, contractAddress string, count uint64, continuationToken *string) ([]types.ContractTxBalanceUpdate, *string, error)
}

type ContractsMemPool interface {
	GetOracleVotingContractDeploys(author common.Address) ([]db2.OracleVotingContract, error)
	GetAddressContractTxs(address, contractAddress string) ([]db2.Transaction, error)
}

type contractsImpl struct {
	dbAccessor       db.Accessor
	contractsMemPool ContractsMemPool
}

func NewContracts(dbAccessor db.Accessor, contractsMemPool ContractsMemPool) Contracts {
	return &contractsImpl{
		dbAccessor:       dbAccessor,
		contractsMemPool: contractsMemPool,
	}
}

func (c *contractsImpl) OracleVotingContracts(authorAddress, oracleAddress string, states []string, all bool, count uint64, continuationToken *string) ([]types.OracleVotingContract, *string, error) {
	var res []types.OracleVotingContract

	const pending = "Pending"
	includePending := false
	for _, state := range states {
		if state == pending {
			includePending = true
			break
		}
	}
	if len(authorAddress) > 0 && continuationToken == nil && includePending && all {
		memPoolContracts, _ := c.contractsMemPool.GetOracleVotingContractDeploys(common.HexToAddress(authorAddress))
		for _, memPoolContract := range memPoolContracts {
			var minPayment *decimal.Decimal
			if memPoolContract.VotingMinPayment != nil {
				v := blockchain.ConvertToFloat(memPoolContract.VotingMinPayment)
				minPayment = &v
			}
			oracleVotingContract := types.OracleVotingContract{
				ContractAddress:      memPoolContract.ContractAddress.Hex(),
				Author:               authorAddress,
				Fact:                 memPoolContract.Fact,
				State:                pending,
				StartTime:            common.TimestampToTime(big.NewInt(int64(memPoolContract.StartTime))).UTC(),
				MinPayment:           minPayment,
				Quorum:               memPoolContract.Quorum,
				CommitteeSize:        memPoolContract.CommitteeSize,
				VotingDuration:       memPoolContract.VotingDuration,
				PublicVotingDuration: memPoolContract.PublicVotingDuration,
				WinnerThreshold:      memPoolContract.WinnerThreshold,
			}
			res = append(res, oracleVotingContract)
		}
	}

	count = count - uint64(len(res))
	var nextContinuationToken *string
	var err error
	if count > 0 {
		var dbRes []types.OracleVotingContract
		dbRes, nextContinuationToken, err = c.dbAccessor.OracleVotingContracts(authorAddress, oracleAddress, states, all, count, continuationToken)
		res = append(res, dbRes...)
	}
	return res, nextContinuationToken, err
}

func (c *contractsImpl) OracleVotingContract(address, oracle string) (types.OracleVotingContract, error) {
	return c.dbAccessor.OracleVotingContract(address, oracle)
}

var contractTxTypes = map[uint16]string{
	15: "DeployContract",
	16: "CallContract",
	17: "TerminateContract",
}

func (c *contractsImpl) AddressContractTxBalanceUpdates(address string, contractAddress string, count uint64, continuationToken *string) ([]types.ContractTxBalanceUpdate, *string, error) {
	var res []types.ContractTxBalanceUpdate
	if continuationToken == nil {
		memPoolTxs, _ := c.contractsMemPool.GetAddressContractTxs(address, contractAddress)
		for _, memPoolTx := range memPoolTxs {
			bu := types.ContractTxBalanceUpdate{
				Hash:            memPoolTx.Hash,
				Type:            contractTxTypes[memPoolTx.Type],
				From:            memPoolTx.From,
				To:              memPoolTx.To,
				Amount:          memPoolTx.Amount,
				Tips:            memPoolTx.Tips,
				MaxFee:          memPoolTx.MaxFee,
				Address:         memPoolTx.From,
				ContractAddress: contractAddress,
			}
			res = append(res, bu)
		}
	}

	count = count - uint64(len(res))
	var nextContinuationToken *string
	var err error
	if count > 0 {
		var dbRes []types.ContractTxBalanceUpdate
		dbRes, nextContinuationToken, err = c.dbAccessor.AddressContractTxBalanceUpdates(address, contractAddress, count, continuationToken)
		res = append(res, dbRes...)
	}
	return res, nextContinuationToken, err
}
