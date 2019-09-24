package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	epochQuery                      = "epoch.sql"
	lastEpochQuery                  = "lastEpoch.sql"
	epochBlocksCountQuery           = "epochBlocksCount.sql"
	epochBlocksQuery                = "epochBlocks.sql"
	epochFlipsCountQuery            = "epochFlipsCount.sql"
	epochFlipsQuery                 = "epochFlips.sql"
	epochFlipStatesQuery            = "epochFlipStates.sql"
	epochFlipQualifiedAnswersQuery  = "epochFlipQualifiedAnswers.sql"
	epochIdentityStatesSummaryQuery = "epochIdentityStatesSummary.sql"
	epochIdentitiesQueryCount       = "epochIdentitiesCount.sql"
	epochIdentitiesQuery            = "epochIdentities.sql"
	epochInvitesCountQuery          = "epochInvitesCount.sql"
	epochInvitesQuery               = "epochInvites.sql"
	epochInvitesSummaryQuery        = "epochInvitesSummary.sql"
	epochTxsCountQuery              = "epochTxsCount.sql"
	epochTxsQuery                   = "epochTxs.sql"
	epochCoinsQuery                 = "epochCoins.sql"
)

func (a *postgresAccessor) LastEpoch() (types.EpochDetail, error) {
	return a.epoch(lastEpochQuery)
}

func (a *postgresAccessor) Epoch(epoch uint64) (types.EpochDetail, error) {
	return a.epoch(epochQuery, epoch)
}

func (a *postgresAccessor) epoch(queryName string, args ...interface{}) (types.EpochDetail, error) {
	res := types.EpochDetail{}
	var validationTime int64
	err := a.db.QueryRow(a.getQuery(queryName), args...).Scan(&res.Epoch,
		&validationTime,
		&res.ValidationFirstBlockHeight)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochDetail{}, err
	}
	res.ValidationTime = common.TimestampToTime(big.NewInt(validationTime))
	return res, nil
}

func (a *postgresAccessor) EpochBlocksCount(epoch uint64) (uint64, error) {
	return a.count(epochBlocksCountQuery, epoch)
}

func (a *postgresAccessor) EpochBlocks(epoch uint64, startIndex uint64, count uint64) ([]types.BlockSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochBlocksQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []types.BlockSummary
	for rows.Next() {
		block := types.BlockSummary{
			Coins: types.AllCoins{},
		}
		var timestamp int64
		err = rows.Scan(&block.Height,
			&block.Hash,
			&timestamp,
			&block.TxCount,
			&block.Proposer,
			&block.IsEmpty,
			&block.Size,
			&block.Coins.Balance.Burnt,
			&block.Coins.Balance.Minted,
			&block.Coins.Balance.Total,
			&block.Coins.Stake.Burnt,
			&block.Coins.Stake.Minted,
			&block.Coins.Stake.Total)
		if err != nil {
			return nil, err
		}
		block.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (a *postgresAccessor) EpochFlipsCount(epoch uint64) (uint64, error) {
	return a.count(epochFlipsCountQuery, epoch)
}

func (a *postgresAccessor) EpochFlips(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return a.readFlips(rows)
}

func (a *postgresAccessor) EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipQualifiedAnswersQuery, epoch)
}

func (a *postgresAccessor) EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipStatesQuery, epoch)
}

func (a *postgresAccessor) EpochIdentitiesCount(epoch uint64) (uint64, error) {
	return a.count(epochIdentitiesQueryCount, epoch)
}

func (a *postgresAccessor) EpochIdentities(epoch uint64, startIndex uint64, count uint64) ([]types.EpochIdentitySummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentitiesQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return a.readEpochIdentitySummaries(rows)
}

func (a *postgresAccessor) EpochIdentityStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochIdentityStatesSummaryQuery, epoch)
}

func (a *postgresAccessor) EpochInvitesSummary(epoch uint64) (types.InvitesSummary, error) {
	res := types.InvitesSummary{}
	err := a.db.QueryRow(a.getQuery(epochInvitesSummaryQuery), epoch).Scan(&res.AllCount, &res.UsedCount)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.InvitesSummary{}, err
	}
	return res, nil
}

func (a *postgresAccessor) EpochInvitesCount(epoch uint64) (uint64, error) {
	return a.count(epochInvitesCountQuery, epoch)
}

func (a *postgresAccessor) EpochInvites(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(epochInvitesQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return a.readInvites(rows)
}

func (a *postgresAccessor) EpochTxsCount(epoch uint64) (uint64, error) {
	return a.count(epochTxsCountQuery, epoch)
}

func (a *postgresAccessor) EpochTxs(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochTxsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return a.readTxs(rows)
}

func (a *postgresAccessor) EpochCoins(epoch uint64) (types.AllCoins, error) {
	return a.coins(epochCoinsQuery, epoch)
}
