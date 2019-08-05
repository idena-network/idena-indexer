package postgres

import (
	"database/sql"
	"github.com/idena-network/idena-indexer/explorer/types"
)

const (
	epochIdentityQuery             = "epochIdentity.sql"
	epochIdentityAnswersQuery      = "epochIdentityAnswers.sql"
	epochIdentityFlipsToSolveQuery = "epochIdentityFlipsToSolve.sql"
	epochIdentityFlipsQuery        = "epochIdentityFlips.sql"
)

func (a *postgresAccessor) EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	res := types.EpochIdentity{}
	err := a.db.QueryRow(a.getQuery(epochIdentityQuery), epoch, address).Scan(&res.State,
		&res.ShortAnswers.Point,
		&res.ShortAnswers.FlipsCount,
		&res.LongAnswers.Point,
		&res.LongAnswers.FlipsCount,
		&res.Approved,
		&res.Missed)
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
	defer rows.Close()
	var res []types.Answer
	for rows.Next() {
		item := types.Answer{}
		err = rows.Scan(&item.Cid, &item.RespAnswer, &item.FlipAnswer)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityFlipsQuery), epoch, address)
	if err != nil {
		return nil, err
	}
	return a.readFlips(rows)
}
