package db

import (
	"database/sql"
	"fmt"
	"github.com/idena-network/idena-go/common"
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
	initQuery                           = "init.sql"
	maxHeightQuery                      = "maxHeight.sql"
	currentFlipsQuery                   = "currentFlips.sql"
	currentFlipCidsWithoutDataQuery     = "currentFlipCidsWithoutData.sql"
	insertFlipDataQuery                 = "insertFlipData.sql"
	flipDataCountQuery                  = "flipDataCount.sql"
	updateFlipSizeQuery                 = "updateFlipSize.sql"
	insertFlipPicQuery                  = "insertFlipPic.sql"
	insertFlipIconQuery                 = "insertFlipIcon.sql"
	insertFlipPicOrderQuery             = "insertFlipPicOrder.sql"
	insertBlockQuery                    = "insertBlock.sql"
	insertBlockProposerQuery            = "insertBlockProposer.sql"
	insertBlockProposerVrfScoreQuery    = "insertBlockProposerVrfScore.sql"
	selectIdentityQuery                 = "selectIdentity.sql"
	selectFlipQuery                     = "selectFlip.sql"
	insertAddressesAndTransactionsQuery = "insertAddressesAndTransactions.sql"
	insertSubmittedFlipQuery            = "insertSubmittedFlip.sql"
	insertFlipKeyQuery                  = "insertFlipKey.sql"
	insertFlipWordsQuery                = "insertFlipWords.sql"
	flipWordsCountQuery                 = "flipWordsCount.sql"
	selectEpochQuery                    = "selectEpoch.sql"
	insertEpochQuery                    = "insertEpoch.sql"
	resetToBlockQuery                   = "resetToBlock.sql"
	saveBalancesQuery                   = "saveBalances.sql"
	saveBirthdaysQuery                  = "saveBirthdays.sql"
	insertCoinsQuery                    = "insertCoins.sql"
	insertBlockFlagQuery                = "insertBlockFlag.sql"
	insertEpochSummaryQuery             = "insertEpochSummary.sql"
	insertPenaltyQuery                  = "insertPenalty.sql"
	selectLastPenaltyQuery              = "selectLastPenalty.sql"
	insertPaidPenaltyQuery              = "insertPaidPenalty.sql"
	insertFailedValidationQuery         = "insertFailedValidation.sql"
	insertMiningRewardsQuery            = "insertMiningRewards.sql"
	insertBurntCoinsQuery               = "insertBurntCoins.sql"
	insertFlipStatsQuery                = "insertFlipStats.sql"
	saveEpochIdentitiesQuery            = "saveEpochIdentities.sql"
	saveEpochRewardsQuery               = "saveEpochRewards.sql"
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

	if ctx.txIdsPerHash, err = a.saveAddressesAndTransactions(ctx, data.Addresses, data.Block.Transactions); err != nil {
		return err
	}

	if err = a.saveProposer(ctx, data.Block.Proposer); err != nil {
		return err
	}

	if err = a.saveProposerVrfScore(ctx, data.Block.ProposerVrfScore); err != nil {
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

	if err = a.saveBurntCoins(ctx, data.BurntCoinsPerAddr); err != nil {
		return err
	}

	if err = a.saveFailedValidation(ctx, data.FailedValidation); err != nil {
		return err
	}
	a.pm.Complete("RunTx")
	a.pm.Start("CommitTx")
	defer a.pm.Complete("CommitTx")
	return tx.Commit()
}

func (a *postgresAccessor) saveFlipsStats(ctx *context, flipsStats []FlipStats) error {
	if len(flipsStats) == 0 {
		return nil
	}
	answersArray, statesArray := getFlipStatsArrays(flipsStats)
	_, err := ctx.tx.Exec(a.getQuery(insertFlipStatsQuery),
		ctx.blockHeight,
		answersArray,
		statesArray)
	return errors.Wrap(err, "unable to save flip stats")
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
		block.BodySize,
		block.VrfProposerThreshold,
		block.FullSize,
		block.FeeRate)
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

func (a *postgresAccessor) saveProposerVrfScore(ctx *context, vrfScore float64) error {
	if vrfScore == 0 {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertBlockProposerVrfScoreQuery), ctx.blockHeight, vrfScore)
	return err
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
	if len(balances) == 0 {
		return nil
	}
	if _, err := tx.Exec(a.getQuery(saveBalancesQuery), pq.Array(balances)); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveBirthdays(tx *sql.Tx, birthdays []Birthday) error {
	if len(birthdays) == 0 {
		return nil
	}
	if _, err := tx.Exec(a.getQuery(saveBirthdaysQuery), pq.Array(birthdays)); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveIdentities(ctx *context, identities []EpochIdentity) error {
	if len(identities) == 0 {
		return nil
	}
	identitiesArray, flipsToSolveArray := getEpochIdentitiesArrays(identities)
	_, err := ctx.tx.Exec(a.getQuery(saveEpochIdentitiesQuery),
		ctx.epoch,
		ctx.blockHeight,
		identitiesArray,
		flipsToSolveArray)
	return errors.Wrap(err, "unable to save identities")
}

func (a *postgresAccessor) saveAddressesAndTransactions(ctx *context, addresses []Address,
	txs []Transaction) (map[string]int64, error) {

	if len(addresses)+len(txs) == 0 {
		return nil, nil
	}

	addressesArray, addressStateChangesArray := getPostgresAddressesAndAddressStateChangesArrays(addresses)
	var txHashIds []txHashId
	err := ctx.tx.QueryRow(a.getQuery(insertAddressesAndTransactionsQuery),
		ctx.blockHeight,
		addressesArray,
		pq.Array(txs),
		addressStateChangesArray).Scan(pq.Array(&txHashIds))
	if err != nil {
		return nil, errors.Wrap(err, "unable to save addresses and transactions")
	}
	txIdsPerHash := make(map[string]int64)
	for _, txHashId := range txHashIds {
		txIdsPerHash[txHashId.Hash] = txHashId.Id
	}
	return txIdsPerHash, nil
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
	err := ctx.tx.QueryRow(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId, flip.Pair).Scan(&id)
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
	return errors.Wrapf(err, "unable to save epoch summary for %v", ctx.epoch)
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
	_, err := ctx.tx.Exec(a.getQuery(saveEpochRewardsQuery),
		ctx.blockHeight,
		pq.Array(epochRewards.BadAuthors),
		pq.Array(epochRewards.GoodAuthors),
		epochRewards.Total,
		pq.Array(epochRewards.ValidationRewards),
		getRewardAgesArray(epochRewards.AgesByAddress),
		pq.Array(epochRewards.FundRewards),
	)
	return errors.Wrap(err, "unable to save epoch rewards")
}

func (a *postgresAccessor) saveMiningRewards(ctx *context, rewards []*MiningReward) error {
	if len(rewards) == 0 {
		return nil
	}
	if _, err := ctx.tx.Exec(a.getQuery(insertMiningRewardsQuery), ctx.blockHeight, pq.Array(rewards)); err != nil {
		return err
	}
	return nil
}

func (a *postgresAccessor) saveBurntCoins(ctx *context, burntCoinsByAddr map[common.Address][]*BurntCoins) error {
	if len(burntCoinsByAddr) == 0 {
		return nil
	}
	if _, err := ctx.tx.Exec(a.getQuery(insertBurntCoinsQuery),
		ctx.blockHeight, pq.Array(getPostgresBurntCoins(burntCoinsByAddr, ctx.txId))); err != nil {
		return errors.Wrap(err, "unable to save burnt coins")
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
