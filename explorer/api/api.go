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

func (api *api) epochs() ([]types.EpochSummary, error) {
	return api.db.Epochs()
}

func (api *api) epoch(epoch uint64) (types.EpochDetail, error) {
	return api.db.Epoch(epoch)
}

func (api *api) epochBlocks(epoch uint64) ([]types.Block, error) {
	return api.db.EpochBlocks(epoch)
}

func (api *api) epochTxs(epoch uint64) ([]types.Transaction, error) {
	return api.db.EpochTxs(epoch)
}

func (api *api) blockTxs(height uint64) ([]types.Transaction, error) {
	return api.db.BlockTxs(height)
}

func (api *api) epochFlips(epoch uint64) ([]types.FlipSummary, error) {
	return api.db.EpochFlips(epoch)
}

func (api *api) epochInvites(epoch uint64) ([]types.Invite, error) {
	return api.db.EpochInvites(epoch)
}

func (api *api) epochIdentities(epoch uint64) ([]types.EpochIdentity, error) {
	return api.db.EpochIdentities(epoch)
}
