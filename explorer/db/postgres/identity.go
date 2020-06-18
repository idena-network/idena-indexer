package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	identityQuery                  = "identity.sql"
	identityAgeQuery               = "identityAge.sql"
	identityAnswerPointsQuery      = "identityAnswerPoints.sql"
	identityCurrentFlipsQuery      = "identityCurrentFlips.sql"
	identityEpochsCountQuery       = "identityEpochsCount.sql"
	identityEpochsQuery            = "identityEpochs.sql"
	identityEpochsOldQuery         = "identityEpochsOld.sql"
	identityFlipStatesQuery        = "identityFlipStates.sql"
	identityFlipRightAnswersQuery  = "identityFlipRightAnswers.sql"
	identityInvitesCountQuery      = "identityInvitesCount.sql"
	identityInvitesQuery           = "identityInvites.sql"
	identityInvitesOldQuery        = "identityInvitesOld.sql"
	identityTxsCountQuery          = "identityTxsCount.sql"
	identityTxsQuery               = "identityTxs.sql"
	identityTxsOldQuery            = "identityTxsOld.sql"
	identityRewardsCountQuery      = "identityRewardsCount.sql"
	identityRewardsQuery           = "identityRewards.sql"
	identityRewardsOldQuery        = "identityRewardsOld.sql"
	identityEpochRewardsCountQuery = "identityEpochRewardsCount.sql"
	identityEpochRewardsQuery      = "identityEpochRewards.sql"
	identityEpochRewardsOldQuery   = "identityEpochRewardsOld.sql"
	identityFlipsCountQuery        = "identityFlipsCount.sql"
	identityFlipsQuery             = "identityFlips.sql"
	identityFlipsOldQuery          = "identityFlipsOld.sql"
)

func (a *postgresAccessor) Identity(address string) (types.Identity, error) {
	identity := types.Identity{}
	err := a.db.QueryRow(a.getQuery(identityQuery), address).Scan(&identity.State)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Identity{}, err
	}
	identity.Address = address

	if identity.TotalShortAnswers, err = a.identityAnswerPoints(address); err != nil {
		return types.Identity{}, err
	}

	return identity, nil
}

func (a *postgresAccessor) identityAnswerPoints(address string) (totalShort types.IdentityAnswersSummary, err error) {
	rows, err := a.db.Query(a.getQuery(identityAnswerPointsQuery), address)
	if err != nil {
		return
	}
	defer rows.Close()
	if !rows.Next() {
		return
	}
	err = rows.Scan(&totalShort.Point, &totalShort.FlipsCount)
	return
}

func (a *postgresAccessor) IdentityAge(address string) (uint64, error) {
	var res uint64
	err := a.db.QueryRow(a.getQuery(identityAgeQuery), address).Scan(&res)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (a *postgresAccessor) IdentityCurrentFlipCids(address string) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(identityCurrentFlipsQuery), address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var item string
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) IdentityEpochsCount(address string) (uint64, error) {
	return a.count(identityEpochsCountQuery, address)
}

func (a *postgresAccessor) IdentityEpochs(address string, count uint64, continuationToken *string) ([]types.EpochIdentity, *string, error) {
	res, nextContinuationToken, err := a.page(identityEpochsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readEpochIdentities(rows)
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.EpochIdentity), nextContinuationToken, nil
}

func (a *postgresAccessor) IdentityEpochsOld(address string, startIndex uint64, count uint64) ([]types.EpochIdentity, error) {
	rows, err := a.db.Query(a.getQuery(identityEpochsOldQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readEpochIdentitiesOld(rows)
}

func (a *postgresAccessor) IdentityFlipStates(address string) ([]types.StrValueCount, error) {
	rows, err := a.db.Query(a.getQuery(identityFlipStatesQuery), address)
	if err != nil {
		return nil, err
	}
	return readStrValueCounts(rows)
}

func (a *postgresAccessor) IdentityFlipQualifiedAnswers(address string) ([]types.StrValueCount, error) {
	rows, err := a.db.Query(a.getQuery(identityFlipRightAnswersQuery), address)
	if err != nil {
		return nil, err
	}
	return readStrValueCounts(rows)
}

func (a *postgresAccessor) IdentityInvitesCount(address string) (uint64, error) {
	return a.count(identityInvitesCountQuery, address)
}

func (a *postgresAccessor) IdentityInvites(address string, count uint64, continuationToken *string) ([]types.Invite, *string, error) {
	res, nextContinuationToken, err := a.page(identityInvitesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readInvites(rows)
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.Invite), nextContinuationToken, nil
}

func (a *postgresAccessor) IdentityInvitesOld(address string, startIndex uint64, count uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(identityInvitesOldQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readInvitesOld(rows)
}

func (a *postgresAccessor) IdentityTxsCount(address string) (uint64, error) {
	return a.count(identityTxsCountQuery, address)
}

func (a *postgresAccessor) IdentityTxs(address string, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.page(identityTxsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readTxs(rows)
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.TransactionSummary), nextContinuationToken, nil
}

func (a *postgresAccessor) IdentityTxsOld(address string, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(identityTxsOldQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readTxsOld(rows)
}

func (a *postgresAccessor) IdentityRewardsCount(address string) (uint64, error) {
	return a.count(identityRewardsCountQuery, address)
}

func (a *postgresAccessor) IdentityRewards(address string, count uint64, continuationToken *string) ([]types.Reward, *string, error) {
	res, nextContinuationToken, err := a.page(identityRewardsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.Reward
		for rows.Next() {
			item := types.Reward{}
			if err := rows.Scan(&item.Address, &item.Epoch, &item.BlockHeight, &item.Balance, &item.Stake, &item.Type); err != nil {
				return nil, 0, err
			}
			res = append(res, item)
		}
		continuationId, _ := parseUintContinuationToken(continuationToken)
		if continuationId == nil {
			v := uint64(0)
			continuationId = &v
		}
		return res, *continuationId + count, nil
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.Reward), nextContinuationToken, nil
}

func (a *postgresAccessor) IdentityRewardsOld(address string, startIndex uint64, count uint64) ([]types.Reward, error) {
	return a.rewardsOld(identityRewardsOldQuery, address, startIndex, count)
}

func (a *postgresAccessor) IdentityEpochRewardsCount(address string) (uint64, error) {
	return a.count(identityEpochRewardsCountQuery, address)
}

func (a *postgresAccessor) IdentityEpochRewards(address string, count uint64, continuationToken *string) ([]types.Rewards, *string, error) {
	res, nextContinuationToken, err := a.page(identityEpochRewardsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.Rewards
		var item *types.Rewards
		var id uint64
		for rows.Next() {
			reward := types.Reward{}
			var epoch uint64
			var prevState, state string
			var age uint16
			if err := rows.Scan(&id, &epoch, &reward.Balance, &reward.Stake, &reward.Type, &prevState, &state, &age); err != nil {
				return nil, 0, err
			}
			if item == nil || item.Epoch != epoch {
				if item != nil {
					res = append(res, *item)
				}
				item = &types.Rewards{
					Epoch:     epoch,
					PrevState: prevState,
					State:     state,
					Age:       age,
				}
			}
			item.Rewards = append(item.Rewards, reward)
		}
		if item != nil {
			res = append(res, *item)
		}
		return res, id, nil
	}, count, continuationToken, address)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.Rewards), nextContinuationToken, nil
}

func (a *postgresAccessor) IdentityEpochRewardsOld(address string, startIndex uint64, count uint64) ([]types.Rewards, error) {
	rows, err := a.db.Query(a.getQuery(identityEpochRewardsOldQuery), address, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Rewards
	var item *types.Rewards
	for rows.Next() {
		reward := types.Reward{}
		var epoch uint64
		var prevState, state string
		var age uint16
		if err := rows.Scan(&epoch, &reward.Balance, &reward.Stake, &reward.Type, &prevState, &state, &age); err != nil {
			return nil, err
		}
		if item == nil || item.Epoch != epoch {
			if item != nil {
				res = append(res, *item)
			}
			item = &types.Rewards{
				Epoch:     epoch,
				PrevState: prevState,
				State:     state,
				Age:       age,
			}
		}
		item.Rewards = append(item.Rewards, reward)
	}
	if item != nil {
		res = append(res, *item)
	}
	return res, nil
}

func (a *postgresAccessor) IdentityFlipsCount(address string) (uint64, error) {
	return a.count(identityFlipsCountQuery, address)
}

func (a *postgresAccessor) IdentityFlips(address string, count uint64, continuationToken *string) ([]types.FlipSummary, *string, error) {
	return a.flips(identityFlipsQuery, count, continuationToken, address)
}

func (a *postgresAccessor) IdentityFlipsOld(address string, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	return a.flipsOld(identityFlipsOldQuery, address, startIndex, count)
}
