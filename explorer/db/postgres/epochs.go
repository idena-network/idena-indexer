package postgres

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"math/big"
)

const (
	epochsCountQuery = "epochsCount.sql"
	epochsQuery      = "epochs.sql"
)

func (a *postgresAccessor) EpochsCount() (uint64, error) {
	return a.count(epochsCountQuery)
}

func (a *postgresAccessor) Epochs(startIndex uint64, count uint64) ([]types.EpochSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochsQuery), startIndex, count)
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
		err = rows.Scan(&epoch.Epoch,
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
			&epoch.Rewards.ZeroWalletFund)
		if err != nil {
			return nil, err
		}
		epoch.ValidationTime = common.TimestampToTime(big.NewInt(validationTime))
		epochs = append(epochs, epoch)
	}
	return epochs, nil
}
