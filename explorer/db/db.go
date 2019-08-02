package db

import (
	"github.com/idena-network/idena-indexer/explorer/types"
)

type Accessor interface {
	Epochs() ([]types.EpochSummary, error)
	Epoch(epoch uint64) (types.EpochDetail, error)
	EpochBlocks(epoch uint64) ([]types.BlockSummary, error)
	EpochTxs(epoch uint64) ([]types.TransactionSummary, error)
	Block(height uint64) (types.BlockDetail, error)
	BlockTxs(height uint64) ([]types.TransactionSummary, error)
	EpochFlips(epoch uint64) ([]types.FlipSummary, error)
	EpochInvites(epoch uint64) ([]types.Invite, error)
	EpochIdentities(epoch uint64) ([]types.EpochIdentitySummary, error)
	Flip(hash string) (types.Flip, error)
	Identity(address string) (types.Identity, error)
	EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error)
	Address(address string) (types.Address, error)
	Transaction(hash string) (types.TransactionDetail, error)
	Destroy()
}
