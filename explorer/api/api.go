package api

import (
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/types"
)

type api struct {
	db db.Accessor
}

func newApi(db db.Accessor) *api {
	return &api{
		db: db,
	}
}

func (a *api) summary() (types.Summary, error) {
	return a.db.Summary()
}

func (a *api) epochs() ([]types.EpochSummary, error) {
	return a.db.Epochs()
}

func (a *api) epoch(epoch uint64) (types.EpochDetail, error) {
	return a.db.Epoch(epoch)
}

func (a *api) epochBlocks(epoch uint64) ([]types.BlockSummary, error) {
	return a.db.EpochBlocks(epoch)
}

func (a *api) epochTxs(epoch uint64) ([]types.TransactionSummary, error) {
	return a.db.EpochTxs(epoch)
}

func (a *api) block(height uint64) (types.BlockDetail, error) {
	return a.db.Block(height)
}

func (a *api) blockTxs(height uint64) ([]types.TransactionSummary, error) {
	return a.db.BlockTxs(height)
}

func (a *api) epochFlips(epoch uint64) ([]types.FlipSummary, error) {
	return a.db.EpochFlips(epoch)
}

func (a *api) epochInvites(epoch uint64) ([]types.Invite, error) {
	return a.db.EpochInvites(epoch)
}

func (a *api) epochIdentities(epoch uint64) ([]types.EpochIdentitySummary, error) {
	return a.db.EpochIdentities(epoch)
}

func (a *api) flip(hash string) (types.Flip, error) {
	return a.db.Flip(hash)
}

func (a *api) identity(address string) (types.Identity, error) {
	return a.db.Identity(address)
}

func (a *api) epochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	return a.db.EpochIdentity(epoch, address)
}

func (a *api) address(address string) (types.Address, error) {
	return a.db.Address(address)
}

func (a *api) transaction(hash string) (types.TransactionDetail, error) {
	return a.db.Transaction(hash)
}
