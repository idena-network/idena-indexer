package indexer

import (
	"github.com/idena-network/idena-go/common"
	"math/big"
)

type indexerState struct {
	lastIndexedHeight uint64
	totalBalance      *big.Int
	totalStake        *big.Int

	actualOracleVotingHolder *actualOracleVotingHolder
}

type actualOracleVotingHolder struct {
	contracts map[common.Address]struct{}
}

func newActualOracleVotingHolder(contracts []common.Address) *actualOracleVotingHolder {
	res := &actualOracleVotingHolder{
		contracts: map[common.Address]struct{}{},
	}
	for _, contract := range contracts {
		res.contracts[contract] = struct{}{}
	}
	return res
}

func (h *actualOracleVotingHolder) add(contracts []common.Address) {
	for _, contract := range contracts {
		h.contracts[contract] = struct{}{}
	}
}

func (h *actualOracleVotingHolder) remove(contracts []common.Address) {
	for _, contract := range contracts {
		delete(h.contracts, contract)
	}
}
