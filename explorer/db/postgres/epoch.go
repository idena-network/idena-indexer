package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	epochQuery                        = "epoch.sql"
	lastEpochQuery                    = "lastEpoch.sql"
	epochBlocksCountQuery             = "epochBlocksCount.sql"
	epochBlocksQuery                  = "epochBlocks.sql"
	epochFlipsCountQuery              = "epochFlipsCount.sql"
	epochFlipsQuery                   = "epochFlips.sql"
	epochFlipStatesQuery              = "epochFlipStates.sql"
	epochFlipQualifiedAnswersQuery    = "epochFlipQualifiedAnswers.sql"
	epochFlipQualifiedWrongWordsQuery = "epochFlipQualifiedWrongWords.sql"
	epochIdentityStatesSummaryQuery   = "epochIdentityStatesSummary.sql"
	epochInviteStatesSummaryQuery     = "epochInviteStatesSummary.sql"
	epochIdentitiesQueryCount         = "epochIdentitiesCount.sql"
	epochIdentitiesQuery              = "epochIdentities.sql"
	epochInvitesCountQuery            = "epochInvitesCount.sql"
	epochInvitesQuery                 = "epochInvites.sql"
	epochInvitesSummaryQuery          = "epochInvitesSummary.sql"
	epochTxsCountQuery                = "epochTxsCount.sql"
	epochTxsQuery                     = "epochTxs.sql"
	epochCoinsQuery                   = "epochCoins.sql"
	epochRewardsSummaryQuery          = "epochRewardsSummary.sql"
	epochBadAuthorsCountQuery         = "epochBadAuthorsCount.sql"
	epochBadAuthorsQuery              = "epochBadAuthors.sql"
	epochGoodAuthorsCountQuery        = "epochGoodAuthorsCount.sql"
	epochGoodAuthorsQuery             = "epochGoodAuthors.sql"
	epochRewardsCountQuery            = "epochRewardsCount.sql"
	epochRewardsQuery                 = "epochRewards.sql"
	epochIdentitiesRewardsCountQuery  = "epochIdentitiesRewardsCount.sql"
	epochIdentitiesRewardsQuery       = "epochIdentitiesRewards.sql"
	epochFundPaymentsQuery            = "epochFundPayments.sql"
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
	err := a.db.QueryRow(a.getQuery(queryName), args...).Scan(&res.Epoch,
		&validationTime,
		&res.ValidationFirstBlockHeight)
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

func (a *postgresAccessor) EpochFlips(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error) {
	return a.flips(epochFlipsQuery, epoch, startIndex, count)
}

func (a *postgresAccessor) EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipQualifiedAnswersQuery, epoch)
}

func (a *postgresAccessor) EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochFlipStatesQuery, epoch)
}

func (a *postgresAccessor) EpochFlipWrongWordsSummary(epoch uint64) ([]types.NullableBoolValueCount, error) {
	return a.nullableBoolValueCounts(epochFlipQualifiedWrongWordsQuery, epoch)
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

func (a *postgresAccessor) EpochIdentities(epoch uint64, prevStates []string, states []string, startIndex uint64, count uint64) ([]types.EpochIdentity, error) {
	prevStateIds, err := convertIdentityStates(prevStates)
	if err != nil {
		return nil, err
	}
	stateIds, err := convertIdentityStates(states)
	if err != nil {
		return nil, err
	}
	rows, err := a.db.Query(a.getQuery(epochIdentitiesQuery),
		epoch,
		pq.Array(prevStateIds),
		pq.Array(stateIds),
		startIndex,
		count,
	)
	if err != nil {
		return nil, err
	}
	return readEpochIdentities(rows)
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

func (a *postgresAccessor) EpochInviteStatesSummary(epoch uint64) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochInviteStatesSummaryQuery, epoch)
}

func (a *postgresAccessor) EpochInvitesCount(epoch uint64) (uint64, error) {
	return a.count(epochInvitesCountQuery, epoch)
}

func (a *postgresAccessor) EpochInvites(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(epochInvitesQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readInvites(rows)
}

func (a *postgresAccessor) EpochTxsCount(epoch uint64) (uint64, error) {
	return a.count(epochTxsCountQuery, epoch)
}

func (a *postgresAccessor) EpochTxs(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochTxsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readTxs(rows)
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

func (a *postgresAccessor) EpochBadAuthors(epoch uint64, startIndex uint64, count uint64) ([]types.BadAuthor, error) {
	rows, err := a.db.Query(a.getQuery(epochBadAuthorsQuery), epoch, startIndex, count)
	if err != nil {
		return nil, err
	}
	return readBadAuthors(rows)
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
