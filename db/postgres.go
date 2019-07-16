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

	ctx := newContext(a)

	epochId, err := a.saveEpoch(tx, data.Epoch)
	if err != nil {
		return err
	}

	_, ctx.txIdsPerHash, err = a.saveBlock(tx, epochId, data.Block)
	if err != nil {
		return err
	}

	if _, err := a.saveSubmittedFlips(tx, ctx, data.SubmittedFlips); err != nil {
		return err
	}

	if err := a.saveFlipKeys(tx, ctx, data.FlipKeys); err != nil {
		return err
	}

	if err = a.saveIdentities(tx, epochId, data.Identities); err != nil {
		return err
	}

	if err = a.saveFlipsStats(tx, ctx, data.FlipStats); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *postgresAccessor) saveFlipsStats(tx *sql.Tx, ctx *context, flipsStats []FlipStats) error {
	for _, flipStats := range flipsStats {
		if err := a.saveFlipStats(tx, ctx, flipStats); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipStats(tx *sql.Tx, ctx *context, flipStats FlipStats) error {
	if err := a.saveAnswers(tx, ctx, flipStats.Cid, flipStats.ShortAnswers, true); err != nil {
		return err
	}
	if err := a.saveAnswers(tx, ctx, flipStats.Cid, flipStats.LongAnswers, false); err != nil {
		return err
	}
	flipId, err := ctx.flipId(tx, flipStats.Cid)
	if err != nil {
		return err
	}
	res, err := tx.Exec(a.getQuery(updateFlipsQuery), flipStats.Status, flipStats.Answer, flipId)
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

func (a *postgresAccessor) saveAnswers(tx *sql.Tx, ctx *context, cid string, answers []Answer,
	isShort bool) error {
	for _, answer := range answers {
		if _, err := a.saveAnswer(tx, ctx, cid, answer, isShort); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveAnswer(tx *sql.Tx, ctx *context, cid string, answer Answer,
	isShort bool) (int64, error) {
	var id int64
	flipId, err := ctx.flipId(tx, cid)
	if err != nil {
		return 0, err
	}
	identityId, err := ctx.identityId(tx, answer.Address)
	if err != nil {
		return 0, err
	}
	err = tx.QueryRow(a.getQuery(insertAnswersQuery), flipId, identityId, isShort, answer.Answer).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveEpoch(tx *sql.Tx, epoch uint64) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(selectEpochQuery), epoch).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = tx.QueryRow(a.getQuery(insertEpochQuery), epoch).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveBlock(tx *sql.Tx, epochId int64, block Block) (id int64, txIdsPerHash map[string]int64,
	err error) {

	err = tx.QueryRow(a.getQuery(insertBlockQuery), block.Height, block.Hash, epochId, block.Time.Int64()).Scan(&id)
	if err != nil {
		return
	}
	txIdsPerHash, err = a.saveTransactions(tx, id, block.Transactions)
	return
}

func (a *postgresAccessor) saveIdentities(tx *sql.Tx, epochId int64, identities []EpochIdentity) error {
	if len(identities) == 0 {
		return nil
	}
	for _, identity := range identities {
		identityId, err := a.saveIdentity(tx, identity)
		if err != nil {
			return err
		}
		if _, err = a.saveEpochIdentity(tx, epochId, identityId, identity); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveIdentity(tx *sql.Tx, identity EpochIdentity) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(selectIdentityQuery), identity.Address).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = tx.QueryRow(a.getQuery(insertIdentityQuery), identity.Address).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveEpochIdentity(tx *sql.Tx, epochId int64, identityId int64, identity EpochIdentity) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(insertEpochIdentityQuery), epochId, identityId, identity.State, identity.ShortPoint,
		identity.ShortFlips, identity.LongPoint, identity.LongFlips).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveTransactions(tx *sql.Tx, blockId int64, txs []Transaction) (map[string]int64, error) {
	if len(txs) == 0 {
		return nil, nil
	}
	txIdsPerHash := make(map[string]int64)
	for _, idenaTx := range txs {
		id, err := a.saveTransaction(tx, blockId, idenaTx)
		if err != nil {
			return nil, err
		}
		txIdsPerHash[idenaTx.Hash] = id
	}
	return txIdsPerHash, nil
}

func (a *postgresAccessor) saveTransaction(tx *sql.Tx, blockId int64, idenaTx Transaction) (int64, error) {
	var id int64
	var to interface{}
	if len(idenaTx.To) > 0 {
		to = idenaTx.To
	} else {
		to = nil
	}
	err := tx.QueryRow(a.getQuery(insertTransactionQuery), idenaTx.Hash, blockId, idenaTx.Type, idenaTx.From, to,
		idenaTx.Amount.Int64(), idenaTx.Fee.Int64()).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveSubmittedFlips(tx *sql.Tx, ctx *context, flips []Flip) (map[string]int64, error) {
	if len(flips) == 0 {
		return nil, nil
	}
	flipIdsPerCid := make(map[string]int64)
	for _, flip := range flips {
		id, err := a.saveSubmittedFlip(tx, ctx.txId(flip.TxHash), flip)
		if err != nil {
			return nil, err
		}
		flipIdsPerCid[flip.Cid] = id
	}
	return flipIdsPerCid, nil
}

func (a *postgresAccessor) saveSubmittedFlip(tx *sql.Tx, txId int64, flip Flip) (int64, error) {
	var id int64
	err := tx.QueryRow(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId).Scan(&id)
	return id, errors.Wrapf(err, "unable to save flip %s", flip.Cid)
}

func (a *postgresAccessor) saveFlipKeys(tx *sql.Tx, ctx *context, keys []FlipKey) error {
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		err := a.saveFlipKey(tx, ctx.txId(key.TxHash), key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipKey(tx *sql.Tx, txId int64, key FlipKey) error {
	_, err := tx.Exec(a.getQuery(insertFlipKeyQuery), txId, key.Key)
	return errors.Wrapf(err, "unable to save flip key %s", key.Key)
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}
