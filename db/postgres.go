package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

type postgresAccessor struct {
	db      *sql.DB
	queries map[string]string
}

const (
	initQuery                       = "init.sql"
	maxHeightQuery                  = "maxHeight.sql"
	currentFlipCidsQuery            = "currentFlipCids.sql"
	currentFlipCidsWithoutDataQuery = "currentFlipCidsWithoutData.sql"
	updateFlipStateQuery            = "updateFlipState.sql"
	updateFlipDataQuery             = "updateFlipData.sql"
	updateFlipMemPoolDataQuery      = "updateFlipMemPoolData.sql"
	insertAnswersQuery              = "insertAnswers.sql"
	insertBlockQuery                = "insertBlock.sql"
	insertProposerQuery             = "insertProposer.sql"
	selectIdentityQuery             = "selectIdentity.sql"
	selectFlipQuery                 = "selectFlip.sql"
	insertEpochIdentityQuery        = "insertEpochIdentity.sql"
	insertTransactionQuery          = "insertTransaction.sql"
	insertSubmittedFlipQuery        = "insertSubmittedFlip.sql"
	insertFlipKeyQuery              = "insertFlipKey.sql"
	selectEpochQuery                = "selectEpoch.sql"
	insertEpochQuery                = "insertEpoch.sql"
	insertFlipsToSolveQuery         = "insertFlipsToSolve.sql"
	selectAddressQuery              = "selectAddress.sql"
	insertAddressQuery              = "insertAddress.sql"
	insertTemporaryIdentityQuery    = "insertTemporaryIdentity.sql"
	archiveAddressStateQuery        = "archiveAddressState.sql"
	insertAddressStateQuery         = "insertAddressState.sql"
	archiveIdentityStateQuery       = "archiveIdentityState.sql"
	insertIdentityStateQuery        = "insertIdentityState.sql"
	resetToBlockQuery               = "resetToBlock.sql"
	insertBalanceQuery              = "insertBalance.sql"
	insertBlockFlagQuery            = "insertBlockFlag.sql"
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

func (a *postgresAccessor) GetCurrentFlipCids(address string) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(currentFlipCidsQuery), address)
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

func (a *postgresAccessor) GetCurrentFlipsWithoutData(limit uint32) ([]string, error) {
	rows, err := a.db.Query(a.getQuery(currentFlipCidsWithoutDataQuery), limit)
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

func (a *postgresAccessor) ResetTo(height uint64) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := strings.Split(a.getQuery(resetToBlockQuery), ";")
	for _, query := range queries {
		if strings.Contains(query, "$") {
			_, err = tx.Exec(query, height)
		} else {
			_, err = tx.Exec(query)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (a *postgresAccessor) Save(data *Data) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx := newContext(a, tx)

	ctx.epochId, err = a.saveEpoch(ctx, data.Epoch, data.ValidationTime)
	if err != nil {
		return err
	}

	ctx.blockId, err = a.saveBlock(ctx, data.Block)
	if err != nil {
		return err
	}

	if err = a.saveBlockFlags(ctx, data.Block.Flags); err != nil {
		return err
	}

	if ctx.addrIdsPerAddr, err = a.saveAddresses(ctx, data.Addresses); err != nil {
		return err
	}

	if err = a.saveProposer(ctx, data.Block); err != nil {
		return err
	}

	ctx.txIdsPerHash, err = a.saveTransactions(ctx, data.Block.Transactions)
	if err != nil {
		return err
	}

	if err = a.saveAddressStates(ctx, data.Addresses); err != nil {
		return err
	}

	if err := a.saveBalances(ctx, data.Balances); err != nil {
		return err
	}

	if _, err := a.saveSubmittedFlips(ctx, data.SubmittedFlips); err != nil {
		return err
	}

	if err := a.saveFlipKeys(ctx, data.FlipKeys); err != nil {
		return err
	}

	if err := a.saveFlipsData(ctx, data.FlipsData); err != nil {
		return err
	}

	if err = a.saveIdentities(ctx, data.Identities); err != nil {
		return err
	}

	if err = a.saveFlipsStats(ctx, data.FlipStats); err != nil {
		return err
	}

	if err := a.saveFlipsMemPoolData(ctx, data.FlipsMemPoolData); err != nil {
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
	if _, err := ctx.tx.Exec(a.getQuery(updateFlipStateQuery), flipStats.Status, flipStats.Answer, ctx.blockId, flipStats.Cid); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveFlipsData(ctx *context, flipsData []FlipData) error {
	if len(flipsData) == 0 {
		return nil
	}
	for _, flipData := range flipsData {
		if err := a.saveFlipData(ctx, flipData); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipData(ctx *context, flipData FlipData) error {
	txId, err := ctx.txId(flipData.TxHash)
	if err != nil {
		return err
	}
	if _, err := ctx.tx.Exec(a.getQuery(updateFlipDataQuery), flipData.Data, txId, flipData.Cid); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveFlipsMemPoolData(ctx *context, flipsData []FlipData) error {
	if len(flipsData) == 0 {
		return nil
	}
	for _, flipData := range flipsData {
		if err := a.saveFlipMemPoolData(ctx, flipData); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipMemPoolData(ctx *context, flipData FlipData) error {
	_, err := ctx.tx.Exec(a.getQuery(updateFlipMemPoolDataQuery), flipData.Data, flipData.Cid)
	return err
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

func (a *postgresAccessor) saveAnswer(ctx *context, cid string, answer Answer, isShort bool) (int64, error) {
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

func (a *postgresAccessor) saveEpoch(ctx *context, epoch uint64, validationTime big.Int) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(selectEpochQuery), epoch).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = ctx.tx.QueryRow(a.getQuery(insertEpochQuery), epoch, validationTime.Int64()).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveBlock(ctx *context, block Block) (id int64, err error) {
	err = ctx.tx.QueryRow(a.getQuery(insertBlockQuery), block.Height, block.Hash, ctx.epochId, block.Time.Int64()).Scan(&id)
	return
}

func (a *postgresAccessor) saveBlockFlags(ctx *context, flags []string) error {
	if len(flags) == 0 {
		return nil
	}
	for _, flag := range flags {
		if err := a.saveBlockFlag(ctx, flag); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveBlockFlag(ctx *context, flag string) error {
	_, err := ctx.tx.Exec(a.getQuery(insertBlockFlagQuery), ctx.blockId, flag)
	return errors.Wrapf(err, "unable to save block flag")
}

func (a *postgresAccessor) saveProposer(ctx *context, block Block) error {
	if len(block.Proposer) == 0 {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertProposerQuery), ctx.blockId, block.Proposer)
	return err
}

func (a *postgresAccessor) saveAddresses(ctx *context, addresses []Address) (map[string]int64, error) {
	if len(addresses) == 0 {
		return nil, nil
	}
	addrIdsPerAddr := make(map[string]int64)
	for _, address := range addresses {
		addressId, err := a.saveAddress(ctx, address)
		if err != nil {
			return nil, err
		}
		addrIdsPerAddr[address.Address] = addressId
		if address.IsTemporary {
			if err = a.saveTemporaryIdentity(ctx, addressId); err != nil {
				return nil, err
			}
		}
	}
	return addrIdsPerAddr, nil
}

func (a *postgresAccessor) saveAddressStates(ctx *context, addresses []Address) error {
	if len(addresses) == 0 {
		return nil
	}
	for _, address := range addresses {
		if len(address.StateChanges) == 0 {
			continue
		}
		for _, stateChange := range address.StateChanges {
			if _, err := a.saveAddressState(ctx, ctx.addrIdsPerAddr[address.Address], stateChange); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *postgresAccessor) saveAddress(ctx *context, address Address) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(selectAddressQuery), address.Address).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	err = ctx.tx.QueryRow(a.getQuery(insertAddressQuery), address.Address, ctx.blockId).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveTemporaryIdentity(ctx *context, addressId int64) error {
	_, err := ctx.tx.Exec(a.getQuery(insertTemporaryIdentityQuery), addressId, ctx.blockId)
	return errors.Wrapf(err, "unable to save temporary identity")
}

func (a *postgresAccessor) saveAddressState(ctx *context, addressId int64, stateChange AddressStateChange) (int64, error) {
	_, err := ctx.tx.Exec(a.getQuery(archiveAddressStateQuery), addressId)
	if err != nil {
		return 0, err
	}
	var id int64
	err = ctx.tx.QueryRow(a.getQuery(insertAddressStateQuery), addressId, stateChange.NewState, ctx.blockId, stateChange.TxHash).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveBalances(ctx *context, balances []Balance) error {
	if len(balances) == 0 {
		return nil
	}
	for _, balance := range balances {
		if err := a.saveBalance(ctx, balance); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveBalance(ctx *context, balance Balance) error {
	_, err := ctx.tx.Exec(a.getQuery(insertBalanceQuery), balance.Address, balance.Balance, balance.Stake, ctx.blockId)
	return errors.Wrapf(err, "unable to save balance")
}

func (a *postgresAccessor) saveIdentities(ctx *context, identities []EpochIdentity) error {
	if len(identities) == 0 {
		return nil
	}
	for _, identity := range identities {
		identityStateId, err := a.saveIdentityState(ctx, identity)
		if err != nil {
			return err
		}
		if _, err = a.saveEpochIdentity(ctx, identityStateId, identity); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveIdentityState(ctx *context, identity EpochIdentity) (int64, error) {
	_, err := ctx.tx.Exec(a.getQuery(archiveIdentityStateQuery), identity.Address)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to execute query %s", archiveIdentityStateQuery)
	}
	var id int64
	err = ctx.tx.QueryRow(a.getQuery(insertIdentityStateQuery), identity.Address, identity.State, ctx.blockId).Scan(&id)
	return id, errors.Wrapf(err, "unable to execute query %s", insertIdentityStateQuery)
}

func (a *postgresAccessor) saveEpochIdentity(ctx *context, identityStateId int64, identity EpochIdentity) (int64, error) {
	var id int64

	if err := ctx.tx.QueryRow(a.getQuery(insertEpochIdentityQuery), ctx.epochId, identityStateId, identity.ShortPoint,
		identity.ShortFlips, identity.TotalShortPoint, identity.TotalShortFlips,
		identity.LongPoint, identity.LongFlips, identity.Approved, identity.Missed).Scan(&id); err != nil {
		return 0, errors.Wrapf(err, "unable to execute query %s", insertEpochIdentityQuery)
	}

	if ctx.epochIdentityIdsPerAddr == nil {
		ctx.epochIdentityIdsPerAddr = make(map[string]int64)
	}
	ctx.epochIdentityIdsPerAddr[identity.Address] = id

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
	return id, errors.Wrapf(err, "unable to execute query %s", insertFlipsToSolveQuery)
}

func (a *postgresAccessor) saveTransactions(ctx *context, txs []Transaction) (map[string]int64, error) {
	if len(txs) == 0 {
		return nil, nil
	}
	txIdsPerHash := make(map[string]int64)
	for _, tx := range txs {
		id, err := a.saveTransaction(ctx, tx)
		if err != nil {
			return nil, err
		}
		txIdsPerHash[tx.Hash] = id
	}
	return txIdsPerHash, nil
}

func (a *postgresAccessor) saveTransaction(ctx *context, tx Transaction) (int64, error) {
	var id int64
	from, err := ctx.addrId(tx.From)
	if err != nil {
		return 0, err
	}
	var to interface{}
	if len(tx.To) > 0 {
		to, err = ctx.addrId(tx.To)
		if err != nil {
			return 0, err
		}
	} else {
		to = nil
	}
	err = ctx.tx.QueryRow(a.getQuery(insertTransactionQuery), tx.Hash, ctx.blockId, tx.Type, from, to,
		tx.Amount, tx.Fee).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveSubmittedFlips(ctx *context, flips []Flip) (map[string]int64, error) {
	if len(flips) == 0 {
		return nil, nil
	}
	flipIdsPerCid := make(map[string]int64)
	for _, flip := range flips {
		txId, err := ctx.txId(flip.TxHash)
		if err != nil {
			return nil, err
		}
		id, err := a.saveSubmittedFlip(ctx, txId, flip)
		if err != nil {
			return nil, err
		}
		flipIdsPerCid[flip.Cid] = id
	}
	return flipIdsPerCid, nil
}

func (a *postgresAccessor) saveSubmittedFlip(ctx *context, txId int64, flip Flip) (int64, error) {
	var id int64
	err := ctx.tx.QueryRow(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId, flip.Size).Scan(&id)
	return id, errors.Wrapf(err, "unable to save flip %s", flip.Cid)
}

func (a *postgresAccessor) saveFlipKeys(ctx *context, keys []FlipKey) error {
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		txId, err := ctx.txId(key.TxHash)
		if err != nil {
			return err
		}
		err = a.saveFlipKey(ctx, txId, key)
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
