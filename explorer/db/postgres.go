package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"math/big"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
	log     log.Logger
}

const (
	epochsQuery                          = "epochs.sql"
	epochQuery                           = "epoch.sql"
	epochBlocksQuery                     = "epochBlocks.sql"
	epochTxsQuery                        = "epochTxs.sql"
	blockQuery                           = "block.sql"
	blockTxsQuery                        = "blockTxs.sql"
	epochFlipsWithKeyQuery               = "epochFlipsWithKey.sql"
	epochFlipsQuery                      = "epochFlips.sql"
	epochInvitesQuery                    = "epochInvites.sql"
	epochIdentitiesQuery                 = "epochIdentities.sql"
	flipQuery                            = "flip.sql"
	flipAnswersQuery                     = "flipAnswers.sql"
	identityQuery                        = "identity.sql"
	identityAnswersQuery                 = "identityAnswers.sql"
	identityFlipsQuery                   = "identityFlips.sql"
	identityEpochsQuery                  = "identityEpochs.sql"
	identityCurrentFlipsQuery            = "identityCurrentFlips.sql"
	epochIdentityQuery                   = "epochIdentity.sql"
	epochIdentityFlipsToSolveQuery       = "epochIdentityFlipsToSolve.sql"
	epochIdentityAnswersQuery            = "epochIdentityAnswers.sql"
	identityStatesSummaryQuery           = "identityStatesSummary.sql"
	latestValidationSummaryQuery         = "latestValidationSummary.sql"
	nextValidationSummaryQuery           = "nextValidationSummary.sql"
	identityTxsQuery                     = "identityTxs.sql"
	identityInvitesQuery                 = "identityInvites.sql"
	addressQuery                         = "address.sql"
	transactionQuery                     = "transaction.sql"
	currentEpochQuery                    = "currentEpoch.sql"
	validationIdentityStatesSummaryQuery = "validationIdentityStatesSummary.sql"
)

type flipWithKey struct {
	cid string
	key string
}

var NoDataFound = errors.New("no data found")

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) Summary() (types.Summary, error) {
	summary := types.Summary{}
	var err error
	var epoch int64
	if epoch, err = a.getCurrentEpoch(); err != nil {
		return types.Summary{}, err
	}
	if summary.Identities, err = a.identitiesSummary(identityStatesSummaryQuery); err != nil {
		return types.Summary{}, err
	}
	if epoch > 0 {
		if summary.LatestValidation, err = a.validationSummary(uint64(epoch) - 1); err != nil {
			return types.Summary{}, err
		}
	}
	if summary.NextValidation, err = a.nextValidationSummary(uint64(epoch)); err != nil {
		return types.Summary{}, err
	}
	return summary, nil
}

func (a *postgresAccessor) getCurrentEpoch() (int64, error) {
	var epoch int64
	err := a.db.QueryRow(a.getQuery(currentEpochQuery)).Scan(&epoch)
	if err != nil {
		return 0, err
	}
	return epoch, nil
}

func (a *postgresAccessor) identitiesSummary(queryName string, args ...interface{}) (types.IdentitiesSummary, error) {
	rows, err := a.db.Query(a.getQuery(queryName), args...)
	if err != nil {
		return types.IdentitiesSummary{}, err
	}
	defer rows.Close()
	var states []types.StateCount
	for rows.Next() {
		item := types.StateCount{}
		err = rows.Scan(&item.State, &item.Count)
		if err != nil {
			return types.IdentitiesSummary{}, err
		}
		states = append(states, item)
	}
	return types.IdentitiesSummary{
		States: states,
	}, nil
}

func (a *postgresAccessor) validationSummary(epoch uint64) (res types.CompletedValidationSummary, err error) {
	rows, err := a.db.Query(a.getQuery(latestValidationSummaryQuery), epoch)
	if err != nil {
		return
	}
	if !rows.Next() {
		return
	}
	var validationTime int64
	if err = rows.Scan(&validationTime, &res.FirstBlockHeight, &res.SubmittedFlips, &res.SolvedFlips,
		&res.QualifiedFlips, &res.WeaklyQualifiedFlips, &res.NotQualifiedFlips, &res.InappropriateFlips); err != nil {
		return
	}
	res.Time = common.TimestampToTime(big.NewInt(validationTime))
	if res.Identities, err = a.identitiesSummary(validationIdentityStatesSummaryQuery, epoch); err != nil {
		return
	}
	return
}

func (a *postgresAccessor) nextValidationSummary(epoch uint64) (res types.NewValidationSummary, err error) {
	rows, err := a.db.Query(a.getQuery(nextValidationSummaryQuery), epoch)
	if err != nil {
		return
	}
	if !rows.Next() {
		return
	}

	var validationTime int64
	if err = rows.Scan(&validationTime, &res.Invites, &res.Flips); err != nil {
		return
	}
	res.Time = common.TimestampToTime(big.NewInt(validationTime))
	return
}

func (a *postgresAccessor) Epochs() ([]types.EpochSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var epochs []types.EpochSummary
	for rows.Next() {
		epoch := types.EpochSummary{}
		err = rows.Scan(&epoch.Epoch, &epoch.VerifiedCount, &epoch.BlockCount, &epoch.FlipCount)
		if err != nil {
			return nil, err
		}
		epochs = append(epochs, epoch)
	}
	return epochs, nil
}

func (a *postgresAccessor) Epoch(epoch uint64) (types.EpochDetail, error) {
	epochInfo := types.EpochDetail{}
	err := a.db.QueryRow(a.getQuery(epochQuery), epoch).Scan(&epochInfo.Epoch, &epochInfo.BlockCount)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochDetail{}, err
	}

	if epoch > 0 {
		if epochInfo.LatestValidation, err = a.validationSummary(epoch - 1); err != nil {
			return types.EpochDetail{}, err
		}
	}

	if epochInfo.NextValidation, err = a.nextValidationSummary(epoch); err != nil {
		return types.EpochDetail{}, err
	}

	return epochInfo, nil
}

func (a *postgresAccessor) EpochBlocks(epoch uint64) ([]types.BlockSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochBlocksQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []types.BlockSummary
	for rows.Next() {
		block := types.BlockSummary{}
		var timestamp int64
		err = rows.Scan(&block.Height, &timestamp, &block.TxCount)
		if err != nil {
			return nil, err
		}
		block.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (a *postgresAccessor) EpochTxs(epoch uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochTxsQuery), epoch)
	if err != nil {
		return nil, err
	}
	return a.readTxs(rows)
}

func (a *postgresAccessor) Block(height uint64) (types.BlockDetail, error) {
	res := types.BlockDetail{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(blockQuery), height).Scan(&res.Height, &timestamp, &res.TxCount, &res.Proposer)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.BlockDetail{}, err
	}
	res.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
	return res, nil
}

func (a *postgresAccessor) BlockTxs(height uint64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(blockTxsQuery), height)
	if err != nil {
		return nil, err
	}
	return a.readTxs(rows)
}

func (a *postgresAccessor) readTxs(rows *sql.Rows) ([]types.TransactionSummary, error) {
	defer rows.Close()
	var txs []types.TransactionSummary
	for rows.Next() {
		tx := types.TransactionSummary{}
		var timestamp int64
		if err := rows.Scan(&tx.Hash, &tx.Type, &timestamp, &tx.From, &tx.To, &tx.Amount, &tx.Fee); err != nil {
			return nil, err
		}
		tx.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
		txs = append(txs, tx)
	}
	return txs, nil
}

func (a *postgresAccessor) epochFlipsWithKey(epoch uint64) ([]flipWithKey, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsWithKeyQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []flipWithKey
	for rows.Next() {
		item := flipWithKey{}
		err = rows.Scan(&item.cid, &item.key)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochFlips(epoch uint64) ([]types.FlipSummary, error) {
	rows, err := a.db.Query(a.getQuery(epochFlipsQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.FlipSummary
	for rows.Next() {
		item := types.FlipSummary{}
		err = rows.Scan(&item.Cid, &item.Author, &item.Status, &item.ShortRespCount, &item.LongRespCount)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochInvites(epoch uint64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(epochInvitesQuery), epoch)
	if err != nil {
		return nil, err
	}
	return a.readInvites(rows)
}

func (a *postgresAccessor) readInvites(rows *sql.Rows) ([]types.Invite, error) {
	defer rows.Close()
	var res []types.Invite
	for rows.Next() {
		item := types.Invite{}
		// todo status (Not activated/Candidate)
		if err := rows.Scan(&item.Id, &item.Author); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) EpochIdentities(epoch uint64) ([]types.EpochIdentitySummary, error) {
	rows, err := a.db.Query(a.getQuery(epochIdentitiesQuery), epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.EpochIdentitySummary
	for rows.Next() {
		item := types.EpochIdentitySummary{}
		err = rows.Scan(&item.Address, &item.State, &item.PrevState, &item.Approved, &item.Missed, &item.RespScore, &item.AuthorScore)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) Flip(hash string) (types.Flip, error) {
	flip := types.Flip{}
	var id uint64
	err := a.db.QueryRow(a.getQuery(flipQuery), hash).Scan(&id, &flip.Answer, &flip.Status, &flip.Data)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Flip{}, err
	}

	answerRows, err := a.db.Query(a.getQuery(flipAnswersQuery), id)
	if err != nil {
		return types.Flip{}, err
	}
	defer answerRows.Close()

	for answerRows.Next() {
		item := types.Answer{}
		var isShort bool
		err = answerRows.Scan(&item.Address, &item.Answer, &isShort)
		if err != nil {
			return types.Flip{}, err
		}
		if isShort {
			flip.ShortAnswers = append(flip.ShortAnswers, item)
		} else {
			flip.LongAnswers = append(flip.LongAnswers, item)
		}
	}
	return flip, nil
}

func (a *postgresAccessor) Identity(address string) (types.Identity, error) {
	identity := types.Identity{}
	var addressId int64
	err := a.db.QueryRow(a.getQuery(identityQuery), address).Scan(&addressId, &identity.State)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Identity{}, err
	}
	identity.Address = address

	if identity.ShortAnswers, identity.LongAnswers, err = a.identityAnswerSummary(addressId); err != nil {
		return types.Identity{}, err
	}

	if identity.SubmittedFlipCount, identity.QualifiedFlipCount, identity.WeaklyQualifiedFlipCount,
		identity.AuthorScore, err = a.identityFlipsSummary(addressId); err != nil {
		return types.Identity{}, err
	}

	if identity.Epochs, err = a.identityEpochs(addressId); err != nil {
		return types.Identity{}, err
	}

	if identity.Txs, err = a.identityTxs(addressId); err != nil {
		return types.Identity{}, err
	}

	if identity.CurrentFlipCids, err = a.identityCurrentFlipCids(addressId); err != nil {
		return types.Identity{}, err
	}

	if identity.Invites, err = a.identityInvites(addressId); err != nil {
		return types.Identity{}, err
	}

	return identity, nil
}

func (a *postgresAccessor) identityAnswerSummary(addressId int64) (short, long types.IdentityAnswerSummary, err error) {
	rows, err := a.db.Query(a.getQuery(identityAnswersQuery), addressId)
	if err != nil {
		return types.IdentityAnswerSummary{}, types.IdentityAnswerSummary{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return types.IdentityAnswerSummary{}, types.IdentityAnswerSummary{}, nil
	}
	err = rows.Scan(&short.RightAnswerCount, &short.AnswerCount, &short.RespScore, &long.RightAnswerCount, &long.AnswerCount, &long.RespScore)
	if err != nil {
		return types.IdentityAnswerSummary{}, types.IdentityAnswerSummary{}, err
	}
	return short, long, nil
}

func (a *postgresAccessor) identityFlipsSummary(identityId int64) (count, qualifiedCount, weaklyQualifiedCount uint32, score float32, err error) {
	rows, err := a.db.Query(a.getQuery(identityFlipsQuery), identityId)
	if err != nil {
		return
	}
	if !rows.Next() {
		return
	}
	if err = rows.Scan(&count, &qualifiedCount, &weaklyQualifiedCount, &score); err != nil {
		return
	}
	return
}

func (a *postgresAccessor) identityEpochs(identityId int64) ([]types.IdentityEpoch, error) {
	rows, err := a.db.Query(a.getQuery(identityEpochsQuery), identityId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []types.IdentityEpoch
	for rows.Next() {
		item := types.IdentityEpoch{}
		err = rows.Scan(&item.Epoch, &item.State, &item.Approved, &item.Missed, &item.RespScore, &item.AuthorScore)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) identityTxs(addressId int64) ([]types.TransactionSummary, error) {
	rows, err := a.db.Query(a.getQuery(identityTxsQuery), addressId)
	if err != nil {
		return nil, err
	}
	return a.readTxs(rows)
}

func (a *postgresAccessor) identityCurrentFlipCids(identityId int64) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(identityCurrentFlipsQuery), identityId)
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

func (a *postgresAccessor) identityInvites(addressId int64) ([]types.Invite, error) {
	rows, err := a.db.Query(a.getQuery(identityInvitesQuery), addressId)
	if err != nil {
		return nil, err
	}
	return a.readInvites(rows)
}

func (a *postgresAccessor) EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error) {
	var id int64
	err := a.db.QueryRow(a.getQuery(epochIdentityQuery), epoch, address).Scan(&id)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.EpochIdentity{}, err
	}
	epochIdentity := types.EpochIdentity{}
	if epochIdentity.ShortFlipsToSolve, epochIdentity.LongFlipsToSolve, err = a.epochIdentityFlipsToSolve(id); err != nil {
		return types.EpochIdentity{}, err
	}

	if epochIdentity.ShortAnswers, epochIdentity.LongAnswers, err = a.epochIdentityAnswers(id); err != nil {
		return types.EpochIdentity{}, err
	}

	return epochIdentity, nil
}

func (a *postgresAccessor) epochIdentityFlipsToSolve(epochIdentityId int64) (short, long []string, err error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityFlipsToSolveQuery), epochIdentityId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var isShort bool
		var item string
		err = rows.Scan(&item, &isShort)
		if err != nil {
			return nil, nil, err
		}
		if isShort {
			short = append(short, item)
		} else {
			long = append(long, item)
		}
	}
	return
}

func (a *postgresAccessor) epochIdentityAnswers(epochIdentityId int64) (short, long []types.Answer, err error) {
	rows, err := a.db.Query(a.getQuery(epochIdentityAnswersQuery), epochIdentityId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		item := types.Answer{}
		var isShort bool
		err = rows.Scan(&item.Cid, &item.Answer, &isShort)
		if err != nil {
			return nil, nil, err
		}
		if isShort {
			short = append(short, item)
		} else {
			long = append(long, item)
		}
	}
	return
}

func (a *postgresAccessor) Address(address string) (types.Address, error) {
	res := types.Address{}
	err := a.db.QueryRow(a.getQuery(addressQuery), address).Scan(&res.Address, &res.Balance, &res.Stake, &res.TxCount)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.Address{}, err
	}
	return res, nil
}

func (a *postgresAccessor) Transaction(hash string) (types.TransactionDetail, error) {
	res := types.TransactionDetail{}
	var timestamp int64
	err := a.db.QueryRow(a.getQuery(transactionQuery), hash).Scan(&res.Epoch, &res.BlockHeight, &res.BlockHash,
		&res.Hash, &res.Type, &timestamp, &res.From, &res.To, &res.Amount, &res.Fee)
	if err == sql.ErrNoRows {
		err = NoDataFound
	}
	if err != nil {
		return types.TransactionDetail{}, err
	}
	res.Timestamp = common.TimestampToTime(big.NewInt(timestamp))
	return res, nil
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		a.log.Error("Unable to close db: %v", err)
	}
}
