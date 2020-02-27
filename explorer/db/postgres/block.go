package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/lib/pq"
)

const (
	blockQueryByHeight         = "blockByHeight.sql"
	blockQueryByHash           = "blockByHash.sql"
	blockTxsCountByHeightQuery = "blockTxsCountByHeight.sql"
	blockTxsCountByHashQuery   = "blockTxsCountByHash.sql"
	blockTxsByHeightQuery      = "blockTxsByHeight.sql"
	blockTxsByHashQuery        = "blockTxsByHash.sql"
	blockCoinsByHeightQuery    = "blockCoinsByHeight.sql"
	blockCoinsByHashQuery      = "blockCoinsByHash.sql"
)

func (a *postgresAccessor) BlockByHeight(height uint64) (types.BlockDetail, error) {
	return a.block(blockQueryByHeight, height)
}

func (a *postgresAccessor) BlockByHash(hash string) (types.BlockDetail, error) {
	return a.block(blockQueryByHash, hash)
}

func (a *postgresAccessor) block(query string, id interface{}) (types.BlockDetail, error) {
	res := types.BlockDetail{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(query), id).Scan(
		&res.Epoch,
		&res.Height,
		&res.Hash,
		&timestamp,
		&res.TxCount,
		&res.ValidatorsCount,
		&res.Proposer,
		&res.ProposerVrfScore,
		&res.IsEmpty,
		&res.BodySize,
		&res.FullSize,
		&res.VrfProposerThreshold,
		&res.FeeRate,
		pq.Array(&res.Flags),
	)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.BlockDetail{}, err
	}
	res.Timestamp = timestampToTimeUTC(timestamp)
	return res, nil
}

func (a *postgresAccessor) BlockTxsCountByHeight(height uint64) (uint64, error) {
	return a.count(blockTxsCountByHeightQuery, height)
}

func (a *postgresAccessor) BlockTxsCountByHash(hash string) (uint64, error) {
	return a.count(blockTxsCountByHashQuery, hash)
}

func (a *postgresAccessor) BlockTxsByHeight(height uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(blockTxsByHeightQuery), height, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readTxs(rows)
}

func (a *postgresAccessor) BlockTxsByHash(hash string, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(blockTxsByHashQuery), hash, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readTxs(rows)
}

func (a *postgresAccessor) BlockCoinsByHeight(height uint64) (types.AllCoins, error) {
	return a.coins(blockCoinsByHeightQuery, height)
}

func (a *postgresAccessor) BlockCoinsByHash(hash string) (types.AllCoins, error) {
	return a.coins(blockCoinsByHashQuery, hash)
}
