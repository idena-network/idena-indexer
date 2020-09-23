package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"strconv"
)

const (
	epochQuery                             = "epoch.sql"
	lastEpochQuery                         = "lastEpoch.sql"
	epochBlocksCountQuery                  = "epochBlocksCount.sql"
	epochBlocksQuery                       = "epochBlocks.sql"
	epochBlocksOldQuery                    = "epochBlocksOld.sql"
	epochFlipsCountQuery                   = "epochFlipsCount.sql"
	epochFlipsQuery                        = "epochFlips.sql"
	epochFlipsOldQuery                     = "epochFlipsOld.sql"
	epochFlipStatesQuery                   = "epochFlipStates.sql"
	epochFlipQualifiedAnswersQuery         = "epochFlipQualifiedAnswers.sql"
	epochFlipQualifiedWrongWordsQuery      = "epochFlipQualifiedWrongWords.sql"
	epochIdentityStatesSummaryQuery        = "epochIdentityStatesSummary.sql"
	epochIdentityStatesInterimSummaryQuery = "epochIdentityStatesInterimSummary.sql"
	epochInviteStatesSummaryQuery          = "epochInviteStatesSummary.sql"
	epochIdentitiesQueryCount              = "epochIdentitiesCount.sql"
	epochIdentitiesQuery                   = "epochIdentities.sql"
	epochIdentitiesOldQuery                = "epochIdentitiesOld.sql"
	epochInvitesCountQuery                 = "epochInvitesCount.sql"
	epochInvitesQuery                      = "epochInvites.sql"
	epochInvitesOldQuery                   = "epochInvitesOld.sql"
	epochInvitesSummaryQuery               = "epochInvitesSummary.sql"
	epochTxsCountQuery                     = "epochTxsCount.sql"
	epochTxsQuery                          = "epochTxs.sql"
	epochTxsOldQuery                       = "epochTxsOld.sql"
	epochCoinsQuery                        = "epochCoins.sql"
	epochRewardsSummaryQuery               = "epochRewardsSummary.sql"
	epochBadAuthorsCountQuery              = "epochBadAuthorsCount.sql"
	epochBadAuthorsQuery                   = "epochBadAuthors.sql"
	epochBadAuthorsOldQuery                = "epochBadAuthorsOld.sql"
	epochRewardsCountQuery                 = "epochRewardsCount.sql"
	epochRewardsQuery                      = "epochRewards.sql"
	epochIdentitiesRewardsCountQuery       = "epochIdentitiesRewardsCount.sql"
	epochIdentitiesRewardsQuery            = "epochIdentitiesRewards.sql"
	epochIdentitiesRewardsOldQuery         = "epochIdentitiesRewardsOld.sql"
	epochFundPaymentsQuery                 = "epochFundPayments.sql"
)

var identityStatesByName = map[string]uint8{
	"Undefined": 0,
	"Invite":    1,
	"Candidate": 2,
	"Verified":  3,
	"Suspended": 4,
	"Killed":    5,
	"Zombie":    6,
	"Newbie":    7,
	"Human":     8,
}

func convertIdentityStates(names []string) ([]uint8, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var res []uint8
	for _, name := range names {
		if state, ok := identityStatesByName[name]; ok {
			res = append(res, state)
		} else {
			return nil, errors.Errorf("Unknown state %s", name)
		}
	}
	return res, nil
}

func (a *postgresAccessor) LastEpoch() (types.EpochDetail, error) {
	return a.epoch(lastEpochQuery)
}

func (a *postgresAccessor) Epoch(epoch uint64) (types.EpochDetail, error) {
	return a.epoch(epochQuery, epoch)
}

func (a *postgresAccessor) epoch(queryName string, args ...interface{}) (types.EpochDetail, error) {
	res := types.EpochDetail{}
	var validationTime int64
	err := a.db.QueryRow(a.getQuery(queryName), args...).Scan(
		&res.Epoch,
		&validationTime,
		&res.ValidationFirstBlockHeight,
		&res.MinScoreForInvite,
	)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochDetail{}, err
	}
	res.ValidationTime = timestampToTimeUTC(validationTime)
	return res, nil
}

func (a *postgresAccessor) EpochBlocksCount(epoch uint64) (uint64, error) {
	return a.count(epochBlocksCountQuery, epoch)
}

func (a *postgresAccessor) EpochBlocks(epoch uint64, count uint64, continuationToken *string) ([]types.BlockSummary, *string, error) {
	res, nextContinuationToken, err := a.page(epochBlocksQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		defer rows.Close()
		var res []types.BlockSummary
		var height uint64
		for rows.Next() {
			block := types.BlockSummary{
				Coins: types.AllCoins{},
			}
			var timestamp int64
			if err := rows.Scan(&height,
				&block.Hash,
				&timestamp,
				&block.TxCount,
				&block.Proposer,
				&block.ProposerVrfScore,
				&block.IsEmpty,
				&block.BodySize,
				&block.FullSize,
				&block.VrfProposerThreshold,
				&block.FeeRate,
				&block.Coins.Burnt,
				&block.Coins.Minted,
				&block.Coins.TotalBalance,
				&block.Coins.TotalStake,
				pq.Array(&block.Flags)); err != nil {
				return nil, 0, err
			}
			block.Height = height
			block.Timestamp = timestampToTimeUTC(timestamp)
			res = append(res, block)
		}
		return res, height, nil
	}, count, continuationToken, epoch)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.BlockSummary), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochBlocksOld(epoch uint64, startIndex uint64, count uint64) ([]types.BlockSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochBlocksOldQuery), epoch, startIndex, count)
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
			&block.ProposerVrfScore,
			&block.IsEmpty,
			&block.BodySize,
			&block.FullSize,
			&block.VrfProposerThreshold,
			&block.FeeRate,
			&block.Coins.Burnt,
			&block.Coins.Minted,
			&block.Coins.TotalBalance,
			&block.Coins.TotalStake,
			pq.Array(&block.Flags))
		if err != nil {
			return nil, err
		}
		block.Timestamp = timestampToTimeUTC(timestamp)
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (a *postgresAccessor) EpochFlipsCount(epoch uint64) (uint64, error) {
	return a.count(epochFlipsCountQuery, epoch)
}

func (a *postgresAccessor) EpochFlips(epoch uint64, count uint64, continuationToken *string) ([]types.FlipSummary, *string, error) {
	return a.flips(epochFlipsQuery, count, continuationToken, epoch)
}

func (a *postgresAccessor) EpochFlipsOld(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	return a.flipsOld(epochFlipsOldQuery, epoch, startIndex, count)
}

func (a *postgresAccessor) EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipQualifiedAnswersQuery, epoch)
}

func (a *postgresAccessor) EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipStatesQuery, epoch)
}

func (a *postgresAccessor) EpochFlipWrongWordsSummary(epoch uint64) ([]types.NullableBoolValueCount, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipQualifiedWrongWordsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.NullableBoolValueCount
	for rows.Next() {
		item := types.NullableBoolValueCount{}
		nullGrade := sql.NullInt32{}
		if err := rows.Scan(&nullGrade, &item.Count); err != nil {
			return nil, err
		}
		if nullGrade.Valid {
			const gradeReported = 1
			reported := nullGrade.Int32 == int32(gradeReported)
			item.Value = &reported
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentitiesCount(epoch uint64, prevStates []string, states []string) (uint64, error) {
	prevStateIds, err := convertIdentityStates(prevStates)
	if err != nil {
		return 0, err
	}
	stateIds, err := convertIdentityStates(states)
	if err != nil {
		return 0, err
	}
	return a.count(epochIdentitiesQueryCount, epoch, pq.Array(prevStateIds), pq.Array(stateIds))
}

func (a *postgresAccessor) EpochIdentities(epoch uint64, prevStates []string, states []string, count uint64,
	continuationToken *string) ([]types.EpochIdentity, *string, error) {
	prevStateIds, err := convertIdentityStates(prevStates)
	if err != nil {
		return nil, nil, err
	}
	stateIds, err := convertIdentityStates(states)
	if err != nil {
		return nil, nil, err
	}
	res, nextContinuationToken, err := a.page(epochIdentitiesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readEpochIdentities(rows)
	}, count, continuationToken, epoch, pq.Array(prevStateIds), pq.Array(stateIds))
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.EpochIdentity), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochIdentitiesOld(epoch uint64, prevStates []string, states []string, startIndex uint64, count uint64) ([]types.EpochIdentity, error) {
	prevStateIds, err := convertIdentityStates(prevStates)
	if err != nil {
		return nil, err
	}
	stateIds, err := convertIdentityStates(states)
	if err != nil {
		return nil, err
	}
	rows, err := a.db.Query(a.getQuery(epochIdentitiesOldQuery),
		epoch,
		pq.Array(prevStateIds),
		pq.Array(stateIds),
		startIndex,
		count,
	)
	if err != nil {
		return nil, err
	}
	return readEpochIdentitiesOld(rows)
}

func (a *postgresAccessor) EpochIdentityStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochIdentityStatesSummaryQuery, epoch)
}

func (a *postgresAccessor) EpochIdentityStatesInterimSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochIdentityStatesInterimSummaryQuery, epoch)
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

func (a *postgresAccessor) EpochInviteStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochInviteStatesSummaryQuery, epoch)
}

func (a *postgresAccessor) EpochInvitesCount(epoch uint64) (uint64, error) {
	return a.count(epochInvitesCountQuery, epoch)
}

func (a *postgresAccessor) EpochInvites(epoch uint64, count uint64, continuationToken *string) ([]types.Invite, *string, error) {
	res, nextContinuationToken, err := a.page(epochInvitesQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readInvites(rows)
	}, count, continuationToken, epoch)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.Invite), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochInvitesOld(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(epochInvitesOldQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readInvitesOld(rows)
}

func (a *postgresAccessor) EpochTxsCount(epoch uint64) (uint64, error) {
	return a.count(epochTxsCountQuery, epoch)
}

func (a *postgresAccessor) EpochTxs(epoch uint64, count uint64, continuationToken *string) ([]types.TransactionSummary, *string, error) {
	res, nextContinuationToken, err := a.page(epochTxsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readTxs(rows)
	}, count, continuationToken, epoch)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.TransactionSummary), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochTxsOld(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochTxsOldQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readTxsOld(rows)
}

func (a *postgresAccessor) EpochCoins(epoch uint64) (types.AllCoins, error) {
	return a.coins(epochCoinsQuery, epoch)
}

func (a *postgresAccessor) EpochRewardsSummary(epoch uint64) (types.RewardsSummary, error) {
	res := types.RewardsSummary{}
	err := a.db.QueryRow(a.getQuery(epochRewardsSummaryQuery), epoch).
		Scan(
			&res.Epoch,
			&res.Total,
			&res.Validation,
			&res.Flips,
			&res.Invitations,
			&res.FoundationPayouts,
			&res.ZeroWalletFund,
			&res.ValidationShare,
			&res.FlipsShare,
			&res.InvitationsShare,
		)
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

func (a *postgresAccessor) EpochBadAuthors(epoch uint64, count uint64, continuationToken *string) ([]types.BadAuthor, *string, error) {
	res, nextContinuationToken, err := a.page(epochBadAuthorsQuery, func(rows *sql.Rows) (interface{}, uint64, error) {
		return readBadAuthors(rows)
	}, count, continuationToken, epoch)
	if err != nil {
		return nil, nil, err
	}
	return res.([]types.BadAuthor), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochBadAuthorsOld(epoch uint64, startIndex uint64, count uint64) ([]types.BadAuthor, error) {
	rows, err := a.db.Query(a.getQuery(epochBadAuthorsOldQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readBadAuthorsOld(rows)
}

func (a *postgresAccessor) EpochRewardsCount(epoch uint64) (uint64, error) {
	return a.count(epochRewardsCountQuery, epoch)
}

func (a *postgresAccessor) EpochRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Reward, error) {
	return a.rewardsOld(epochRewardsQuery, epoch, startIndex, count)
}

func (a *postgresAccessor) EpochIdentitiesRewardsCount(epoch uint64) (uint64, error) {
	return a.count(epochIdentitiesRewardsCountQuery, epoch)
}

func (a *postgresAccessor) EpochIdentitiesRewards(epoch uint64, count uint64, continuationToken *string) ([]types.Rewards, *string, error) {
	var continuationId *uint64
	var err error
	if continuationId, err = parseUintContinuationToken(continuationToken); err != nil {
		return nil, nil, err
	}
	if continuationId == nil {
		v := uint64(0)
		continuationId = &v
	}
	rows, err := a.db.Query(a.getQuery(epochIdentitiesRewardsQuery), epoch, count+1, continuationId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var res []types.Rewards
	var item *types.Rewards
	for rows.Next() {
		reward := types.Reward{}
		var address, prevState, state string
		var age uint16
		if err := rows.Scan(&address, &reward.Balance, &reward.Stake, &reward.Type, &prevState, &state, &age); err != nil {
			return nil, nil, err
		}
		if item == nil || item.Address != address {
			if item != nil {
				res = append(res, *item)
			}
			item = &types.Rewards{
				Address:   address,
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
	resSlice, nextContinuationToken := getResWithContinuationToken(strconv.FormatUint(*continuationId+count, 10), count, res)
	return resSlice.([]types.Rewards), nextContinuationToken, nil
}

func (a *postgresAccessor) EpochIdentitiesRewardsOld(epoch uint64, startIndex uint64, count uint64) ([]types.Rewards, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentitiesRewardsOldQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.Rewards
	var item *types.Rewards
	for rows.Next() {
		reward := types.Reward{}
		var address, prevState, state string
		var age uint16
		if err := rows.Scan(&address, &reward.Balance, &reward.Stake, &reward.Type, &prevState, &state, &age); err != nil {
			return nil, err
		}
		if item == nil || item.Address != address {
			if item != nil {
				res = append(res, *item)
			}
			item = &types.Rewards{
				Address:   address,
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
