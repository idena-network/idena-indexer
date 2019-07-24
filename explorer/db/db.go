package db

import (
	"github.com/idena-network/idena-indexer/explorer/types"
)

type Accessor interface {
	Summary() (types.Summary, error)
	Epochs() ([]types.EpochSummary, error)
	Epoch(epoch uint64) (types.EpochDetail, error)
	EpochBlocks(epoch uint64) ([]types.Block, error)
	EpochTxs(epoch uint64) ([]types.Transaction, error)
	BlockTxs(height uint64) ([]types.Transaction, error)
	EpochFlips(epoch uint64) ([]types.FlipSummary, error)
	EpochInvites(epoch uint64) ([]types.Invite, error)
	EpochIdentities(epoch uint64) ([]types.EpochIdentitySummary, error)
	Flip(hash string) (types.Flip, error)
	Identity(address string) (types.Identity, error)
	EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error)
	Address(address string) (types.Address, error)
	Destroy()
}
