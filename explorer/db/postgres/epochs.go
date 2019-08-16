package postgres

import "github.com/idena-network/idena-indexer/explorer/types"

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
			Coins: types.AllCoins{},
		}
		err = rows.Scan(&epoch.Epoch,
			&epoch.ValidatedCount,
			&epoch.BlockCount,
			&epoch.TxCount,
			&epoch.InviteCount,
			&epoch.FlipCount,
			&epoch.Coins.Balance.Burnt,
			&epoch.Coins.Balance.Minted,
			&epoch.Coins.Balance.Total,
			&epoch.Coins.Stake.Burnt,
			&epoch.Coins.Stake.Minted,
			&epoch.Coins.Stake.Total)
		if err != nil {
			return nil, err
		}
		epochs = append(epochs, epoch)
	}
	return epochs, nil
}
