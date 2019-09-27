package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	epochIdentityQuery              = "epochIdentity.sql"
	epochIdentityAnswersQuery       = "epochIdentityAnswers.sql"
	epochIdentityFlipsToSolveQuery  = "epochIdentityFlipsToSolve.sql"
	epochIdentityFlipsQuery         = "epochIdentityFlips.sql"
	epochIdentityValidationTxsQuery = "epochIdentityValidationTxs.sql"
	epochIdentityRewardsCountQuery  = "epochIdentityRewardsCount.sql"
	epochIdentityRewardsQuery       = "epochIdentityRewards.sql"
)

func (a *postgresAccessor) EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	res := types.EpochIdentity{}
	err := a.db.QueryRow(a.getQuery(epochIdentityQuery), epoch, address).Scan(&res.State,
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
		&res.MadeFlips)
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
	return a.readAnswers(rows)
}

func (a *postgresAccessor) EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityFlipsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	return a.readFlips(rows)
}

func (a *postgresAccessor) EpochIdentityValidationTxs(epoch uint64, address string) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityValidationTxsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	return a.readTxs(rows)
}

func (a *postgresAccessor) EpochIdentityRewards(epoch uint64, address string) ([]types.Reward, error) {
	return a.rewards(epochIdentityRewardsQuery, epoch, address)
}
