package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	epochQuery                       = "epoch.sql"
	lastEpochQuery                   = "lastEpoch.sql"
	epochBlocksCountQuery            = "epochBlocksCount.sql"
	epochBlocksQuery                 = "epochBlocks.sql"
	epochFlipsCountQuery             = "epochFlipsCount.sql"
	epochFlipsQuery                  = "epochFlips.sql"
	epochFlipStatesQuery             = "epochFlipStates.sql"
	epochFlipQualifiedAnswersQuery   = "epochFlipQualifiedAnswers.sql"
	epochIdentityStatesSummaryQuery  = "epochIdentityStatesSummary.sql"
	epochIdentitiesQueryCount        = "epochIdentitiesCount.sql"
	epochIdentitiesQuery             = "epochIdentities.sql"
	epochInvitesCountQuery           = "epochInvitesCount.sql"
	epochInvitesQuery                = "epochInvites.sql"
	epochInvitesSummaryQuery         = "epochInvitesSummary.sql"
	epochTxsCountQuery               = "epochTxsCount.sql"
	epochTxsQuery                    = "epochTxs.sql"
	epochCoinsQuery                  = "epochCoins.sql"
	epochRewardsSummaryQuery         = "epochRewardsSummary.sql"
	epochBadAuthorsCountQuery        = "epochBadAuthorsCount.sql"
	epochBadAuthorsQuery             = "epochBadAuthors.sql"
	epochGoodAuthorsCountQuery       = "epochGoodAuthorsCount.sql"
	epochGoodAuthorsQuery            = "epochGoodAuthors.sql"
	epochRewardsCountQuery           = "epochRewardsCount.sql"
	epochRewardsQuery                = "epochRewards.sql"
	epochIdentitiesRewardsCountQuery = "epochIdentitiesRewardsCount.sql"
	epochIdentitiesRewardsQuery      = "epochIdentitiesRewards.sql"
	epochFundPaymentsQuery           = "epochFundPayments.sql"
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
	return a.flips(epochFlipsQuery, epoch, startIndex, count)
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

func (a *postgresAccessor) EpochRewardsSummary(epoch uint64) (types.RewardsSummary, error) {
	res := types.RewardsSummary{}
	err := a.db.QueryRow(a.getQuery(epochRewardsSummaryQuery), epoch).
		Scan(&res.Epoch,
			&res.Total,
			&res.Validation,
			&res.Flips,
			&res.Invitations,
			&res.FoundationPayouts,
			&res.ZeroWalletFund)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.RewardsSummary{}, err
	}
	return res, nil
}

func (a *postgresAccessor) EpochBadAuthorsCount(epoch uint64) (uint64, error) {
	return a.count(epochBadAuthorsCountQuery, epoch)
}

func (a *postgresAccessor) EpochBadAuthors(epoch uint64, startIndex uint64, count uint64) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(epochBadAuthorsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochGoodAuthorsCount(epoch uint64) (uint64, error) {
	return a.count(epochGoodAuthorsCountQuery, epoch)
}

func (a *postgresAccessor) EpochGoodAuthors(epoch uint64, startIndex uint64, count uint64) ([]types.AuthorValidationSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochGoodAuthorsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.AuthorValidationSummary
	for rows.Next() {
		item := types.AuthorValidationSummary{}
		if err := rows.Scan(&item.Address,
			&item.StrongFlips,
			&item.WeakFlips,
			&item.SuccessfulInvites); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochRewardsCount(epoch uint64) (uint64, error) {
	return a.count(epochRewardsCountQuery, epoch)
}

func (a *postgresAccessor) EpochRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Reward, error) {
	return a.rewards(epochRewardsQuery, epoch, startIndex, count)
}

func (a *postgresAccessor) EpochIdentitiesRewardsCount(epoch uint64) (uint64, error) {
	return a.count(epochIdentitiesRewardsCountQuery, epoch)
}

func (a *postgresAccessor) EpochIdentitiesRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Rewards, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentitiesRewardsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Rewards
	var item *types.Rewards
	for rows.Next() {
		reward := types.Reward{}
		var address string
		if err := rows.Scan(&address, &reward.Balance, &reward.Stake, &reward.Type); err != nil {
			return nil, err
		}
		if item == nil || item.Address != address {
			if item != nil {
				res = append(res, *item)
			}
			item = &types.Rewards{
				Address: address,
			}
		}
		item.Rewards = append(item.Rewards, reward)
	}
	if item != nil {
		res = append(res, *item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochFundPayments(epoch uint64) ([]types.FundPayment, error) {
	rows, err := a.db.Query(a.getQuery(epochFundPaymentsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.FundPayment
	for rows.Next() {
		item := types.FundPayment{}
		if err := rows.Scan(&item.Address, &item.Balance, &item.Type); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
