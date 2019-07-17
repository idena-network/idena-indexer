package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
}

const (
	initQuery                = "init.sql"
	maxHeightQuery           = "maxHeight.sql"
	updateFlipsQuery         = "updateFlips.sql"
	insertAnswersQuery       = "insertAnswers.sql"
	insertBlockQuery         = "insertBlock.sql"
	selectIdentityQuery      = "selectIdentity.sql"
	selectFlipQuery          = "selectFlip.sql"
	insertIdentityQuery      = "insertIdentity.sql"
	insertEpochIdentityQuery = "insertEpochIdentity.sql"
	insertTransactionQuery   = "insertTransaction.sql"
	insertSubmittedFlipQuery = "insertSubmittedFlip.sql"
	insertFlipKeyQuery       = "insertFlipKey.sql"
	selectEpochQuery         = "selectEpoch.sql"
	insertEpochQuery         = "insertEpoch.sql"
	insertFlipsToSolveQuery  = "insertFlipsToSolve.sql"
)

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) getIdentityId(tx *sql.Tx, address string) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(selectIdentityQuery), address).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *postgresAccessor) getFlipId(tx *sql.Tx, cid string) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(selectFlipQuery), cid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *postgresAccessor) GetLastHeight() (uint64, error) {
	var maxHeight int64
	err := a.db.QueryRow(a.getQuery(maxHeightQuery)).Scan(&maxHeight)
	if err != nil {
		return 0, err
	}
	return uint64(maxHeight), nil
}

func (a *postgresAccessor) Save(data *Data) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx := newContext(a, tx)

	ctx.epochId, err = a.saveEpoch(ctx, data.Epoch)
	if err != nil {
		return err
	}

	_, ctx.txIdsPerHash, err = a.saveBlock(ctx, data.Block)
	if err != nil {
		return err
	}

	if _, err := a.saveSubmittedFlips(ctx, data.SubmittedFlips); err != nil {
		return err
	}

	if err := a.saveFlipKeys(ctx, data.FlipKeys); err != nil {
		return err
	}

	if err = a.saveIdentities(ctx, data.Identities); err != nil {
		return err
	}

	if err = a.saveFlipsStats(ctx, data.FlipStats); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *postgresAccessor) saveFlipsStats(ctx *context, flipsStats []FlipStats) error {
	for _, flipStats := range flipsStats {
		if err := a.saveFlipStats(ctx, flipStats); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipStats(ctx *context, flipStats FlipStats) error {
	if err := a.saveAnswers(ctx, flipStats.Cid, flipStats.ShortAnswers, true); err != nil {
		return err
	}
	if err := a.saveAnswers(ctx, flipStats.Cid, flipStats.LongAnswers, false); err != nil {
		return err
	}
	flipId, err := ctx.flipId(flipStats.Cid)
	if err != nil {
		return err
	}
	res, err := ctx.tx.Exec(a.getQuery(updateFlipsQuery), flipStats.Status, flipStats.Answer, flipId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("Wrong flips number for saving status and answer")
	}
	return nil
}

func (a *postgresAccessor) saveAnswers(ctx *context, cid string, answers []Answer,
	isShort bool) error {
	for _, answer := range answers {
		if _, err := a.saveAnswer(ctx, cid, answer, isShort); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveAnswer(ctx *context, cid string, answer Answer,
	isShort bool) (int64, error) {
	var id int64
	flipId, err := ctx.flipId(cid)
	if err != nil {
		return 0, err
	}
	epochIdentityId, err := ctx.epochIdentityId(answer.Address)
	if err != nil {
		return 0, err
	}
	err = ctx.tx.QueryRow(a.getQuery(insertAnswersQuery), flipId, epochIdentityId, isShort, answer.Answer).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveEpoch(ctx *context, epoch uint64) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(selectEpochQuery), epoch).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = ctx.tx.QueryRow(a.getQuery(insertEpochQuery), epoch).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveBlock(ctx *context, block Block) (id int64, txIdsPerHash map[string]int64,
	err error) {

	err = ctx.tx.QueryRow(a.getQuery(insertBlockQuery), block.Height, block.Hash, ctx.epochId, block.Time.Int64()).Scan(&id)
	if err != nil {
		return
	}
	txIdsPerHash, err = a.saveTransactions(ctx, id, block.Transactions)
	return
}

func (a *postgresAccessor) saveIdentities(ctx *context, identities []EpochIdentity) error {
	if len(identities) == 0 {
		return nil
	}
	for _, identity := range identities {
		identityId, err := a.saveIdentity(ctx, identity)
		if err != nil {
			return err
		}
		if _, err = a.saveEpochIdentity(ctx, identityId, identity); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveIdentity(ctx *context, identity EpochIdentity) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(selectIdentityQuery), identity.Address).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = ctx.tx.QueryRow(a.getQuery(insertIdentityQuery), identity.Address).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveEpochIdentity(ctx *context, identityId int64, identity EpochIdentity) (int64, error) {
	var id int64

	if err := ctx.tx.QueryRow(a.getQuery(insertEpochIdentityQuery), ctx.epochId, identityId, identity.State, identity.ShortPoint,
		identity.ShortFlips, identity.LongPoint, identity.LongFlips, identity.Approved, identity.Missed).Scan(&id); err != nil {
		return 0, err
	}

	if ctx.epochIdentityIdsPerIdentityId == nil {
		ctx.epochIdentityIdsPerIdentityId = make(map[int64]int64)
	}
	ctx.epochIdentityIdsPerIdentityId[identityId] = id

	if err := a.saveFlipsToSolve(ctx, id, identity.ShortFlipCidsToSolve, true); err != nil {
		return 0, err
	}

	if err := a.saveFlipsToSolve(ctx, id, identity.LongFlipCidsToSolve, false); err != nil {
		return 0, err
	}

	return id, nil
}

func (a *postgresAccessor) saveFlipsToSolve(ctx *context, epochIdentityId int64, cids []string, isShort bool) error {
	for _, cid := range cids {
		if _, err := a.saveFlipToSolve(ctx, epochIdentityId, cid, isShort); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipToSolve(ctx *context, epochIdentityId int64, cid string, isShort bool) (int64, error) {
	flipId, err := ctx.flipId(cid)
	if err != nil {
		return 0, err
	}
	var id int64
	err = ctx.tx.QueryRow(a.getQuery(insertFlipsToSolveQuery), epochIdentityId, flipId, isShort).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveTransactions(ctx *context, blockId int64, txs []Transaction) (map[string]int64, error) {
	if len(txs) == 0 {
		return nil, nil
	}
	txIdsPerHash := make(map[string]int64)
	for _, idenaTx := range txs {
		id, err := a.saveTransaction(ctx, blockId, idenaTx)
		if err != nil {
			return nil, err
		}
		txIdsPerHash[idenaTx.Hash] = id
	}
	return txIdsPerHash, nil
}

func (a *postgresAccessor) saveTransaction(ctx *context, blockId int64, idenaTx Transaction) (int64, error) {
	var id int64
	var to interface{}
	if len(idenaTx.To) > 0 {
		to = idenaTx.To
	} else {
		to = nil
	}
	err := ctx.tx.QueryRow(a.getQuery(insertTransactionQuery), idenaTx.Hash, blockId, idenaTx.Type, idenaTx.From, to,
		idenaTx.Amount.Int64(), idenaTx.Fee.Int64()).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveSubmittedFlips(ctx *context, flips []Flip) (map[string]int64, error) {
	if len(flips) == 0 {
		return nil, nil
	}
	flipIdsPerCid := make(map[string]int64)
	for _, flip := range flips {
		id, err := a.saveSubmittedFlip(ctx, ctx.txId(flip.TxHash), flip)
		if err != nil {
			return nil, err
		}
		flipIdsPerCid[flip.Cid] = id
	}
	return flipIdsPerCid, nil
}

func (a *postgresAccessor) saveSubmittedFlip(ctx *context, txId int64, flip Flip) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId).Scan(&id)
	return id, errors.Wrapf(err, "unable to save flip %s", flip.Cid)
}

func (a *postgresAccessor) saveFlipKeys(ctx *context, keys []FlipKey) error {
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		err := a.saveFlipKey(ctx, ctx.txId(key.TxHash), key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipKey(ctx *context, txId int64, key FlipKey) error {
	_, err := ctx.tx.Exec(a.getQuery(insertFlipKeyQuery), txId, key.Key)
	return errors.Wrapf(err, "unable to save flip key %s", key.Key)
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}
