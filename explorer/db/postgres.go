package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
	log     log.Logger
}

const (
	epochsQuery            = "epochs.sql"
	epochQuery             = "epoch.sql"
	epochBlocksQuery       = "epochBlocks.sql"
	epochTxsQuery          = "epochTxs.sql"
	blockTxsQuery          = "blockTxs.sql"
	epochFlipsWithKeyQuery = "epochFlipsWithKey.sql"
	epochFlipsQuery        = "epochFlips.sql"
	epochInvitesQuery      = "epochInvites.sql"
	epochIdentitiesQuery   = "epochIdentities.sql"
)

type flipWithKey struct {
	cid string
	key string
}

var NoDataFound = errors.New("no data found")

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) Epochs() ([]types.EpochSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var epochs []types.EpochSummary
	for rows.Next() {
		epoch := types.EpochSummary{}
		err = rows.Scan(&epoch.Epoch, &epoch.VerifiedCount, &epoch.BlockCount, &epoch.FlipsCount)
		if err != nil {
			return nil, err
		}
		epochs = append(epochs, epoch)
	}
	return epochs, nil
}

func (a *postgresAccessor) Epoch(epoch uint64) (types.EpochDetail, error) {
	epochInfo := types.EpochDetail{}
	err := a.db.QueryRow(a.getQuery(epochQuery), epoch).Scan(&epochInfo.Epoch, &epochInfo.VerifiedCount, &epochInfo.BlockCount,
		&epochInfo.FlipsCount, &epochInfo.QualifiedFlipsCount, &epochInfo.WeaklyQualifiedFlipsCount)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochDetail{}, err
	}

	flipsWithKey, err := a.epochFlipsWithKey(epoch)
	if err != nil {
		return types.EpochDetail{}, err
	}
	epochInfo.FlipsWithKeyCount = uint32(len(flipsWithKey))

	return epochInfo, nil
}

func (a *postgresAccessor) EpochBlocks(epoch uint64) ([]types.Block, error) {
	rows, err := a.db.Query(a.getQuery(epochBlocksQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []types.Block
	for rows.Next() {
		block := types.Block{}
		var timestamp int64
		err = rows.Scan(&block.Height, &timestamp, &block.TxCount)
		if err != nil {
			return nil, err
		}
		block.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (a *postgresAccessor) EpochTxs(epoch uint64) ([]types.Transaction, error) {
	rows, err := a.db.Query(a.getQuery(epochTxsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txs []types.Transaction
	for rows.Next() {
		tx := types.Transaction{}
		var timestamp int64
		var amount, fee int64
		err = rows.Scan(&tx.Hash, &tx.Type, &timestamp, &tx.From, &tx.To, &amount, &fee)
		if err != nil {
			return nil, err
		}
		tx.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		tx.Amount = big.NewInt(amount)
		tx.Fee = big.NewInt(fee)
		txs = append(txs, tx)
	}
	return txs, nil
}

func (a *postgresAccessor) BlockTxs(height uint64) ([]types.Transaction, error) {
	rows, err := a.db.Query(a.getQuery(blockTxsQuery), height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var txs []types.Transaction
	for rows.Next() {
		tx := types.Transaction{}
		var timestamp int64
		var amount, fee int64
		err = rows.Scan(&tx.Hash, &tx.Type, &timestamp, &tx.From, &tx.To, &amount, &fee)
		if err != nil {
			return nil, err
		}
		tx.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		tx.Amount = big.NewInt(amount)
		tx.Fee = big.NewInt(fee)
		txs = append(txs, tx)
	}
	return txs, nil
}

func (a *postgresAccessor) epochFlipsWithKey(epoch uint64) ([]flipWithKey, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsWithKeyQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []flipWithKey
	for rows.Next() {
		item := flipWithKey{}
		err = rows.Scan(&item.cid, &item.key)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochFlips(epoch uint64) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.FlipSummary
	for rows.Next() {
		item := types.FlipSummary{}
		err = rows.Scan(&item.Cid, &item.Author, &item.Status, &item.ShortRespCount, &item.LongRespCount)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochInvites(epoch uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(epochInvitesQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Invite
	for rows.Next() {
		item := types.Invite{}
		// todo status (Not activated/Candidate)
		err = rows.Scan(&item.Id, &item.Author)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentities(epoch uint64) ([]types.EpochIdentity, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentitiesQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.EpochIdentity
	for rows.Next() {
		item := types.EpochIdentity{}
		err = rows.Scan(&item.Address, &item.State, &item.RespScore, &item.AuthorScore)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		a.log.Error("Unable to close db: %v", err)
	}
}
