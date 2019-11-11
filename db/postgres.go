package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/monitoring"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
	"sync"
)

type postgresAccessor struct {
	db      *sql.DB
	pm      monitoring.PerformanceMonitor
	queries map[string]string
	mutex   sync.Mutex
}

const (
	initQuery                       = "init.sql"
	maxHeightQuery                  = "maxHeight.sql"
	currentFlipsQuery               = "currentFlips.sql"
	currentFlipCidsWithoutDataQuery = "currentFlipCidsWithoutData.sql"
	updateFlipStateQuery            = "updateFlipState.sql"
	insertFlipDataQuery             = "insertFlipData.sql"
	flipDataCountQuery              = "flipDataCount.sql"
	updateFlipSizeQuery             = "updateFlipSize.sql"
	insertFlipPicQuery              = "insertFlipPic.sql"
	insertFlipIconQuery             = "insertFlipIcon.sql"
	insertFlipPicOrderQuery         = "insertFlipPicOrder.sql"
	insertAnswersQuery              = "insertAnswers.sql"
	insertBlockQuery                = "insertBlock.sql"
	insertBlockProposerQuery        = "insertBlockProposer.sql"
	selectIdentityQuery             = "selectIdentity.sql"
	selectFlipQuery                 = "selectFlip.sql"
	insertEpochIdentityQuery        = "insertEpochIdentity.sql"
	insertTransactionQuery          = "insertTransaction.sql"
	insertSubmittedFlipQuery        = "insertSubmittedFlip.sql"
	insertFlipKeyQuery              = "insertFlipKey.sql"
	insertFlipWordsQuery            = "insertFlipWords.sql"
	flipWordsCountQuery             = "flipWordsCount.sql"
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
	updateBalanceQuery              = "updateBalance.sql"
	insertBirthdayQuery             = "insertBirthday.sql"
	updateBirthdayQuery             = "updateBirthday.sql"
	insertCoinsQuery                = "insertCoins.sql"
	insertBlockFlagQuery            = "insertBlockFlag.sql"
	insertEpochSummaryQuery         = "insertEpochSummary.sql"
	insertPenaltyQuery              = "insertPenalty.sql"
	selectLastPenaltyQuery          = "selectLastPenalty.sql"
	insertPaidPenaltyQuery          = "insertPaidPenalty.sql"
	insertBadAuthorQuery            = "insertBadAuthor.sql"
	insertGoodAuthorQuery           = "insertGoodAuthor.sql"
	insertTotalRewardsQuery         = "insertTotalRewards.sql"
	insertValidationRewardQuery     = "insertValidationReward.sql"
	insertRewardAgeQuery            = "insertRewardAge.sql"
	insertFundRewardQuery           = "insertFundReward.sql"
	insertFailedValidationQuery     = "insertFailedValidation.sql"
	insertMiningRewardsQuery        = "insertMiningRewards.sql"
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

func (a *postgresAccessor) GetCurrentFlips(address string) ([]Flip, error) {
	rows, err := a.db.Query(a.getQuery(currentFlipsQuery), address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Flip
	for rows.Next() {
		var item Flip
		err = rows.Scan(&item.Id, &item.Cid, &item.Pair)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) GetCurrentFlipsWithoutData(limit uint32) ([]AddressFlipCid, error) {
	rows, err := a.db.Query(a.getQuery(currentFlipCidsWithoutDataQuery), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []AddressFlipCid
	for rows.Next() {
		item := AddressFlipCid{}
		err = rows.Scan(&item.FlipId, &item.Cid, &item.Address)
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
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.pm.Start("InitTx")
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	a.pm.Complete("InitTx")
	a.pm.Start("RunTx")
	ctx := newContext(a, tx, data.Epoch, data.Block.Height)

	if err = a.saveEpoch(ctx, data.Epoch, data.ValidationTime); err != nil {
		return err
	}

	if err = a.saveBlock(ctx, data.Block); err != nil {
		return err
	}

	if err = a.saveBlockFlags(ctx, data.Block.Flags); err != nil {
		return err
	}

	if ctx.addrIdsPerAddr, err = a.saveAddresses(ctx, data.Addresses); err != nil {
		return err
	}

	if err = a.saveProposer(ctx, data.Block.Proposer); err != nil {
		return err
	}

	ctx.txIdsPerHash, err = a.saveTransactions(ctx, data.Block.Transactions)
	if err != nil {
		return err
	}

	if err = a.saveAddressStates(ctx, data.Addresses); err != nil {
		return err
	}

	if err := a.saveCoins(ctx, data.Coins); err != nil {
		return err
	}

	if err := a.saveBalances(ctx.tx, data.BalanceUpdates); err != nil {
		return err
	}

	if err := a.saveBirthdays(ctx.tx, data.Birthdays); err != nil {
		return err
	}

	if _, err := a.saveSubmittedFlips(ctx, data.SubmittedFlips); err != nil {
		return err
	}

	if err := a.saveFlipKeys(ctx, data.FlipKeys); err != nil {
		return err
	}

	if err := a.saveFlipsWords(ctx, data.FlipsWords); err != nil {
		return err
	}

	if err := a.saveFlipsData(ctx, data.FlipsData); err != nil {
		return err
	}

	if err := a.updateFlipSizes(ctx, data.FlipSizeUpdates); err != nil {
		return err
	}

	if err = a.saveIdentities(ctx, data.Identities); err != nil {
		return err
	}

	if err = a.saveFlipsStats(ctx, data.FlipStats); err != nil {
		return err
	}

	if data.SaveEpochSummary {
		if err = a.saveEpochSummary(ctx, data.Coins); err != nil {
			return err
		}
	}

	if err = a.savePenalty(ctx, data.Penalty); err != nil {
		return err
	}

	if err = a.savePaidPenalties(ctx, data.BurntPenalties); err != nil {
		return err
	}

	if err = a.saveEpochRewards(ctx, data.EpochRewards); err != nil {
		return err
	}

	a.pm.Start("saveMiningRewards")
	if err = a.saveMiningRewards(ctx, data.MiningRewards); err != nil {
		return err
	}
	a.pm.Complete("saveMiningRewards")

	if err = a.saveFailedValidation(ctx, data.FailedValidation); err != nil {
		return err
	}
	a.pm.Complete("RunTx")
	a.pm.Start("CommitTx")
	defer a.pm.Complete("CommitTx")
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
	if _, err := ctx.tx.Exec(a.getQuery(updateFlipStateQuery),
		flipStats.Status,
		flipStats.Answer,
		flipStats.WrongWords,
		ctx.blockHeight,
		flipStats.Cid); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveFlipsData(ctx *context, flipsData []FlipData) error {
	for _, flipData := range flipsData {
		if err := a.saveFlipData(ctx, flipData); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipData(ctx *context, flipData FlipData) error {
	count, err := a.getFlipDataCount(ctx, flipData.FlipId)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Warn(fmt.Sprintf("ignored duplicated flip data, flip id: %v, tx: %v", flipData.FlipId, flipData.TxHash))
		return nil
	}
	var txId *int64
	if len(flipData.TxHash) > 0 {
		id, err := ctx.txId(flipData.TxHash)
		if err != nil {
			return err
		}
		txId = &id
	}
	var flipDataId int64
	if err := ctx.tx.QueryRow(a.getQuery(insertFlipDataQuery), flipData.FlipId, ctx.blockHeight, txId).Scan(&flipDataId); err != nil {
		return errors.Wrapf(err, "unable to save flip data, flip id: %v, tx: %v", flipData.FlipId, flipData.TxHash)
	}
	for picIndex, pic := range flipData.Content.Pics {
		if err := a.saveFlipPic(ctx, byte(picIndex), pic, flipDataId); err != nil {
			return err
		}
	}
	for answerIndex, order := range flipData.Content.Orders {
		for posIndex, flipPicIndex := range order {
			if err := a.saveFlipPicOrder(ctx, byte(answerIndex), byte(posIndex), flipPicIndex, flipDataId); err != nil {
				return err
			}
		}
	}
	if flipData.Content.Icon != nil {
		if err := a.saveFlipIcon(ctx, flipData.Content.Icon, flipDataId); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) getFlipDataCount(ctx *context, flipId uint64) (int, error) {
	var count int
	err := ctx.tx.QueryRow(a.getQuery(flipDataCountQuery), flipId).Scan(&count)
	return count, errors.Wrapf(err, "unable to get flip data count for flip id %v", flipId)
}

func (a *postgresAccessor) updateFlipSizes(ctx *context, flipSizeUpdates []FlipSizeUpdate) error {
	if len(flipSizeUpdates) == 0 {
		return nil
	}
	for _, flipSizeUpdate := range flipSizeUpdates {
		if err := a.updateFlipSize(ctx, flipSizeUpdate); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) updateFlipSize(ctx *context, flipSizeUpdate FlipSizeUpdate) error {
	_, err := ctx.tx.Exec(a.getQuery(updateFlipSizeQuery), flipSizeUpdate.Size, flipSizeUpdate.Cid)
	return err
}

func (a *postgresAccessor) saveFlipPic(ctx *context, picIndex byte, pic []byte, flipDataId int64) error {
	_, err := ctx.tx.Exec(a.getQuery(insertFlipPicQuery), flipDataId, picIndex, pic)
	return err
}

func (a *postgresAccessor) saveFlipIcon(ctx *context, icon []byte, flipDataId int64) error {
	if icon == nil {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertFlipIconQuery), flipDataId, icon)
	return err
}

func (a *postgresAccessor) saveFlipPicOrder(ctx *context, answerIndex, posIndex, flipPicIndex byte, flipDataId int64) error {
	_, err := ctx.tx.Exec(a.getQuery(insertFlipPicOrderQuery), flipDataId, answerIndex, posIndex, flipPicIndex)
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
	err = ctx.tx.QueryRow(a.getQuery(insertAnswersQuery),
		flipId,
		epochIdentityId,
		isShort,
		answer.Answer,
		answer.WrongWords,
		answer.Point).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveEpoch(ctx *context, epoch uint64, validationTime big.Int) error {
	var savedEpoch int64
	err := ctx.tx.QueryRow(a.getQuery(selectEpochQuery), epoch).Scan(&savedEpoch)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = ctx.tx.Exec(a.getQuery(insertEpochQuery), epoch, validationTime.Int64())
	return err
}

func (a *postgresAccessor) saveBlock(ctx *context, block Block) error {
	_, err := ctx.tx.Exec(a.getQuery(insertBlockQuery),
		block.Height,
		block.Hash,
		ctx.epoch,
		block.Time.Int64(),
		block.IsEmpty,
		block.ValidatorsCount,
		block.Size)
	return err
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
	_, err := ctx.tx.Exec(a.getQuery(insertBlockFlagQuery), ctx.blockHeight, flag)
	return errors.Wrapf(err, "unable to save block flag")
}

func (a *postgresAccessor) saveProposer(ctx *context, proposer string) error {
	if len(proposer) == 0 {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertBlockProposerQuery), ctx.blockHeight, proposer)
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
	err = ctx.tx.QueryRow(a.getQuery(insertAddressQuery), address.Address, ctx.blockHeight).Scan(&id)
	return id, err
}

func (a *postgresAccessor) saveTemporaryIdentity(ctx *context, addressId int64) error {
	_, err := ctx.tx.Exec(a.getQuery(insertTemporaryIdentityQuery), addressId, ctx.blockHeight)
	return errors.Wrapf(err, "unable to save temporary identity")
}

func (a *postgresAccessor) saveAddressState(ctx *context, addressId int64, stateChange AddressStateChange) (int64, error) {
	var prevId int64
	err := ctx.tx.QueryRow(a.getQuery(archiveAddressStateQuery), addressId).Scan(&prevId)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	var id int64
	if prevId > 0 {
		err = ctx.tx.QueryRow(a.getQuery(insertAddressStateQuery), addressId, stateChange.NewState, ctx.blockHeight, stateChange.TxHash, prevId).Scan(&id)
	} else {
		err = ctx.tx.QueryRow(a.getQuery(insertAddressStateQuery), addressId, stateChange.NewState, ctx.blockHeight, stateChange.TxHash, nil).Scan(&id)
	}
	return id, err
}

func (a *postgresAccessor) saveCoins(ctx *context, coins Coins) error {
	_, err := ctx.tx.Exec(a.getQuery(insertCoinsQuery),
		ctx.blockHeight,
		coins.Burnt,
		coins.Minted,
		coins.TotalBalance,
		coins.TotalStake)

	return errors.Wrapf(err, "unable to save coins %v", coins)
}

func (a *postgresAccessor) saveBalances(tx *sql.Tx, balances []Balance) error {
	for _, balance := range balances {
		if err := a.saveBalance(tx, balance); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveBalance(tx *sql.Tx, balance Balance) error {
	var addressId uint64
	err := tx.QueryRow(a.getQuery(updateBalanceQuery),
		balance.Address,
		balance.Balance,
		balance.Stake).Scan(&addressId)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(a.getQuery(insertBalanceQuery),
			balance.Address,
			balance.Balance,
			balance.Stake)
	}
	return errors.Wrapf(err, "unable to save balance")
}

func (a *postgresAccessor) saveBirthdays(tx *sql.Tx, birthdays []Birthday) error {
	for _, birthday := range birthdays {
		if err := a.saveBirthday(tx, birthday); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveBirthday(tx *sql.Tx, birthday Birthday) error {
	var addressId uint64
	err := tx.QueryRow(a.getQuery(updateBirthdayQuery),
		birthday.Address,
		birthday.BirthEpoch).Scan(&addressId)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(a.getQuery(insertBirthdayQuery),
			birthday.Address,
			birthday.BirthEpoch)
	}
	return errors.Wrapf(err, "unable to save birthday %v", birthday)
}

func (a *postgresAccessor) saveIdentities(ctx *context, identities []EpochIdentity) error {
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
	var prevId int64
	err := ctx.tx.QueryRow(a.getQuery(archiveIdentityStateQuery), identity.Address).Scan(&prevId)
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.Wrapf(err, "unable to execute query %s", archiveIdentityStateQuery)
	}
	var id int64
	if prevId > 0 {
		err = ctx.tx.QueryRow(a.getQuery(insertIdentityStateQuery), identity.Address, identity.State, ctx.blockHeight, prevId).Scan(&id)
	} else {
		err = ctx.tx.QueryRow(a.getQuery(insertIdentityStateQuery), identity.Address, identity.State, ctx.blockHeight, nil).Scan(&id)
	}
	return id, errors.Wrapf(err, "unable to execute query %s", insertIdentityStateQuery)
}

func (a *postgresAccessor) saveEpochIdentity(ctx *context, identityStateId int64, identity EpochIdentity) (int64, error) {
	var id int64

	if err := ctx.tx.QueryRow(a.getQuery(insertEpochIdentityQuery), ctx.epoch, identityStateId, identity.ShortPoint,
		identity.ShortFlips, identity.TotalShortPoint, identity.TotalShortFlips,
		identity.LongPoint, identity.LongFlips, identity.Approved, identity.Missed,
		identity.RequiredFlips, identity.MadeFlips).Scan(&id); err != nil {
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
	err = ctx.tx.QueryRow(a.getQuery(insertTransactionQuery),
		tx.Hash,
		ctx.blockHeight,
		tx.Type,
		from,
		to,
		tx.Amount,
		tx.Tips,
		tx.MaxFee,
		tx.Fee,
		tx.Size).Scan(&id)
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
	err := ctx.tx.QueryRow(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId, flip.Size, flip.Pair).Scan(&id)
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

func (a *postgresAccessor) saveFlipsWords(ctx *context, words []FlipWords) error {
	for _, key := range words {
		txId, err := ctx.txId(key.TxHash)
		if err != nil {
			return err
		}
		if err := a.saveFlipWords(ctx, txId, key); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFlipWords(ctx *context, txId int64, words FlipWords) error {
	count, err := a.getFlipWordsCount(ctx, words.FlipId)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Warn(fmt.Sprintf("ignored duplicated flip words: %v", words))
		return nil
	}
	_, err = ctx.tx.Exec(a.getQuery(insertFlipWordsQuery), words.FlipId, words.Word1, words.Word2, txId)
	return errors.Wrapf(err, "unable to save flip words %v", words)
}

func (a *postgresAccessor) getFlipWordsCount(ctx *context, flipId uint64) (int, error) {
	var count int
	err := ctx.tx.QueryRow(a.getQuery(flipWordsCountQuery), flipId).Scan(&count)
	return count, errors.Wrapf(err, "unable to get flip words count for flip id %v", flipId)
}

func (a *postgresAccessor) saveEpochSummary(ctx *context, coins Coins) error {
	_, err := ctx.tx.Exec(a.getQuery(insertEpochSummaryQuery), ctx.epoch, ctx.blockHeight, coins.TotalBalance, coins.TotalStake)
	return errors.Wrapf(err, "unable to save epoch summary for %s", ctx.epoch)
}

func (a *postgresAccessor) savePenalty(ctx *context, penalty *Penalty) error {
	if penalty == nil {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertPenaltyQuery), penalty.Address, penalty.Penalty, ctx.blockHeight)
	return errors.Wrapf(err, "unable to save penalty")
}

func (a *postgresAccessor) savePaidPenalties(ctx *context, burntPenalties []Penalty) error {
	if len(burntPenalties) == 0 {
		return nil
	}
	for _, burntPenalty := range burntPenalties {
		if err := a.savePaidPenalty(ctx, burntPenalty); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) savePaidPenalty(ctx *context, burntPenalty Penalty) error {
	var id int64
	var penalty decimal.Decimal
	err := ctx.tx.QueryRow(a.getQuery(selectLastPenaltyQuery), burntPenalty.Address).Scan(&id, &penalty)
	if err != nil {
		return errors.Wrapf(err, "unable to get last penalty")
	}
	paidPenalty := penalty.Sub(burntPenalty.Penalty)
	_, err = ctx.tx.Exec(a.getQuery(insertPaidPenaltyQuery), id, paidPenalty, ctx.blockHeight)
	return errors.Wrapf(err, "unable to save paid penalty")
}

func (a *postgresAccessor) saveEpochRewards(ctx *context, epochRewards *EpochRewards) error {
	if epochRewards == nil {
		return nil
	}
	if err := a.saveBadAuthors(ctx, epochRewards.BadAuthors); err != nil {
		return err
	}
	if err := a.saveGoodAuthors(ctx, epochRewards.GoodAuthors); err != nil {
		return err
	}
	if err := a.saveTotalRewards(ctx, epochRewards.Total); err != nil {
		return err
	}
	if err := a.saveValidationRewards(ctx, epochRewards.ValidationRewards); err != nil {
		return err
	}
	if err := a.saveRewardAges(ctx, epochRewards.AgesByAddress); err != nil {
		return err
	}
	if err := a.saveFundRewards(ctx, epochRewards.FundRewards); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveBadAuthors(ctx *context, badAuthors []string) error {
	for _, badAuthor := range badAuthors {
		if err := a.saveBadAuthor(ctx, badAuthor); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveBadAuthor(ctx *context, badAuthor string) error {
	_, err := ctx.tx.Exec(a.getQuery(insertBadAuthorQuery), ctx.epochIdentityIdsPerAddr[badAuthor])
	return errors.Wrapf(err, "unable to save bad author")
}

func (a *postgresAccessor) saveGoodAuthors(ctx *context, goodAuthors []*ValidationResult) error {
	for _, goodAuthor := range goodAuthors {
		if err := a.saveGoodAuthor(ctx, goodAuthor); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveGoodAuthor(ctx *context, goodAuthor *ValidationResult) error {
	_, err := ctx.tx.Exec(a.getQuery(insertGoodAuthorQuery),
		ctx.epochIdentityIdsPerAddr[goodAuthor.Address],
		goodAuthor.StrongFlips,
		goodAuthor.WeakFlips,
		goodAuthor.SuccessfulInvites)
	return errors.Wrapf(err, "unable to save good author")
}

func (a *postgresAccessor) saveTotalRewards(ctx *context, totalRewards *TotalRewards) error {
	if totalRewards == nil {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertTotalRewardsQuery),
		ctx.blockHeight,
		totalRewards.Total,
		totalRewards.Validation,
		totalRewards.Flips,
		totalRewards.Invitations,
		totalRewards.FoundationPayouts,
		totalRewards.ZeroWalletFund)
	return errors.Wrapf(err, "unable to save total rewards")
}

func (a *postgresAccessor) saveValidationRewards(ctx *context, rewards []*Reward) error {
	for _, reward := range rewards {
		if err := a.saveValidationReward(ctx, reward); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveValidationReward(ctx *context, reward *Reward) error {
	_, err := ctx.tx.Exec(a.getQuery(insertValidationRewardQuery),
		ctx.epochIdentityIdsPerAddr[reward.Address],
		reward.Balance,
		reward.Stake,
		reward.Type)
	return errors.Wrapf(err, "unable to save validation reward")
}

func (a *postgresAccessor) saveRewardAges(ctx *context, agesByAddress map[string]uint16) error {
	for address, age := range agesByAddress {
		if err := a.saveRewardAge(ctx, address, age); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveRewardAge(ctx *context, address string, age uint16) error {
	_, err := ctx.tx.Exec(a.getQuery(insertRewardAgeQuery), ctx.epochIdentityIdsPerAddr[address], age)
	return errors.Wrapf(err, "unable to save reward age")
}

func (a *postgresAccessor) saveFundRewards(ctx *context, rewards []*Reward) error {
	for _, reward := range rewards {
		if err := a.saveFundReward(ctx, reward); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveFundReward(ctx *context, reward *Reward) error {
	_, err := ctx.tx.Exec(a.getQuery(insertFundRewardQuery),
		reward.Address,
		ctx.blockHeight,
		reward.Balance,
		reward.Type)
	return errors.Wrapf(err, "unable to save fund reward: %v", reward)
}

func (a *postgresAccessor) saveMiningRewards(ctx *context, rewards []*Reward) error {
	if len(rewards) == 0 {
		return nil
	}
	if _, err := ctx.tx.Exec(a.getQuery(insertMiningRewardsQuery), ctx.blockHeight, pq.Array(rewards)); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveFailedValidation(ctx *context, failed bool) error {
	if !failed {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertFailedValidationQuery), ctx.blockHeight)
	return errors.Wrapf(err, "unable to save failed validation")
}

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}
