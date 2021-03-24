package indexer

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/vrf"
	"github.com/idena-network/idena-go/pengings"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/migration/runtime"
	"sync"
)

func getProposer(block *types.Block) string {
	if block.IsEmpty() {
		return ""
	}
	return conversion.ConvertAddress(block.Header.Coinbase())
}

func getProposerVrfScore(
	block *types.Block,
	proposerByRound pengings.ProposerByRound,
	pendingProofs *sync.Map,
	secondaryStorage *runtime.SecondaryStorage,
	validatorsCache *validators.ValidatorsCache,
) (float64, bool) {
	if block.Header.ProposedHeader == nil {
		return 0, false
	}
	if secondaryStorage != nil {
		score, err := secondaryStorage.GetProposerVrfScore(block.Height())
		if err != nil {
			log.Error("Unable to get proposer vrf score from previous db. Skipped.", "height",
				block.Height(), "err", err)
			return 0, false
		}
		return score, true
	}
	var hash common.Hash
	var ok bool
	coinbase := block.Header.Coinbase()
	if hash, ok = getProposerScoreByRound(block.Height(), coinbase, proposerByRound); !ok {
		if hash, ok = searchProofsByHashVrfScore(block.Height(), coinbase, pendingProofs); !ok {
			return 0, false
		}
	}
	modifier := int64(validatorsCache.PoolSize(coinbase))
	q := common.HashToFloat(hash, modifier)
	f, _ := q.Float64()
	return f, true
}

func getProposerScoreByRound(round uint64, address common.Address, proposerByRound pengings.ProposerByRound) (common.Hash, bool) {
	hash, proposerPubKey, ok := proposerByRound(round)
	if !ok {
		return common.Hash{}, false
	}
	if proposerAddress, _ := crypto.PubKeyBytesToAddress(proposerPubKey); proposerAddress != address {
		return common.Hash{}, false
	}
	return hash, true
}

func searchProofsByHashVrfScore(round uint64, address common.Address, proofsByHash *sync.Map) (common.Hash, bool) {
	found := false
	var hash common.Hash
	proofsByHash.Range(func(key, value interface{}) bool {
		proofProposal, ok := value.(*types.ProofProposal)
		if !ok {
			log.Error("proofsByHash value is not *types.ProofProposal")
			return true
		}
		if proofProposal.Round != round {
			return true
		}
		pubKey, err := types.ProofProposalPubKey(proofProposal)
		if err != nil {
			return true
		}
		if proofAddress, _ := crypto.PubKeyBytesToAddress(pubKey); proofAddress == address {
			h, err := vrf.HashFromProof(proofProposal.Proof)
			if err != nil {
				return true
			}
			hash = h
			found = true
			return false
		}
		return true
	})
	return hash, found
}
