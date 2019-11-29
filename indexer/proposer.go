package indexer

import (
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/pengings"
	"github.com/idena-network/idena-indexer/core/conversion"
	"math/big"
	"sync"
)

func getProposer(block *types.Block) string {
	if block.IsEmpty() {
		return ""
	}
	return conversion.ConvertAddress(block.Header.ProposedHeader.Coinbase)
}

func getProposerVrfScore(block *types.Block, proofsByRound, pendingProofs *sync.Map) (float64, bool) {
	if block.Header.ProposedHeader == nil {
		return 0, false
	}
	var hash common.Hash
	var ok bool
	if hash, ok = searchVrfScore(block.Height(), block.Header.ProposedHeader.Coinbase, proofsByRound); !ok {
		if hash, ok = searchVrfScore(block.Height(), block.Header.ProposedHeader.Coinbase, pendingProofs); !ok {
			return 0, false
		}
	}
	v := new(big.Float).SetInt(new(big.Int).SetBytes(hash[:]))
	q := new(big.Float).Quo(v, blockchain.MaxHash).SetPrec(10)
	f, _ := q.Float64()
	return f, true
}

func searchVrfScore(round uint64, address common.Address, proofsByRound *sync.Map) (common.Hash, bool) {
	m, ok := proofsByRound.Load(round)
	if !ok {
		return common.Hash{}, false
	}
	proofsByHash := m.(*sync.Map)
	ok = false
	var hash common.Hash
	proofsByHash.Range(func(key, value interface{}) bool {
		if proofAddress, _ := crypto.PubKeyBytesToAddress(value.(*pengings.Proof).PubKey); proofAddress == address {
			hash = value.(*pengings.Proof).Hash
			ok = true
			return false
		}
		return true
	})
	return hash, ok
}
