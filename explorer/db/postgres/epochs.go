package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	epochsCountQuery = "epochsCount.sql"
	epochsQuery      = "epochs.sql"
	epochsOldQuery   = "epochsOld.sql"
)

func (a *postgresAccessor) EpochsCount() (uint64, error) {
	return a.count(epochsCountQuery)
}

func (a *postgresAccessor) Epochs(count uint64, continuationToken *string) ([]types.EpochSummary, *string, error) {
	res, nextContinuationToken, err := a.page(epochsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.EpochSummary
		var epoch uint64
		for rows.Next() {
			item := types.EpochSummary{
				Coins:   types.AllCoins{},
				Rewards: types.RewardsSummary{},
			}
			var validationTime int64
			if err := rows.Scan(
				&epoch,
				&validationTime,
				&item.ValidatedCount,
				&item.BlockCount,
				&item.EmptyBlockCount,
				&item.TxCount,
				&item.InviteCount,
				&item.FlipCount,
				&item.Coins.Burnt,
				&item.Coins.Minted,
				&item.Coins.TotalBalance,
				&item.Coins.TotalStake,
				&item.Rewards.Total,
				&item.Rewards.Validation,
				&item.Rewards.Flips,
				&item.Rewards.Invitations,
				&item.Rewards.FoundationPayouts,
				&item.Rewards.ZeroWalletFund,
				&item.MinScoreForInvite,
			); err != nil {
				return nil, 0, err
			}
			item.ValidationTime = timestampToTimeUTC(validationTime)
			item.Epoch = epoch
			res = append(res, item)
		}
		return res, epoch, nil
	}, count, continuationToken)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.EpochSummary), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochsOld(startIndex uint64, count uint64) ([]types.EpochSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochsOldQuery), startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var epochs []types.EpochSummary
	for rows.Next() {
		epoch := types.EpochSummary{
			Coins:   types.AllCoins{},
			Rewards: types.RewardsSummary{},
		}
		var validationTime int64
		err = rows.Scan(
			&epoch.Epoch,
			&validationTime,
			&epoch.ValidatedCount,
			&epoch.BlockCount,
			&epoch.EmptyBlockCount,
			&epoch.TxCount,
			&epoch.InviteCount,
			&epoch.FlipCount,
			&epoch.Coins.Burnt,
			&epoch.Coins.Minted,
			&epoch.Coins.TotalBalance,
			&epoch.Coins.TotalStake,
			&epoch.Rewards.Total,
			&epoch.Rewards.Validation,
			&epoch.Rewards.Flips,
			&epoch.Rewards.Invitations,
			&epoch.Rewards.FoundationPayouts,
			&epoch.Rewards.ZeroWalletFund,
			&epoch.MinScoreForInvite,
		)
		if err != nil {
			return nil, err
		}
		epoch.ValidationTime = timestampToTimeUTC(validationTime)
		epochs = append(epochs, epoch)
	}
	return epochs, nil
}
