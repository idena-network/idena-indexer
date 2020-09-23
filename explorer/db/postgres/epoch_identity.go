package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	epochIdentityQuery                    = "epochIdentity.sql"
	epochIdentityAnswersQuery             = "epochIdentityAnswers.sql"
	epochIdentityFlipsToSolveQuery        = "epochIdentityFlipsToSolve.sql"
	epochIdentityFlipsQuery               = "epochIdentityFlips.sql"
	epochIdentityRewardedFlipsQuery       = "epochIdentityRewardedFlips.sql"
	epochIdentityReportedFlipRewardsQuery = "epochIdentityReportedFlipRewards.sql"
	epochIdentityValidationTxsQuery       = "epochIdentityValidationTxs.sql"
	epochIdentityRewardsQuery             = "epochIdentityRewards.sql"
	epochIdentityBadAuthorQuery           = "epochIdentityBadAuthor.sql"
	epochIdentityRewardedInvitesQuery     = "epochIdentityRewardedInvites.sql"
	epochIdentitySavedInviteRewardsQuery  = "epochIdentitySavedInviteRewards.sql"
	epochIdentityAvailableInvitesQuery    = "epochIdentityAvailableInvites.sql"
)

func (a *postgresAccessor) EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	res := types.EpochIdentity{}
	err := a.db.QueryRow(a.getQuery(epochIdentityQuery), epoch, address).Scan(
		&res.State,
		&res.PrevState,
		&res.ShortAnswers.Point,
		&res.ShortAnswers.FlipsCount,
		&res.TotalShortAnswers.Point,
		&res.TotalShortAnswers.FlipsCount,
		&res.LongAnswers.Point,
		&res.LongAnswers.FlipsCount,
		&res.Approved,
		&res.Missed,
		&res.RequiredFlips,
		&res.MadeFlips,
		&res.AvailableFlips,
		&res.TotalValidationReward,
		&res.BirthEpoch,
		&res.ShortAnswersCount,
		&res.LongAnswersCount,
	)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochIdentity{}, err
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentityShortFlipsToSolve(epoch uint64, address string) ([]string, error) {
	return a.epochIdentityFlipsToSolve(epoch, address, true)
}

func (a *postgresAccessor) EpochIdentityLongFlipsToSolve(epoch uint64, address string) ([]string, error) {
	return a.epochIdentityFlipsToSolve(epoch, address, false)
}

func (a *postgresAccessor) epochIdentityFlipsToSolve(epoch uint64, address string, isShort bool) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityFlipsToSolveQuery), epoch, address, isShort)
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

func (a *postgresAccessor) EpochIdentityShortAnswers(epoch uint64, address string) ([]types.Answer, error) {
	return a.epochIdentityAnswers(epoch, address, true)
}

func (a *postgresAccessor) EpochIdentityLongAnswers(epoch uint64, address string) ([]types.Answer, error) {
	return a.epochIdentityAnswers(epoch, address, false)
}

func (a *postgresAccessor) epochIdentityAnswers(epoch uint64, address string, isShort bool) ([]types.Answer, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityAnswersQuery), epoch, address, isShort)
	if err != nil {
		return nil, err
	}
	return readAnswers(rows)
}

func (a *postgresAccessor) EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error) {
	return a.flipsOld(epochIdentityFlipsQuery, epoch, address)
}

func (a *postgresAccessor) EpochIdentityFlipsWithRewardFlag(epoch uint64, address string) ([]types.FlipWithRewardFlag, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityRewardedFlipsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.FlipWithRewardFlag
	for rows.Next() {
		item := types.FlipWithRewardFlag{}
		var timestamp int64
		words := types.FlipWords{}
		err := rows.Scan(
			&item.Cid,
			&item.Size,
			&item.Author,
			&item.Epoch,
			&item.Status,
			&item.Answer,
			&item.WrongWords,
			&item.WrongWordsVotes,
			&item.ShortRespCount,
			&item.LongRespCount,
			&timestamp,
			&item.Icon,
			&words.Word1.Index,
			&words.Word1.Name,
			&words.Word1.Desc,
			&words.Word2.Index,
			&words.Word2.Name,
			&words.Word2.Desc,
			&item.WithPrivatePart,
			&item.Grade,
			&item.Rewarded,
		)
		if err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		if !words.IsEmpty() {
			item.Words = &words
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentityReportedFlipRewards(epoch uint64, address string) ([]types.ReportedFlipReward, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityReportedFlipRewardsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.ReportedFlipReward
	for rows.Next() {
		item := types.ReportedFlipReward{}
		err := rows.Scan(
			&item.Cid,
			&item.Balance,
			&item.Stake,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentityValidationTxs(epoch uint64, address string) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityValidationTxsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	return readTxsOld(rows)
}

func (a *postgresAccessor) EpochIdentityRewards(epoch uint64, address string) ([]types.Reward, error) {
	return a.rewardsOld(epochIdentityRewardsQuery, epoch, address)
}

func (a *postgresAccessor) EpochIdentityBadAuthor(epoch uint64, address string) (*types.BadAuthor, error) {
	res := types.BadAuthor{}
	err := a.db.QueryRow(a.getQuery(epochIdentityBadAuthorQuery), epoch, address).Scan(
		&res.Address,
		&res.Epoch,
		&res.WrongWords,
		&res.Reason,
		&res.PrevState,
		&res.State,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (a *postgresAccessor) EpochIdentityInvitesWithRewardFlag(epoch uint64, address string) ([]types.InviteWithRewardFlag, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityRewardedInvitesQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.InviteWithRewardFlag
	for rows.Next() {
		item := types.InviteWithRewardFlag{}
		var timestamp, activationTimestamp, killInviteeTimestamp int64
		if err := rows.Scan(
			&item.Hash,
			&item.Author,
			&timestamp,
			&item.Epoch,
			&item.ActivationHash,
			&item.ActivationAuthor,
			&activationTimestamp,
			&item.State,
			&item.KillInviteeHash,
			&killInviteeTimestamp,
			&item.KillInviteeEpoch,
			&item.RewardType,
		); err != nil {
			return nil, err
		}
		item.Timestamp = timestampToTimeUTC(timestamp)
		if activationTimestamp > 0 {
			t := timestampToTimeUTC(activationTimestamp)
			item.ActivationTimestamp = &t
		}
		if killInviteeTimestamp > 0 {
			t := timestampToTimeUTC(killInviteeTimestamp)
			item.KillInviteeTimestamp = &t
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentitySavedInviteRewards(epoch uint64, address string) ([]types.StrValueCount, error) {
	return a.strValueCounts(epochIdentitySavedInviteRewardsQuery, epoch, address)
}

func (a *postgresAccessor) EpochIdentityAvailableInvites(epoch uint64, address string) ([]types.EpochInvites, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityAvailableInvitesQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.EpochInvites
	for rows.Next() {
		item := types.EpochInvites{}
		if err := rows.Scan(&item.Epoch, &item.Invites); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}
