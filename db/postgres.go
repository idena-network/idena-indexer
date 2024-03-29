package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/monitoring"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"sync"
)

type postgresAccessor struct {
	db                        *sql.DB
	pm                        monitoring.PerformanceMonitor
	queries                   map[string]string
	mutex                     sync.Mutex
	changesHistoryBlocksCount int
	miningRewards             bool
	dataTable, dataStateTable string
}

const (
	initQuery                           = "init.sql"
	maxHeightQuery                      = "maxHeight.sql"
	insertBlockQuery                    = "insertBlock.sql"
	insertBlockProposerQuery            = "insertBlockProposer.sql"
	insertBlockProposerVrfScoreQuery    = "insertBlockProposerVrfScore.sql"
	insertAddressesAndTransactionsQuery = "insertAddressesAndTransactions.sql"
	insertSubmittedFlipQuery            = "insertSubmittedFlip.sql"
	insertFlipKeyQuery                  = "insertFlipKey.sql"
	insertEpochQuery                    = "insertEpoch.sql"
	resetToBlockQuery                   = "resetToBlock.sql"
	saveBalancesQuery                   = "saveBalances.sql"
	saveBirthdaysQuery                  = "saveBirthdays.sql"
	insertCoinsQuery                    = "insertCoins.sql"
	insertPenaltyQuery                  = "insertPenalty.sql"
	insertMiningRewardsQuery            = "insertMiningRewards.sql"
	insertBurntCoinsQuery               = "insertBurntCoins.sql"
	saveEpochResultQuery                = "saveEpochResult.sql"
	saveFlipsWordsQuery                 = "saveFlipsWords.sql"
)

func (a *postgresAccessor) getQuery(name string) string {
	if query, present := a.queries[name]; present {
		return query
	}
	panic(fmt.Sprintf("There is no query '%s'", name))
}

func (a *postgresAccessor) GetLastHeight() (uint64, error) {
	var maxHeight int64
	err := a.db.QueryRow(a.getQuery(maxHeightQuery)).Scan(&maxHeight)
	if err != nil {
		return 0, err
	}
	return uint64(maxHeight), nil
}

func (a *postgresAccessor) ResetTo(height uint64) error {
	_, err := a.db.Exec(a.getQuery(resetToBlockQuery), height)
	return getResultError(err)
}

func (a *postgresAccessor) Save(data *Data) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.pm.Start("InitTx")
	tx, err := a.db.Begin()
	if err != nil {
		return getResultError(err)
	}
	defer tx.Rollback()

	a.pm.Complete("InitTx")
	a.pm.Start("RunTx")
	ctx := newContext(a, tx, data.Epoch, data.Block.Height)

	a.pm.Start("saveEpoch")
	if err = a.saveEpoch(ctx, data.Epoch, data.ValidationTime, data.PrevStateRoot, data.DiscriminationStakeThreshold); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveEpoch")

	a.pm.Start("saveBlock")
	if err = a.saveBlock(ctx, data.Block); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveBlock")

	a.pm.Start("saveAddressesAndTransactions")
	if ctx.txIdsPerHash, err = a.saveAddressesAndTransactions(
		ctx,
		data.Addresses,
		data.Block.Transactions,
		data.ActivationTxTransfers,
		data.KillTxTransfers,
		data.KillInviteeTxTransfers,
		data.DeletedFlips,
		data.ActivationTxs,
		data.KillInviteeTxs,
		data.BecomeOnlineTxs,
		data.BecomeOfflineTxs,
		data.Contracts,
		data.OracleVotingContracts,
		data.OracleVotingContractCallStarts,
		data.OracleVotingContractCallVoteProofs,
		data.OracleVotingContractCallVotes,
		data.OracleVotingContractCallFinishes,
		data.OracleVotingContractCallProlongations,
		data.OracleVotingContractCallAddStakes,
		data.OracleVotingContractTerminations,
		data.OracleVotingContractsToProlong,
		data.ClearOldEpochCommittees,
		data.OracleLockContracts,
		data.OracleLockContractCallCheckOracleVotings,
		data.OracleLockContractCallPushes,
		data.OracleLockContractTerminations,
		data.RefundableOracleLockContracts,
		data.RefundableOracleLockContractCallDeposits,
		data.RefundableOracleLockContractCallPushes,
		data.RefundableOracleLockContractCallRefunds,
		data.RefundableOracleLockContractTerminations,
		data.MultisigContracts,
		data.MultisigContractCallAdds,
		data.MultisigContractCallSends,
		data.MultisigContractCallPushes,
		data.MultisigContractTerminations,
		data.TimeLockContracts,
		data.TimeLockContractCallTransfers,
		data.TimeLockContractTerminations,
		data.ContractTxsBalanceUpdates,
		data.TxReceipts,
		data.DelegationSwitches,
		data.UpgradesVotes,
		data.PoolSizes,
		data.MinersHistoryItem,
		data.RemovedTransitiveDelegations,
		data.EpochSummaryUpdate,
		data.Tokens,
		data.TokenBalanceUpdates,
		data.DelegationHistoryUpdates,
	); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveAddressesAndTransactions")

	a.pm.Start("saveProposer")
	if err = a.saveProposer(ctx, data.Block.Proposer); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveProposer")

	a.pm.Start("saveProposerVrfScore")
	if err = a.saveProposerVrfScore(ctx, data.Block.ProposerVrfScore); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveProposerVrfScore")

	a.pm.Start("saveCoins")
	if err := a.saveCoins(ctx, data.Coins); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveCoins")

	a.pm.Start("saveBalances")
	if err := a.saveBalances(ctx.tx, ctx.blockHeight, data.ChangedBalances, data.BalanceUpdates, data.CommitteeRewardShare); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveBalances")

	a.pm.Start("saveSubmittedFlips")
	if err := a.saveSubmittedFlips(ctx, data.SubmittedFlips); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveSubmittedFlips")

	a.pm.Start("saveFlipKeys")
	if err := a.saveFlipKeys(ctx, data.FlipKeys); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveFlipKeys")

	a.pm.Start("saveFlipsWords")
	if err := a.saveFlipsWords(ctx, data.FlipsWords); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveFlipsWords")

	a.pm.Start("savePenalties")
	if err = a.savePenalties(ctx, data.Penalties); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("savePenalties")

	if a.miningRewards {
		a.pm.Start("saveMiningRewards")
		if err = a.saveMiningRewards(ctx, data.MiningRewards); err != nil {
			return getResultError(err)
		}
		a.pm.Complete("saveMiningRewards")
	}

	a.pm.Start("saveBurntCoins")
	if err = a.saveBurntCoins(ctx, data.BurntCoinsPerAddr); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveBurntCoins")

	a.pm.Start("saveEpochResult")
	if err = a.saveEpochResult(ctx.tx, ctx.epoch, ctx.blockHeight, data.EpochResult); err != nil {
		return getResultError(err)
	}
	a.pm.Complete("saveEpochResult")

	a.pm.Complete("RunTx")
	a.pm.Start("CommitTx")
	defer a.pm.Complete("CommitTx")
	return tx.Commit()
}

func (a *postgresAccessor) saveEpochResult(
	tx *sql.Tx,
	epoch uint64,
	height uint64,
	epochResult *EpochResult,
) error {
	if epochResult == nil {
		return nil
	}
	var identitiesArray, flipsToSolveArray, answersArray, statesArray, badAuthors, totalRewards, validationRewards,
		rewardAges, fundRewards, rewardedFlipCids, rewardedExtraFlipCids, rewardedInvitations, savedInviteRewards,
		reportedFlipRewards, rewardStakedAmounts, rewardFailedStakedAmounts, rewardedInvitees interface {
		driver.Valuer
	}
	var shortAnswerCountsByAddr, longAnswerCountsByAdds, wrongWordsFlipsCountsByAddr map[string]int
	if len(epochResult.FlipStats) > 0 {
		answersArray, statesArray, shortAnswerCountsByAddr, longAnswerCountsByAdds, wrongWordsFlipsCountsByAddr = getFlipStatsArrays(epochResult.FlipStats)
	}
	if len(epochResult.Identities) > 0 {
		identitiesArray, flipsToSolveArray = getEpochIdentitiesArrays(epochResult.Identities, shortAnswerCountsByAddr, longAnswerCountsByAdds, wrongWordsFlipsCountsByAddr)
	}
	epochRewards := epochResult.EpochRewards
	if epochRewards != nil {
		badAuthors = pq.Array(epochRewards.BadAuthors)
		totalRewards = epochRewards.Total
		validationRewards = pq.Array(epochRewards.ValidationRewards)
		rewardAges = getRewardAgesArray(epochRewards.AgesByAddress)
		rewardStakedAmounts = getRewardStakedAmountsArray(epochRewards.StakedAmountsByAddress)
		rewardFailedStakedAmounts = getRewardStakedAmountsArray(epochRewards.FailedStakedAmountsByAddress)
		fundRewards = pq.Array(epochRewards.FundRewards)
		rewardedFlipCids = pq.Array(epochRewards.RewardedFlipCids)
		rewardedExtraFlipCids = pq.Array(epochRewards.RewardedExtraFlipCids)
		rewardedInvitations = pq.Array(epochRewards.RewardedInvitations)
		savedInviteRewards = pq.Array(epochRewards.SavedInviteRewards)
		reportedFlipRewards = pq.Array(epochRewards.ReportedFlipRewards)
		rewardedInvitees = pq.Array(epochRewards.RewardedInvitees)
	}
	data := getEpochResultData(epochResult)
	if _, err := tx.Exec(
		a.getQuery(saveEpochResultQuery),
		epoch,
		height,
		pq.Array(epochResult.Birthdays),
		identitiesArray,
		flipsToSolveArray,
		pq.Array(epochResult.MemPoolFlipKeys),
		answersArray,
		statesArray,
		badAuthors,
		totalRewards,
		validationRewards,
		rewardAges,
		rewardStakedAmounts,
		rewardFailedStakedAmounts,
		fundRewards,
		rewardedFlipCids,
		rewardedExtraFlipCids,
		rewardedInvitations,
		savedInviteRewards,
		reportedFlipRewards,
		epochResult.FailedValidation,
		epochResult.MinScoreForInvite,
		rewardedInvitees,
		data,
	); err != nil {
		return errors.Wrap(err, "unable to save epoch result")
	}
	return nil
}

func (a *postgresAccessor) saveEpoch(ctx *context, epoch uint64, validationTime big.Int, root string, discriminationStakeThreshold *big.Int) error {
	var discriminationStakeThresholdD *decimal.Decimal
	if discriminationStakeThreshold != nil {
		v := blockchain.ConvertToFloat(discriminationStakeThreshold)
		discriminationStakeThresholdD = &v
	}
	_, err := ctx.tx.Exec(a.getQuery(insertEpochQuery), epoch, validationTime.Int64(), root, discriminationStakeThresholdD)
	return err
}

func (a *postgresAccessor) saveBlock(ctx *context, block Block) error {
	_, err := ctx.tx.Exec(a.getQuery(insertBlockQuery),
		block.Height,
		block.Hash,
		ctx.epoch,
		block.Time,
		block.IsEmpty,
		block.OriginalValidatorsCount,
		block.PoolValidatorsCount,
		block.BodySize,
		block.VrfProposerThreshold,
		block.FullSize,
		block.FeeRate,
		block.Upgrade,
		block.OfflineAddress,
		pq.Array(block.Flags),
		block.GasUsed,
	)
	return err
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

func (a *postgresAccessor) saveBalances(tx *sql.Tx, blockHeight uint64, balances []Balance,
	balanceUpdates []*BalanceUpdate, committeeRewardShare *big.Int) error {
	if _, err := tx.Exec(
		a.getQuery(saveBalancesQuery),
		blockHeight,
		pq.Array(balances),
		pq.Array(balanceUpdates),
		a.changesHistoryBlocksCount,
		blockchain.ConvertToFloat(committeeRewardShare),
	); err != nil {
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

func (a *postgresAccessor) saveAddressesAndTransactions(
	ctx *context,
	addresses []Address,
	txs []Transaction,
	activationTxTransfers []ActivationTxTransfer,
	killTxs []KillTxTransfer,
	killInviteeTxTransfers []KillInviteeTxTransfer,
	deletedFlips []DeletedFlip,
	activationTxs []ActivationTx,
	killInviteeTxs []KillInviteeTx,
	becomeOnlineTxs []string,
	becomeOfflineTxs []string,
	contracts []*Contract,
	oracleVotingContracts []*OracleVotingContract,
	oracleVotingContractCallStarts []*OracleVotingContractCallStart,
	oracleVotingContractCallVoteProofs []*OracleVotingContractCallVoteProof,
	oracleVotingContractCallVotes []*OracleVotingContractCallVote,
	oracleVotingContractCallFinishes []*OracleVotingContractCallFinish,
	oracleVotingContractCallProlongations []*OracleVotingContractCallProlongation,
	oracleVotingContractCallAddStakes []*OracleVotingContractCallAddStake,
	oracleVotingContractTerminations []*OracleVotingContractTermination,
	oracleVotingContractsToProlong []common.Address,
	clearOldEpochCommittees bool,
	oracleLockContracts []*OracleLockContract,
	oracleLockContractCallCheckOracleVotings []*OracleLockContractCallCheckOracleVoting,
	oracleLockContractCallPushes []*OracleLockContractCallPush,
	oracleLockContractTerminations []*OracleLockContractTermination,
	refundableOracleLockContracts []*RefundableOracleLockContract,
	refundableOracleLockContractCallDeposits []*RefundableOracleLockContractCallDeposit,
	refundableOracleLockContractCallPushes []*RefundableOracleLockContractCallPush,
	refundableOracleLockContractCallRefunds []*RefundableOracleLockContractCallRefund,
	refundableOracleLockContractTerminations []*RefundableOracleLockContractTermination,
	multisigContracts []*MultisigContract,
	multisigContractCallAdds []*MultisigContractCallAdd,
	multisigContractCallSends []*MultisigContractCallSend,
	multisigContractCallPushes []*MultisigContractCallPush,
	multisigContractTerminations []*MultisigContractTermination,
	timeLockContracts []*TimeLockContract,
	timeLockContractCallTransfers []*TimeLockContractCallTransfer,
	timeLockContractTerminations []*TimeLockContractTermination,
	contractTxsBalanceUpdates []*ContractTxBalanceUpdates,
	txReceipts []*TxReceipt,
	delegationSwitches []*DelegationSwitch,
	upgradesVotes []*UpgradeVotes,
	poolSizes []PoolSize,
	minersHistoryItem *MinersHistoryItem,
	removedTransitiveDelegations []RemovedTransitiveDelegation,
	epochSummaryUpdate EpochSummaryUpdate,
	tokens []Token,
	tokenBalanceUpdates []TokenBalance,
	delegationHistoryUpdates []DelegationHistoryUpdate,
) (map[string]int64, error) {

	addressesArray, addressStateChangesArray := getPostgresAddressesAndAddressStateChangesArrays(addresses)
	var txHashIds []txHashId
	data := getData(
		txs, delegationSwitches, upgradesVotes, poolSizes, minersHistoryItem, removedTransitiveDelegations,
		epochSummaryUpdate, oracleVotingContractsToProlong, txReceipts, contracts, tokens, tokenBalanceUpdates, delegationHistoryUpdates)
	err := ctx.tx.QueryRow(a.getQuery(insertAddressesAndTransactionsQuery),
		ctx.blockHeight,
		a.changesHistoryBlocksCount,
		addressesArray,
		pq.Array(activationTxTransfers),
		pq.Array(killTxs),
		pq.Array(killInviteeTxTransfers),
		addressStateChangesArray,
		pq.Array(deletedFlips),
		pq.Array(activationTxs),
		pq.Array(killInviteeTxs),
		pq.Array(becomeOnlineTxs),
		pq.Array(becomeOfflineTxs),
		pq.Array(oracleVotingContracts),
		pq.Array(oracleVotingContractCallStarts),
		pq.Array(oracleVotingContractCallVoteProofs),
		pq.Array(oracleVotingContractCallVotes),
		pq.Array(oracleVotingContractCallFinishes),
		pq.Array(oracleVotingContractCallProlongations),
		pq.Array(oracleVotingContractCallAddStakes),
		pq.Array(oracleVotingContractTerminations),
		clearOldEpochCommittees,
		pq.Array(oracleLockContracts),
		pq.Array(oracleLockContractCallCheckOracleVotings),
		pq.Array(oracleLockContractCallPushes),
		pq.Array(oracleLockContractTerminations),
		pq.Array(refundableOracleLockContracts),
		pq.Array(refundableOracleLockContractCallDeposits),
		pq.Array(refundableOracleLockContractCallPushes),
		pq.Array(refundableOracleLockContractCallRefunds),
		pq.Array(refundableOracleLockContractTerminations),
		pq.Array(timeLockContracts),
		pq.Array(timeLockContractCallTransfers),
		pq.Array(timeLockContractTerminations),
		pq.Array(multisigContracts),
		pq.Array(multisigContractCallAdds),
		pq.Array(multisigContractCallSends),
		pq.Array(multisigContractCallPushes),
		pq.Array(multisigContractTerminations),
		pq.Array(contractTxsBalanceUpdates),
		data,
	).Scan(pq.Array(&txHashIds))
	if err != nil {
		return nil, errors.Wrap(err, "unable to save addresses and transactions")
	}
	txIdsPerHash := make(map[string]int64)
	for _, txHashId := range txHashIds {
		txIdsPerHash[txHashId.Hash] = txHashId.Id
	}
	return txIdsPerHash, nil
}

func (a *postgresAccessor) saveSubmittedFlips(ctx *context, flips []Flip) error {
	if len(flips) == 0 {
		return nil
	}
	for _, flip := range flips {
		txId, err := ctx.txId(flip.TxHash)
		if err != nil {
			return err
		}
		if err := a.saveSubmittedFlip(ctx, txId, flip); err != nil {
			return err
		}
	}
	return nil
}

func (a *postgresAccessor) saveSubmittedFlip(ctx *context, txId int64, flip Flip) error {
	_, err := ctx.tx.Exec(a.getQuery(insertSubmittedFlipQuery), flip.Cid, txId, flip.Pair)
	return errors.Wrapf(err, "unable to save flip %s", flip.Cid)
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
	if len(words) == 0 {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(saveFlipsWordsQuery), pq.Array(words))
	return errors.Wrap(err, "unable to save flips words")
}

func (a *postgresAccessor) savePenalties(ctx *context, penalties []Penalty) error {
	for _, penalty := range penalties {
		if err := a.savePenalty(ctx, &penalty); err != nil {
			return errors.Wrapf(err, "unable to save penalties")
		}
	}
	return nil
}

func (a *postgresAccessor) savePenalty(ctx *context, penalty *Penalty) error {
	if penalty == nil {
		return nil
	}
	_, err := ctx.tx.Exec(a.getQuery(insertPenaltyQuery), penalty.Address, penalty.Penalty, penalty.Seconds, penalty.InheritedFrom, ctx.blockHeight)
	return errors.Wrapf(err, "unable to save penalty")
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

func (a *postgresAccessor) Destroy() {
	err := a.db.Close()
	if err != nil {
		log.Error("Unable to close db: %v", err)
	}
}

func getResultError(err error) error {
	if pgErr, ok := errors.Cause(err).(*pq.Error); ok {
		return errors.Wrap(errors.New(pgErrDetails(pgErr)), err.Error())
	}
	return err
}

func pgErrDetails(err *pq.Error) string {
	return fmt.Sprintf("Severity: %v, Code: %v, Message: %v, Detail: %v, Hint: %v, Position: %v, InternalPosition: %v, InternalQuery: %v, Where: %v, Schema: %v, Table: %v, Column: %v, DataTypeName: %v, Constraint: %v, File: %v, Line: %v, Routine: %v", err.Severity,
		err.Code,
		err.Message,
		err.Detail,
		err.Hint,
		err.Position,
		err.InternalPosition,
		err.InternalQuery,
		err.Where,
		err.Schema,
		err.Table,
		err.Column,
		err.DataTypeName,
		err.Constraint,
		err.File,
		err.Line,
		err.Routine,
	)
}
