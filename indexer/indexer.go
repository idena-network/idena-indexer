package indexer

import (
	"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/rlp"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/flip"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/migration/runtime"
	"github.com/idena-network/idena-indexer/monitoring"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	_ "image/png"
	"math/big"
	"time"
)

const (
	requestRetryInterval = time.Second * 5
)

var (
	blockFlags = map[types.BlockFlag]string{
		types.IdentityUpdate:          "IdentityUpdate",
		types.FlipLotteryStarted:      "FlipLotteryStarted",
		types.ShortSessionStarted:     "ShortSessionStarted",
		types.LongSessionStarted:      "LongSessionStarted",
		types.AfterLongSessionStarted: "AfterLongSessionStarted",
		types.ValidationFinished:      "ValidationFinished",
		types.Snapshot:                "Snapshot",
		types.OfflinePropose:          "OfflinePropose",
		types.OfflineCommit:           "OfflineCommit",
	}
)

type Indexer struct {
	listener           incoming.Listener
	memPoolIndexer     *mempool.Indexer
	db                 db.Accessor
	restorer           *restore.Restorer
	state              *indexerState
	secondaryStorage   *runtime.SecondaryStorage
	genesisBlockHeight uint64
	restore            bool
	pm                 monitoring.PerformanceMonitor
	flipLoader         flip.Loader
}

type result struct {
	dbData  *db.Data
	resData *resultData
}

type resultData struct {
	totalBalance *big.Int
	totalStake   *big.Int
	flipTxs      []flipTx
}

type flipTx struct {
	txHash common.Hash
	cid    []byte
}

func NewIndexer(
	listener incoming.Listener,
	mempoolIndexer *mempool.Indexer,
	db db.Accessor,
	restorer *restore.Restorer,
	secondaryStorage *runtime.SecondaryStorage,
	genesisBlockHeight uint64,
	restoreInitially bool,
	pm monitoring.PerformanceMonitor,
	flipLoader flip.Loader,
) *Indexer {
	return &Indexer{
		listener:           listener,
		memPoolIndexer:     mempoolIndexer,
		db:                 db,
		restorer:           restorer,
		secondaryStorage:   secondaryStorage,
		genesisBlockHeight: genesisBlockHeight,
		restore:            restoreInitially,
		pm:                 pm,
		flipLoader:         flipLoader,
	}
}

func (indexer *Indexer) Start() {
	indexer.memPoolIndexer.Initialize(indexer.listener.NodeEventBus())
	indexer.listener.Listen(indexer.indexBlock, indexer.getHeightToIndex()-1)
}

func (indexer *Indexer) WaitForNodeStop() {
	indexer.listener.WaitForStop()
}

func (indexer *Indexer) Destroy() {
	indexer.listener.Destroy()
	indexer.memPoolIndexer.Destroy()
	indexer.db.Destroy()
}

func (indexer *Indexer) statsHolder() stats.StatsHolder {
	return indexer.listener.StatsCollector().(stats.StatsHolder)
}

func (indexer *Indexer) indexBlock(block *types.Block) {
	var genesisBlock *types.Block
	for {
		heightToIndex := indexer.getHeightToIndex()
		if genesisBlock != nil {
			heightToIndex++
		}

		if !indexer.isFirstBlock(block) && block.Height() > heightToIndex {
			panic(fmt.Sprintf("Incoming block height=%d is greater than expected %d", block.Height(), heightToIndex))
		}
		if block.Height() < heightToIndex {
			log.Info(fmt.Sprintf("Incoming block height=%d is less than expected %d, start resetting indexer db...", block.Height(), heightToIndex))
			heightToReset := block.Height() - 1

			if indexer.isGenesis(block.Header.ParentHash()) {
				log.Info(fmt.Sprintf("Block %d is first after genesis", block.Height()))
				heightToReset--
				genesisBlock = indexer.listener.NodeCtx().Blockchain.GetBlock(
					indexer.listener.NodeCtx().Blockchain.Genesis(),
				)
				if genesisBlock == nil {
					log.Error("Unable to get genesis block")
					indexer.waitForRetry()
					continue
				}
			}
			if err := indexer.resetTo(heightToReset); err != nil {
				log.Error(fmt.Sprintf("Unable to reset to height=%d", heightToReset), "err", err)
				indexer.waitForRetry()
			} else {
				log.Info(fmt.Sprintf("Indexer db has been reset to height=%d", heightToReset))
				indexer.restore = indexer.restore || !indexer.isFirstBlockHeight(block.Height())
			}
			// retry in any case to ensure incoming height equals to expected height to index after reset
			continue
		}

		if indexer.restore {
			log.Info("Start restoring DB data...")
			indexer.restorer.Restore()
			log.Info("DB data has been restored")
			indexer.restore = false
		}

		indexer.pm.Start("Convert")
		err := indexer.initializeStateIfNeeded(block)
		if err != nil {
			panic(err)
		}

		if genesisBlock != nil {
			res, err := indexer.convertIncomingData(genesisBlock)
			if err != nil {
				panic(err)
			}
			indexer.saveData(res.dbData)
			log.Info(fmt.Sprintf("Processed genesis block %d", genesisBlock.Height()))
			genesisBlock = nil
		}

		res, err := indexer.convertIncomingData(block)
		if err != nil {
			panic(err)
		}
		indexer.pm.Complete("Convert")

		indexer.pm.Start("Save")
		indexer.saveData(res.dbData)
		indexer.pm.Complete("Save")

		indexer.pm.Start("Flips")
		indexer.loadFlips(res.resData.flipTxs)
		indexer.pm.Complete("Flips")

		indexer.applyOnState(res)

		if indexer.secondaryStorage != nil && block.Height() >= indexer.secondaryStorage.GetLastBlockHeight() {
			indexer.secondaryStorage.Destroy()
			indexer.secondaryStorage = nil
			log.Info("Completed runtime migration")
		}

		log.Info(fmt.Sprintf("Processed block %d", block.Height()))

		return
	}
}

func (indexer *Indexer) isGenesis(hash common.Hash) bool {
	return hash == indexer.listener.NodeCtx().Blockchain.Genesis()
}

func (indexer *Indexer) resetTo(height uint64) error {
	err := indexer.db.ResetTo(height)
	if err != nil {
		return err
	}
	indexer.state = indexer.loadState()
	return nil
}

func (indexer *Indexer) getHeightToIndex() uint64 {
	if indexer.state == nil {
		indexer.state = indexer.loadState()
	}
	return indexer.state.lastHeight + 1
}

func (indexer *Indexer) loadState() *indexerState {
	for {
		lastHeight, err := indexer.db.GetLastHeight()
		if err != nil {
			log.Error(fmt.Sprintf("Unable to get last saved height: %v", err))
			indexer.waitForRetry()
			continue
		}
		return &indexerState{
			lastHeight: lastHeight,
		}
	}
}

func (indexer *Indexer) initializeStateIfNeeded(block *types.Block) error {
	if indexer.state.totalStake != nil && indexer.state.totalBalance != nil {
		return nil
	}
	prevState, err := indexer.listener.AppStateReadonly(block.Height() - 1)
	if err != nil {
		return err
	}
	totalBalance := big.NewInt(0)
	totalStake := big.NewInt(0)
	prevState.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		addr := conversion.BytesToAddr(key)
		totalBalance.Add(totalBalance, prevState.State.GetBalance(addr))
		totalStake.Add(totalStake, prevState.State.GetStakeBalance(addr))
		return false
	})
	indexer.state.totalBalance = totalBalance
	indexer.state.totalStake = totalStake
	return nil
}

func (indexer *Indexer) convertIncomingData(incomingBlock *types.Block) (*result, error) {
	indexer.pm.Start("InitCtx")
	isGenesisBlock := incomingBlock.Hash() == indexer.listener.NodeCtx().Blockchain.Genesis()
	var prevState *appstate.AppState
	var err error
	if isGenesisBlock {
		prevState, err = indexer.listener.AppStateReadonly(incomingBlock.Height())
	} else {
		prevState, err = indexer.listener.AppStateReadonly(incomingBlock.Height() - 1)
	}
	if err != nil {
		return nil, err
	}
	newState, err := indexer.listener.AppStateReadonly(incomingBlock.Height())
	if err != nil {
		return nil, err
	}
	ctx := &conversionContext{
		blockHeight:       incomingBlock.Height(),
		prevStateReadOnly: prevState,
		newStateReadOnly:  newState,
	}
	collector := &conversionCollector{
		addresses: make(map[string]*db.Address),
	}
	collectorStats := indexer.statsHolder().GetStats()
	epoch := uint64(prevState.State.Epoch())

	indexer.pm.Complete("InitCtx")
	indexer.pm.Start("ConvertBlock")
	block, err := indexer.convertBlock(incomingBlock, ctx, collector)
	if err != nil {
		return nil, err
	}
	indexer.pm.Complete("ConvertBlock")
	epochResult := indexer.detectEpochResult(incomingBlock, ctx)

	firstAddresses := indexer.detectFirstAddresses(incomingBlock, ctx)
	for _, addr := range firstAddresses {
		if curAddr, present := collector.addresses[addr.Address]; present {
			curAddr.StateChanges = append(curAddr.StateChanges, addr.StateChanges...)
		} else {
			collector.addresses[addr.Address] = addr
		}
	}

	balanceUpdates, diff := determineBalanceUpdates(indexer.isFirstBlock(incomingBlock),
		collectorStats.BalanceUpdateAddrs,
		ctx.prevStateReadOnly,
		ctx.newStateReadOnly)

	coins, totalBalance, totalStake := indexer.getCoins(indexer.isFirstBlock(incomingBlock), diff)

	dbData := &db.Data{
		Epoch:                  epoch,
		ValidationTime:         *big.NewInt(ctx.newStateReadOnly.State.NextValidationTime().Unix()),
		Block:                  block,
		ActivationTxTransfers:  collectorStats.ActivationTxTransfers,
		KillTxTransfers:        collectorStats.KillTxTransfers,
		KillInviteeTxTransfers: collectorStats.KillInviteeTxTransfers,
		ActivationTxs:          collector.activationTxs,
		KillInviteeTxs:         collector.killInviteeTxs,
		BecomeOnlineTxs:        collector.becomeOnlineTxs,
		BecomeOfflineTxs:       collector.becomeOfflineTxs,
		SubmittedFlips:         collector.submittedFlips,
		DeletedFlips:           collector.deletedFlips,
		FlipKeys:               collector.flipKeys,
		FlipsWords:             collector.flipsWords,
		Addresses:              collector.getAddresses(),
		ChangedBalances:        balanceUpdates,
		Coins:                  coins,
		Penalty:                detectChargedPenalty(incomingBlock, ctx.newStateReadOnly),
		BurntPenalties:         convertBurntPenalties(collectorStats.BurntPenaltiesByAddr),
		MiningRewards:          collectorStats.MiningRewards,
		BurntCoinsPerAddr:      collectorStats.BurntCoinsByAddr,
		BalanceUpdates:         collectorStats.BalanceUpdates,
		CommitteeRewardShare:   collectorStats.CommitteeRewardShare,
		EpochResult:            epochResult,
	}
	resData := &resultData{
		totalBalance: totalBalance,
		totalStake:   totalStake,
		flipTxs:      collector.flipTxs,
	}
	return &result{
		dbData:  dbData,
		resData: resData,
	}, nil
}

func (indexer *Indexer) getCoins(
	isFirstBlock bool,
	diff balanceDiff,
) (dbCoins db.Coins, totalBalance, totalStake *big.Int) {

	minted := indexer.statsHolder().GetStats().MintedCoins
	// Genesis minted coins
	if isFirstBlock {
		if minted == nil {
			minted = big.NewInt(0)
		}
		minted.Add(minted, indexer.state.totalBalance)
		minted.Add(minted, indexer.state.totalStake)
	}
	totalBalance = new(big.Int).Add(indexer.state.totalBalance, diff.balance)
	totalStake = new(big.Int).Add(indexer.state.totalStake, diff.stake)
	dbCoins = db.Coins{
		Minted:       blockchain.ConvertToFloat(minted),
		Burnt:        blockchain.ConvertToFloat(indexer.statsHolder().GetStats().BurntCoins),
		TotalBalance: blockchain.ConvertToFloat(totalBalance),
		TotalStake:   blockchain.ConvertToFloat(totalStake),
	}
	return
}

func (indexer *Indexer) isFirstBlock(incomingBlock *types.Block) bool {
	return indexer.isFirstBlockHeight(incomingBlock.Height())
}

func (indexer *Indexer) isFirstBlockHeight(height uint64) bool {
	return height == indexer.genesisBlockHeight+1
}

func (indexer *Indexer) detectFirstAddresses(incomingBlock *types.Block, ctx *conversionContext) []*db.Address {
	if !indexer.isFirstBlock(incomingBlock) {
		return nil
	}
	var addresses []*db.Address
	var withZeroWallet bool
	ctx.newStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		if !withZeroWallet && addr == (common.Address{}) {
			withZeroWallet = true
		}
		addresses = append(addresses, &db.Address{
			Address: conversion.ConvertAddress(addr),
			StateChanges: []db.AddressStateChange{
				{
					PrevState: convertIdentityState(ctx.prevStateReadOnly.State.GetIdentityState(addr)),
					NewState:  convertIdentityState(identity.State),
				},
			},
		})
	})
	if !withZeroWallet {
		addresses = append(addresses, &db.Address{
			Address: conversion.ConvertAddress(common.Address{}),
			StateChanges: []db.AddressStateChange{
				{
					PrevState: convertIdentityState(state.Undefined),
					NewState:  convertIdentityState(state.Undefined),
				},
			},
		})
	}
	return addresses
}

func (indexer *Indexer) convertBlock(
	incomingBlock *types.Block,
	ctx *conversionContext,
	collector *conversionCollector,
) (db.Block, error) {
	var txs []db.Transaction
	if len(incomingBlock.Body.Transactions) > 0 {
		stateToApply, err := ctx.newStateReadOnly.Readonly(ctx.blockHeight - 1)
		if err != nil {
			return db.Block{}, err
		}
		txs = indexer.convertTransactions(incomingBlock.Body.Transactions, ctx, stateToApply, collector)
	}

	incomingBlock.Header.Flags()
	proposerVrfScore, _ := getProposerVrfScore(
		incomingBlock,
		indexer.listener.NodeCtx().ProofsByRound,
		indexer.listener.NodeCtx().PendingProofs,
		indexer.secondaryStorage,
	)
	encodedBlock, _ := rlp.EncodeToBytes(incomingBlock)
	return db.Block{
		Height:               incomingBlock.Height(),
		Hash:                 conversion.ConvertHash(incomingBlock.Hash()),
		Time:                 *incomingBlock.Header.Time(),
		Transactions:         txs,
		Proposer:             getProposer(incomingBlock),
		Flags:                convertFlags(incomingBlock.Header.Flags()),
		IsEmpty:              incomingBlock.IsEmpty(),
		BodySize:             len(incomingBlock.Body.Bytes()),
		FullSize:             len(encodedBlock),
		ValidatorsCount:      len(indexer.statsHolder().GetStats().FinalCommittee),
		VrfProposerThreshold: ctx.prevStateReadOnly.State.VrfProposerThreshold(),
		ProposerVrfScore:     proposerVrfScore,
		FeeRate:              blockchain.ConvertToFloat(ctx.prevStateReadOnly.State.FeePerByte()),
	}, nil
}

func convertFlags(incomingFlags types.BlockFlag) []string {
	var flags []string
	for incomingFlag, flag := range blockFlags {
		if incomingFlags.HasFlag(incomingFlag) {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (indexer *Indexer) convertTransactions(
	incomingTxs []*types.Transaction,
	ctx *conversionContext,
	stateToApply *appstate.AppState,
	collector *conversionCollector,
) []db.Transaction {
	if len(incomingTxs) == 0 {
		return nil
	}
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, indexer.convertTransaction(incomingTx, ctx, stateToApply, collector))
	}
	return txs
}

func (indexer *Indexer) convertTransaction(
	incomingTx *types.Transaction,
	ctx *conversionContext,
	stateToApply *appstate.AppState,
	collector *conversionCollector,
) db.Transaction {
	if f, h := detectSubmittedFlip(incomingTx); f != nil {
		collector.submittedFlips = append(collector.submittedFlips, *f)
		collector.flipTxs = append(collector.flipTxs, *h)
	}

	if deletedFlip := detectDeletedFlip(incomingTx); deletedFlip != nil {
		collector.deletedFlips = append(collector.deletedFlips, *deletedFlip)
	}

	if activationTx := detectActivationTx(incomingTx, ctx.prevStateReadOnly); activationTx != nil {
		collector.activationTxs = append(collector.activationTxs, *activationTx)
	}

	if killInviteeTx := detectKillInviteeTx(incomingTx, ctx.prevStateReadOnly); killInviteeTx != nil {
		collector.killInviteeTxs = append(collector.killInviteeTxs, *killInviteeTx)
	}

	if becomeOnlineTxHash, becomeOfflineTxHash := detectOnlineStatusTx(incomingTx); becomeOnlineTxHash != nil {
		collector.becomeOnlineTxs = append(collector.becomeOnlineTxs, *becomeOnlineTxHash)
	} else if becomeOfflineTxHash != nil {
		collector.becomeOfflineTxs = append(collector.becomeOfflineTxs, *becomeOfflineTxHash)
	}

	indexer.convertShortAnswers(incomingTx, ctx, collector)
	txHash := conversion.ConvertHash(incomingTx.Hash())

	sender, _ := types.Sender(incomingTx)
	from := conversion.ConvertAddress(sender)
	if _, present := collector.addresses[from]; !present {
		collector.addresses[from] = &db.Address{
			Address: from,
		}
	}

	var to string
	var recipientPrevState *state.IdentityState
	if incomingTx.To != nil {
		to = conversion.ConvertAddress(*incomingTx.To)
		if _, present := collector.addresses[to]; !present {
			collector.addresses[to] = &db.Address{
				Address: to,
			}
		}
		st := stateToApply.State.GetIdentityState(*incomingTx.To)
		recipientPrevState = &st
	}

	senderPrevState := stateToApply.State.GetIdentityState(sender)
	fee, err := indexer.listener.Blockchain().ApplyTxOnState(stateToApply, incomingTx, nil)
	if err != nil {
		log.Error("Unable to apply tx on state", "tx", txHash, "err", err)
	}

	senderNewState := stateToApply.State.GetIdentityState(sender)

	if senderNewState != senderPrevState {
		if incomingTx.Type == types.ActivationTx && senderNewState == state.Killed {
			collector.addresses[from].IsTemporary = true
		}
		collector.addresses[from].StateChanges = append(collector.addresses[from].StateChanges,
			db.AddressStateChange{
				PrevState: convertIdentityState(senderPrevState),
				NewState:  convertIdentityState(senderNewState),
				TxHash:    txHash,
			})
	}
	if recipientPrevState != nil && *incomingTx.To != sender {
		recipientNewState := stateToApply.State.GetIdentityState(*incomingTx.To)
		if recipientNewState != *recipientPrevState {
			collector.addresses[to].StateChanges = append(collector.addresses[to].StateChanges,
				db.AddressStateChange{
					PrevState: convertIdentityState(*recipientPrevState),
					NewState:  convertIdentityState(recipientNewState),
					TxHash:    txHash,
				})
		}
	}

	tx := db.Transaction{
		Type:    convertTxType(incomingTx.Type),
		Payload: incomingTx.Payload,
		Hash:    txHash,
		From:    from,
		To:      to,
		Amount:  blockchain.ConvertToFloat(incomingTx.Amount),
		Tips:    blockchain.ConvertToFloat(incomingTx.Tips),
		MaxFee:  blockchain.ConvertToFloat(incomingTx.MaxFee),
		Fee:     blockchain.ConvertToFloat(fee),
		Size:    incomingTx.Size(),
	}

	return tx
}

func detectActivationTx(tx *types.Transaction, prevState *appstate.AppState) *db.ActivationTx {
	if tx.Type != types.ActivationTx {
		return nil
	}
	sender, _ := types.Sender(tx)
	inviter := prevState.State.GetInviter(sender)
	return &db.ActivationTx{
		TxHash:       conversion.ConvertHash(tx.Hash()),
		InviteTxHash: conversion.ConvertHash(inviter.TxHash),
	}
}

func detectKillInviteeTx(tx *types.Transaction, prevState *appstate.AppState) *db.KillInviteeTx {
	if tx.Type != types.KillInviteeTx {
		return nil
	}
	inviter := prevState.State.GetInviter(*tx.To)
	return &db.KillInviteeTx{
		TxHash:       conversion.ConvertHash(tx.Hash()),
		InviteTxHash: conversion.ConvertHash(inviter.TxHash),
	}
}

func detectOnlineStatusTx(tx *types.Transaction) (becomeOnlineTxHash, becomeOfflineTxHash *string) {
	if tx.Type != types.OnlineStatusTx {
		return
	}
	attachment := attachments.ParseOnlineStatusAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse online status payload. Skipped.", "tx", tx.Hash())
		return
	}
	h := conversion.ConvertHash(tx.Hash())
	if attachment.Online {
		becomeOnlineTxHash = &h
	} else {
		becomeOfflineTxHash = &h
	}
	return
}

func convertTxType(txType types.TxType) uint16 {
	return txType
}

func convertIdentityState(state state.IdentityState) uint8 {
	return uint8(state)
}

func convertFlipStatus(status ceremony.FlipStatus) byte {
	return byte(status)
}

func convertAnswer(answer types.Answer) byte {
	return byte(answer)
}

func convertStatsAnswers(incomingAnswers []statsTypes.FlipAnswerStats) []db.Answer {
	var answers []db.Answer
	for _, answer := range incomingAnswers {
		answers = append(answers, convertStatsAnswer(answer))
	}
	return answers
}

func convertStatsAnswer(incomingAnswer statsTypes.FlipAnswerStats) db.Answer {
	return db.Answer{
		Address:    conversion.ConvertAddress(incomingAnswer.Respondent),
		Answer:     convertAnswer(incomingAnswer.Answer),
		WrongWords: incomingAnswer.WrongWords,
		Point:      incomingAnswer.Point,
	}
}

func convertCid(cid cid.Cid) string {
	return cid.String()
}

func (indexer *Indexer) detectEpochResult(block *types.Block, ctx *conversionContext) *db.EpochResult {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil
	}

	var birthdays []db.Birthday
	var identities []db.EpochIdentity
	memPoolFlipKeysToMigrate := indexer.getMemPoolFlipKeysToMigrate(ctx.prevStateReadOnly.State.Epoch())
	memPoolFlipKeys := memPoolFlipKeysToMigrate
	validationStats := indexer.statsHolder().GetStats().ValidationStats

	authorAddressesByFlipCid := make(map[string]string)

	ctx.prevStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		convertedAddress := conversion.ConvertAddress(addr)
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.Address = convertedAddress
		convertedIdentity.State = convertIdentityState(ctx.newStateReadOnly.State.GetIdentityState(addr))
		convertedIdentity.TotalShortPoint = ctx.prevStateReadOnly.State.GetShortFlipPoints(addr)
		convertedIdentity.TotalShortFlips = ctx.prevStateReadOnly.State.GetQualifiedFlipsCount(addr)
		convertedIdentity.RequiredFlips = ctx.prevStateReadOnly.State.GetRequiredFlips(addr)
		identityPrevState := ctx.prevStateReadOnly.State.GetIdentity(addr)
		convertedIdentity.AvailableFlips = identityPrevState.GetMaximumAvailableFlips()
		convertedIdentity.MadeFlips = ctx.prevStateReadOnly.State.GetMadeFlips(addr)
		convertedIdentity.NextEpochInvites = ctx.newStateReadOnly.State.GetInvites(addr)
		if identityStats, present := validationStats.IdentitiesPerAddr[addr]; present {
			convertedIdentity.ShortPoint = identityStats.ShortPoint
			convertedIdentity.TotalShortPoint += identityStats.ShortPoint
			convertedIdentity.ShortFlips = identityStats.ShortFlips
			convertedIdentity.TotalShortFlips += identityStats.ShortFlips
			convertedIdentity.LongPoint = identityStats.LongPoint
			convertedIdentity.LongFlips = identityStats.LongFlips
			convertedIdentity.Approved = identityStats.Approved
			convertedIdentity.Missed = identityStats.Missed
			convertedIdentity.ShortFlipCidsToSolve = convertCids(identityStats.ShortFlipsToSolve, validationStats.FlipCids, block)
			convertedIdentity.LongFlipCidsToSolve = convertCids(identityStats.LongFlipsToSolve, validationStats.FlipCids, block)
		} else {
			convertedIdentity.Approved = false
			convertedIdentity.Missed = true
		}

		if identityPrevState.State == state.Invite || identityPrevState.State == state.Candidate {
			convertedIdentity.BirthEpoch = uint64(ctx.prevStateReadOnly.State.Epoch())
		} else {
			convertedIdentity.BirthEpoch = uint64(identityPrevState.Birthday)
		}

		identities = append(identities, convertedIdentity)

		birthday := detectBirthday(addr, identity.Birthday, ctx.newStateReadOnly.State.GetIdentity(addr).Birthday)
		if birthday != nil {
			birthdays = append(birthdays, *birthday)
		}

		if memPoolFlipKeysToMigrate == nil {
			memPoolFlipKey := indexer.detectMemPoolFlipKey(addr, identity)
			if memPoolFlipKey != nil {
				memPoolFlipKeys = append(memPoolFlipKeys, memPoolFlipKey)
			}
		}

		for _, identityFlip := range identity.Flips {
			flipCid, err := cid.Parse(identityFlip.Cid)
			if err != nil {
				log.Error(fmt.Sprintf("Unable to parse flip cid %v", identityFlip.Cid))
				continue
			}
			authorAddressesByFlipCid[convertCid(flipCid)] = convertedAddress
		}
	})

	var flipsStats []db.FlipStats
	for flipIdx, flipStats := range validationStats.FlipsPerIdx {
		flipCid, err := cid.Parse(validationStats.FlipCids[flipIdx])
		if err != nil {
			log.Error("Unable to parse flip cid. Skipped.", "b", block.Height(), "idx", flipIdx, "err", err)
			continue
		}
		flipCidStr := convertCid(flipCid)
		flipStats := db.FlipStats{
			Author:       authorAddressesByFlipCid[flipCidStr],
			Cid:          flipCidStr,
			ShortAnswers: convertStatsAnswers(flipStats.ShortAnswers),
			LongAnswers:  convertStatsAnswers(flipStats.LongAnswers),
			Status:       convertFlipStatus(ceremony.FlipStatus(flipStats.Status)),
			Answer:       convertAnswer(flipStats.Answer),
			WrongWords:   flipStats.WrongWords,
		}
		flipsStats = append(flipsStats, flipStats)
	}

	collectorStats := indexer.statsHolder().GetStats()
	var minScoreForInvite float32 = 0
	if collectorStats.MinScoreForInvite != nil {
		minScoreForInvite = *collectorStats.MinScoreForInvite
	}

	return &db.EpochResult{
		Identities:        identities,
		FlipStats:         flipsStats,
		Birthdays:         birthdays,
		MemPoolFlipKeys:   memPoolFlipKeys,
		FailedValidation:  validationStats.Failed,
		EpochRewards:      indexer.detectEpochRewards(block),
		MinScoreForInvite: minScoreForInvite,
	}
}

func (indexer *Indexer) detectMemPoolFlipKey(addr common.Address, identity state.Identity) *db.MemPoolFlipKey {
	if len(identity.Flips) == 0 {
		return nil
	}
	key := indexer.listener.KeysPool().GetPublicFlipKey(addr)
	if key == nil {
		log.Warn(fmt.Sprintf("Not found mem pool flip key for %s", addr.Hex()))
		return nil
	}
	return &db.MemPoolFlipKey{
		Address: conversion.ConvertAddress(addr),
		Key:     hex.EncodeToString(crypto.FromECDSA(key.ExportECDSA())),
	}

}

func (indexer *Indexer) getMemPoolFlipKeysToMigrate(epoch uint16) []*db.MemPoolFlipKey {
	if indexer.secondaryStorage == nil {
		return nil
	}
	keys, err := indexer.secondaryStorage.GetMemPoolFlipKeys(epoch)
	if err != nil {
		log.Error(errors.Wrap(err, "Unable to get mem pool flip keys to migrate").Error())
		return nil
	}
	return keys
}

func detectBirthday(address common.Address, prevBirthday, newBirthday uint16) *db.Birthday {
	if prevBirthday == newBirthday {
		return nil
	}
	return &db.Birthday{
		Address:    conversion.ConvertAddress(address),
		BirthEpoch: uint64(newBirthday),
	}
}

func detectSubmittedFlip(tx *types.Transaction) (*db.Flip, *flipTx) {
	if tx.Type != types.SubmitFlipTx {
		return nil, nil
	}
	attachment := attachments.ParseFlipSubmitAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse submitted flip payload. Skipped.", "tx", tx.Hash())
		return nil, nil
	}
	flipCid, err := cid.Parse(attachment.Cid)
	if err != nil {
		log.Error("Unable to parse flip cid. Skipped.", "tx", tx.Hash(), "err", err)
		return nil, nil
	}
	f := &db.Flip{
		TxHash: conversion.ConvertHash(tx.Hash()),
		Cid:    convertCid(flipCid),
		Pair:   attachment.Pair,
	}
	h := &flipTx{
		txHash: tx.Hash(),
		cid:    attachment.Cid,
	}
	return f, h
}

func detectDeletedFlip(tx *types.Transaction) *db.DeletedFlip {
	if tx.Type != types.DeleteFlipTx {
		return nil
	}
	attachment := attachments.ParseDeleteFlipAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse delete flip tx payload. Skipped.", "tx", tx.Hash())
		return nil
	}
	flipCid, err := cid.Parse(attachment.Cid)
	if err != nil {
		log.Error("Unable to parse deleted flip cid. Skipped.", "tx", tx.Hash(), "err", err)
		return nil
	}
	return &db.DeletedFlip{
		TxHash: conversion.ConvertHash(tx.Hash()),
		Cid:    convertCid(flipCid),
	}
}

func (indexer *Indexer) convertShortAnswers(
	tx *types.Transaction,
	ctx *conversionContext,
	collector *conversionCollector,
) {
	if tx.Type != types.SubmitShortAnswersTx {
		return
	}
	attachment := attachments.ParseShortAnswerAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse short answers payload. Skipped.", "tx", tx.Hash())
		return
	}

	sender, _ := types.Sender(tx)
	from := conversion.ConvertAddress(sender)
	senderFlips, err := indexer.db.GetCurrentFlips(from)
	if err != nil {
		log.Error("Unable to get current flips. Skipped.", "err", err, "tx", tx.Hash())
		return
	}

	if len(attachment.Key) > 0 {
		collector.flipKeys = append(collector.flipKeys, db.FlipKey{
			TxHash: conversion.ConvertHash(tx.Hash()),
			Key:    hex.EncodeToString(attachment.Key),
		})
	}

	if len(attachment.Proof) > 0 {
		for _, f := range senderFlips {
			word1, word2, err := getFlipWords(sender, attachment, int(f.Pair), ctx.prevStateReadOnly)
			if err != nil {
				log.Error("Unable to get flip words. Skipped.", "tx", tx.Hash(), "cid", f.Cid, "err", err)
				continue
			}
			collector.flipsWords = append(collector.flipsWords, db.FlipWords{
				FlipTxId: f.TxId,
				TxHash:   conversion.ConvertHash(tx.Hash()),
				Word1:    uint16(word1),
				Word2:    uint16(word2),
			})
		}
	} else {
		log.Error("Empty proof for flip words. Skipped.", "tx", tx.Hash())
	}
}

func getFlipWords(addr common.Address, attachment *attachments.ShortAnswerAttachment, pairId int, appState *appstate.AppState) (word1, word2 int, err error) {
	seed := appState.State.FlipWordsSeed().Bytes()
	proof := attachment.Proof
	identity := appState.State.GetIdentity(addr)
	return ceremony.GetWords(seed, proof, identity.PubKey, common.WordDictionarySize, identity.GetTotalWordPairsCount(), pairId, appState.State.Epoch())
}

func convertCids(idxs []int, cids [][]byte, block *types.Block) []string {
	var res []string
	for _, idx := range idxs {
		if idx >= len(cids) {
			log.Error("Unable to get flip cid by idx. Skipped.", "b", block.Height(), "idx", idx)
			continue
		}
		c, err := cid.Parse(cids[idx])
		if err != nil {
			log.Error("Unable to parse cid. Skipped.", "b", block.Height(), "idx", idx, "err", err)
			continue
		}
		res = append(res, convertCid(c))
	}
	return res
}

func (indexer *Indexer) loadFlips(flipTxs []flipTx) {
	for _, flipTx := range flipTxs {
		indexer.flipLoader.SubmitToLoad(flipTx.cid, flipTx.txHash)
	}
}

func (indexer *Indexer) saveData(data *db.Data) {
	for {
		if err := indexer.db.Save(data); err != nil {
			log.Error(fmt.Sprintf("Unable to save incoming data: %v", err))
			indexer.waitForRetry()
			continue
		}
		return
	}
}

func (indexer *Indexer) applyOnState(data *result) {
	indexer.state.lastHeight = data.dbData.Block.Height
	indexer.state.totalBalance = data.resData.totalBalance
	indexer.state.totalStake = data.resData.totalStake
}

func (indexer *Indexer) waitForRetry() {
	time.Sleep(requestRetryInterval)
}
