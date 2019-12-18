package indexer

import (
	"bufio"
	"bytes"
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
	"github.com/idena-network/idena-go/crypto/ecies"
	"github.com/idena-network/idena-go/rlp"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
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
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	_ "image/png"
	"math/big"
	"time"
)

const (
	requestRetryInterval      = time.Second * 5
	flipLimitToGetMemPoolData = 500
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
}

type result struct {
	dbData  *db.Data
	resData *resultData
}

type resultData struct {
	totalBalance *big.Int
	totalStake   *big.Int
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
	for {
		heightToIndex := indexer.getHeightToIndex()

		if !indexer.isFirstBlock(block) && block.Height() > heightToIndex {
			panic(fmt.Sprintf("Incoming block height=%d is greater than expected %d", block.Height(), heightToIndex))
		}

		if block.Height() < heightToIndex {
			height := block.Height() - 1
			log.Info(fmt.Sprintf("Incoming block height=%d is less than expected %d, start resetting indexer db...", block.Height(), heightToIndex))
			if err := indexer.resetTo(height); err != nil {
				log.Error(fmt.Sprintf("Unable to reset to height=%d", height), "err", err)
				indexer.waitForRetry()
			} else {
				log.Info(fmt.Sprintf("Indexer db has been reset to height=%d", height))
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
		indexer.initializeStateIfNeeded(block)
		res := indexer.convertIncomingData(block)
		indexer.pm.Complete("Convert")
		indexer.pm.Start("Save")
		indexer.saveData(res.dbData)
		indexer.pm.Complete("Save")
		indexer.applyOnState(res)

		if indexer.secondaryStorage != nil && block.Height() >= indexer.secondaryStorage.GetLastBlockHeight() {
			indexer.secondaryStorage.Destroy()
			indexer.secondaryStorage = nil
			log.Info("Completed runtime migration")
		}

		if block.Height()%1000 == 0 {
			log.Info(fmt.Sprintf("Processed block %d", block.Height()))
		} else {
			log.Debug(fmt.Sprintf("Processed block %d", block.Height()))
		}

		return
	}
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

func (indexer *Indexer) initializeStateIfNeeded(block *types.Block) {
	if indexer.state.totalStake != nil && indexer.state.totalBalance != nil {
		return
	}
	prevState := indexer.listener.AppStateReadonly(block.Height() - 1)
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
}

func (indexer *Indexer) convertIncomingData(incomingBlock *types.Block) *result {
	indexer.pm.Start("InitCtx")
	prevState := indexer.listener.AppStateReadonly(incomingBlock.Height() - 1)
	newState := indexer.listener.AppStateReadonly(incomingBlock.Height())
	ctx := &conversionContext{
		blockHeight:       incomingBlock.Height(),
		prevStateReadOnly: prevState,
		newStateReadOnly:  newState,
		addresses:         make(map[string]*db.Address),
	}
	epoch := uint64(prevState.State.Epoch())

	indexer.pm.Complete("InitCtx")
	indexer.pm.Start("ConvertBlock")
	block := indexer.convertBlock(incomingBlock, ctx)
	indexer.pm.Complete("ConvertBlock")
	identities, flipStats, flipsMemPoolData, birthdays, notFailedValidation := indexer.detectEpochResult(incomingBlock, ctx)

	firstAddresses := indexer.detectFirstAddresses(incomingBlock, ctx)
	for _, addr := range firstAddresses {
		if curAddr, present := ctx.addresses[addr.Address]; present {
			curAddr.StateChanges = append(curAddr.StateChanges, addr.StateChanges...)
		} else {
			ctx.addresses[addr.Address] = addr
		}
	}

	balanceUpdates, diff := determineBalanceUpdates(indexer.isFirstBlock(incomingBlock),
		indexer.statsHolder().GetStats().BalanceUpdateAddrs,
		ctx.prevStateReadOnly,
		ctx.newStateReadOnly)

	coins, totalBalance, totalStake := indexer.getCoins(ctx, indexer.isFirstBlock(incomingBlock), diff)

	dbData := &db.Data{
		Epoch:             epoch,
		ValidationTime:    *big.NewInt(ctx.newStateReadOnly.State.NextValidationTime().Unix()),
		Block:             block,
		Identities:        identities,
		SubmittedFlips:    ctx.submittedFlips,
		FlipKeys:          ctx.flipKeys,
		FlipsWords:        ctx.flipsWords,
		FlipsData:         append(ctx.flipsData, flipsMemPoolData...),
		FlipSizeUpdates:   ctx.flipSizeUpdates,
		FlipStats:         flipStats,
		Addresses:         ctx.getAddresses(),
		BalanceUpdates:    balanceUpdates,
		Birthdays:         birthdays,
		Coins:             coins,
		SaveEpochSummary:  incomingBlock.Header.Flags().HasFlag(types.ValidationFinished),
		Penalty:           detectChargedPenalty(incomingBlock, ctx.newStateReadOnly),
		BurntPenalties:    convertBurntPenalties(indexer.statsHolder().GetStats().BurntPenaltiesByAddr),
		EpochRewards:      indexer.detectEpochRewards(incomingBlock),
		MiningRewards:     indexer.statsHolder().GetStats().MiningRewards,
		BurntCoinsPerAddr: indexer.statsHolder().GetStats().BurntCoinsByAddr,
		FailedValidation:  !notFailedValidation,
	}
	resData := &resultData{
		totalBalance: totalBalance,
		totalStake:   totalStake,
	}
	return &result{
		dbData:  dbData,
		resData: resData,
	}
}

func (indexer *Indexer) getCoins(
	ctx *conversionContext,
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

func (indexer *Indexer) convertBlock(incomingBlock *types.Block, ctx *conversionContext) db.Block {
	stateToApply := ctx.newStateReadOnly.Readonly(ctx.blockHeight - 1)
	txs := indexer.convertTransactions(incomingBlock.Body.Transactions, stateToApply, ctx)

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
		Hash:                 convertHash(incomingBlock.Hash()),
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
	}
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

func (indexer *Indexer) convertTransactions(incomingTxs []*types.Transaction, stateToApply *appstate.AppState, ctx *conversionContext) []db.Transaction {
	if len(incomingTxs) == 0 {
		return nil
	}
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, indexer.convertTransaction(incomingTx, ctx, stateToApply))
	}
	return txs
}

func (indexer *Indexer) convertTransaction(incomingTx *types.Transaction, ctx *conversionContext, stateToApply *appstate.AppState) db.Transaction {
	if f := indexer.detectSubmittedFlip(incomingTx, ctx); f != nil {
		ctx.submittedFlips = append(ctx.submittedFlips, *f)
	}

	indexer.convertShortAnswers(incomingTx, ctx)
	txHash := convertHash(incomingTx.Hash())

	sender, _ := types.Sender(incomingTx)
	from := conversion.ConvertAddress(sender)
	if _, present := ctx.addresses[from]; !present {
		ctx.addresses[from] = &db.Address{
			Address: from,
		}
	}

	var to string
	var recipientPrevState *state.IdentityState
	if incomingTx.To != nil {
		to = conversion.ConvertAddress(*incomingTx.To)
		if _, present := ctx.addresses[to]; !present {
			ctx.addresses[to] = &db.Address{
				Address: to,
			}
		}
		st := stateToApply.State.GetIdentityState(*incomingTx.To)
		recipientPrevState = &st
	}

	senderPrevState := stateToApply.State.GetIdentityState(sender)
	fee, err := indexer.listener.Blockchain().ApplyTxOnState(stateToApply, incomingTx, nil)
	if err != nil {
		log.Error("Unable to apply tx on state", "tx", incomingTx.Hash(), "err", err)
	}

	senderNewState := stateToApply.State.GetIdentityState(sender)

	if senderNewState != senderPrevState {
		if incomingTx.Type == types.ActivationTx && senderNewState == state.Killed {
			ctx.addresses[from].IsTemporary = true
		}
		ctx.addresses[from].StateChanges = append(ctx.addresses[from].StateChanges, db.AddressStateChange{
			PrevState: convertIdentityState(senderPrevState),
			NewState:  convertIdentityState(senderNewState),
			TxHash:    txHash,
		})
	}
	if recipientPrevState != nil && *incomingTx.To != sender {
		recipientNewState := stateToApply.State.GetIdentityState(*incomingTx.To)
		if recipientNewState != *recipientPrevState {
			ctx.addresses[to].StateChanges = append(ctx.addresses[to].StateChanges, db.AddressStateChange{
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

func convertHash(hash common.Hash) string {
	return hash.Hex()
}

func convertCid(cid cid.Cid) string {
	return cid.String()
}

func (indexer *Indexer) detectEpochResult(block *types.Block, ctx *conversionContext) ([]db.EpochIdentity,
	[]db.FlipStats, []db.FlipData, []db.Birthday, bool) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil, nil, nil, true
	}

	var birthdays []db.Birthday
	var identities []db.EpochIdentity
	validationStats := indexer.statsHolder().GetStats().ValidationStats

	ctx.prevStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.Address = conversion.ConvertAddress(addr)
		convertedIdentity.State = convertIdentityState(ctx.newStateReadOnly.State.GetIdentityState(addr))
		convertedIdentity.TotalShortPoint = ctx.prevStateReadOnly.State.GetShortFlipPoints(addr)
		convertedIdentity.TotalShortFlips = ctx.prevStateReadOnly.State.GetQualifiedFlipsCount(addr)
		convertedIdentity.RequiredFlips = ctx.prevStateReadOnly.State.GetRequiredFlips(addr)
		convertedIdentity.MadeFlips = ctx.prevStateReadOnly.State.GetMadeFlips(addr)
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
		identities = append(identities, convertedIdentity)

		birthday := detectBirthday(addr, identity.Birthday, ctx.newStateReadOnly.State.GetIdentity(addr).Birthday)
		if birthday != nil {
			birthdays = append(birthdays, *birthday)
		}
	})

	var flipsStats []db.FlipStats
	for flipIdx, flipStats := range validationStats.FlipsPerIdx {
		flipCid, err := cid.Parse(validationStats.FlipCids[flipIdx])
		if err != nil {
			log.Error("Unable to parse flip cid. Skipped.", "b", block.Height(), "idx", flipIdx, "err", err)
			continue
		}
		flipStats := db.FlipStats{
			Cid:          convertCid(flipCid),
			ShortAnswers: convertStatsAnswers(flipStats.ShortAnswers),
			LongAnswers:  convertStatsAnswers(flipStats.LongAnswers),
			Status:       convertFlipStatus(ceremony.FlipStatus(flipStats.Status)),
			Answer:       convertAnswer(flipStats.Answer),
			WrongWords:   flipStats.WrongWords,
		}
		flipsStats = append(flipsStats, flipStats)
	}

	return identities, flipsStats, indexer.getFlipsMemPoolKeyData(ctx), birthdays, !validationStats.Failed
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

func (indexer *Indexer) getFlipsMemPoolKeyData(ctx *conversionContext) []db.FlipData {
	flipCidsWithoutData, err := indexer.db.GetCurrentFlipsWithoutData(flipLimitToGetMemPoolData)
	if err != nil {
		log.Error("Unable to get cids without data to try to load it with key from mem pool. Skipped.", "err", err)
		return nil
	}
	if len(flipCidsWithoutData) == 0 {
		return nil
	}
	log.Info(fmt.Sprintf("Flip count for loading data with key from mem pool: %d", len(flipCidsWithoutData)))
	var flipsMemPoolKeyData []db.FlipData
	var parsedData db.FlipContent
	for _, addrFlipCid := range flipCidsWithoutData {
		if indexer.secondaryStorage == nil {
			flipKey := indexer.listener.KeysPool().GetFlipKey(common.HexToAddress(addrFlipCid.Address))
			if flipKey == nil || flipKey.Key == nil {
				log.Error("Missed mem pool key. Skipped.", "cid", addrFlipCid)
				continue
			}
			flipCid, _ := cid.Decode(addrFlipCid.Cid)
			data, err := indexer.getFlipData(flipCid.Bytes(), flipKey.Key, addrFlipCid.Cid, ctx)
			if err != nil {
				log.Error("Unable to get flip data with key from mem pool. Skipped.", "cid", addrFlipCid, "err", err)
				continue
			}
			parsedData, err = parseFlip(addrFlipCid.Cid, data)
			if err != nil {
				log.Error("Unable to parse flip data with key from mem pool. Skipped.", "cid", addrFlipCid, "err", err)
				continue
			}
		} else {
			parsedData, err = indexer.secondaryStorage.GetFlipContent(addrFlipCid.Cid)
			if err != nil {
				log.Error("Unable to get flip data from previous db. Skipped.", "cid", addrFlipCid, "err", err)
				continue
			}
			log.Info("Migrated flip content from previous db", "cid", addrFlipCid)
		}
		flipsMemPoolKeyData = append(flipsMemPoolKeyData, db.FlipData{
			FlipId:  addrFlipCid.FlipId,
			Content: parsedData,
		})
	}
	return flipsMemPoolKeyData
}

func parseFlip(flipCidStr string, data []byte) (db.FlipContent, error) {
	arr := make([]interface{}, 2)
	err := rlp.DecodeBytes(data, &arr)
	if err != nil {
		return db.FlipContent{}, err
	}
	var pics [][]byte
	for _, b := range arr[0].([]interface{}) {
		pics = append(pics, b.([]byte))
	}
	var allOrders [][]byte
	for _, b := range arr[1].([]interface{}) {
		var orders []byte
		for _, bb := range b.([]interface{}) {
			var order byte
			if len(bb.([]byte)) > 0 {
				order = bb.([]byte)[0]
			}
			orders = append(orders, order)
		}
		allOrders = append(allOrders, orders)
	}
	var icon []byte

	if len(pics) > 0 {
		icon, err = compressPic(pics[0])
		if err != nil {
			log.Error("Unable to create flip icon, src pic will be used instead", "cid", flipCidStr, "err", err)
			icon = pics[0]
		}
	}
	return db.FlipContent{
		Pics:   pics,
		Orders: allOrders,
		Icon:   icon,
	}, nil
}

func compressPic(src []byte) ([]byte, error) {
	srcImage, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	var x, y int
	if srcImage.Bounds().Max.X > srcImage.Bounds().Max.Y {
		x = 64
		y = int(float32(srcImage.Bounds().Max.Y) / float32(srcImage.Bounds().Max.X) * 64)
	} else {
		y = 64
		x = int(float32(srcImage.Bounds().Max.X) / float32(srcImage.Bounds().Max.Y) * 64)
	}

	dr := image.Rect(0, 0, x, y)
	dst := image.NewRGBA(dr)
	draw.CatmullRom.Scale(dst, dr, srcImage, srcImage.Bounds(), draw.Src, nil)

	var res bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&res), dst, nil)
	if err != nil {
		return nil, err
	}
	if len(res.Bytes()) == 0 {
		return nil, errors.New("empty converted pic")
	}
	return res.Bytes(), nil
}

func (indexer *Indexer) detectSubmittedFlip(tx *types.Transaction, ctx *conversionContext) *db.Flip {
	if tx.Type != types.SubmitFlipTx {
		return nil
	}
	attachment := attachments.ParseFlipSubmitAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse submitted flip payload. Skipped.", "tx", tx.Hash())
		return nil
	}
	flipCid, err := cid.Parse(attachment.Cid)
	if err != nil {
		log.Error("Unable to parse flip cid. Skipped.", "tx", tx.Hash(), "err", err)
		return nil
	}
	ipfsFlip, err := indexer.listener.Flipper().GetRawFlip(attachment.Cid)
	var size uint32
	if err != nil {
		log.Error("Unable to get flip data to define flip size.", "cid", flipCid, "err", err)
	} else {
		size = uint32(len(ipfsFlip.Data))
	}
	f := &db.Flip{
		TxHash: convertHash(tx.Hash()),
		Cid:    convertCid(flipCid),
		Size:   size,
		Pair:   attachment.Pair,
	}
	return f
}

func (indexer *Indexer) convertShortAnswers(tx *types.Transaction, ctx *conversionContext) {
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
		flipsData, err := indexer.getFlipsData(tx, senderFlips, attachment, ctx)
		if err != nil {
			log.Error("Unable to get flips data. Skipped.", "tx", tx.Hash(), "err", err)
		} else if flipsData != nil {
			ctx.flipsData = append(ctx.flipsData, flipsData...)
		}

		ctx.flipKeys = append(ctx.flipKeys, db.FlipKey{
			TxHash: convertHash(tx.Hash()),
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
			ctx.flipsWords = append(ctx.flipsWords, db.FlipWords{
				FlipId: f.Id,
				TxHash: convertHash(tx.Hash()),
				Word1:  uint16(word1),
				Word2:  uint16(word2),
			})
		}
	} else {
		log.Error("Empty proof for flip words. Skipped.", "tx", tx.Hash())
	}
}

func (indexer *Indexer) getFlipsData(tx *types.Transaction, keyAuthorFlips []db.Flip,
	attachment *attachments.ShortAnswerAttachment, ctx *conversionContext) ([]db.FlipData, error) {

	if len(keyAuthorFlips) == 0 {
		return nil, nil
	}
	var flipsData []db.FlipData
	for _, f := range keyAuthorFlips {
		flipCidStr := f.Cid
		flipCid, _ := cid.Decode(flipCidStr)
		flipData, err := indexer.getFlipData(flipCid.Bytes(), attachment.Key, flipCidStr, ctx)
		if err != nil {
			log.Error("Unable to get flip data. Skipped.", "tx", tx.Hash(), "cid", flipCidStr, "err", err)
			continue
		}
		parsedData, err := parseFlip(flipCidStr, flipData)
		if err != nil {
			log.Error("Unable to parse flip data. Skipped.", "tx", tx.Hash(), "cid", flipCidStr, "err", err)
			continue
		}
		flipsData = append(flipsData, db.FlipData{
			FlipId:  f.Id,
			TxHash:  convertHash(tx.Hash()),
			Content: parsedData,
		})
	}
	return flipsData, nil
}

func getFlipWords(addr common.Address, attachment *attachments.ShortAnswerAttachment, pairId int, appState *appstate.AppState) (word1, word2 int, err error) {
	seed := appState.State.FlipWordsSeed().Bytes()
	proof := attachment.Proof
	identity := appState.State.GetIdentity(addr)
	return ceremony.GetWords(seed, proof, identity.PubKey, common.WordDictionarySize, identity.GetTotalWordPairsCount(), pairId)
}

func (indexer *Indexer) getFlipData(cid []byte, key []byte, cidStr string, ctx *conversionContext) ([]byte, error) {
	ipfsFlip, err := indexer.listener.Flipper().GetRawFlip(cid)
	if err != nil {
		return nil, err
	}
	ctx.flipSizeUpdates = append(ctx.flipSizeUpdates, db.FlipSizeUpdate{
		Cid:  cidStr,
		Size: uint32(len(ipfsFlip.Data)),
	})
	ecdsaKey, _ := crypto.ToECDSA(key)
	encryptionKey := ecies.ImportECDSA(ecdsaKey)
	decryptedFlip, err := encryptionKey.Decrypt(ipfsFlip.Data, nil, nil)
	if err != nil {
		return nil, err
	}
	return decryptedFlip, nil
}

func convertCids(idxs []int, cids [][]byte, block *types.Block) []string {
	var res []string
	for _, idx := range idxs {
		c, err := cid.Parse(cids[idx])
		if err != nil {
			log.Error("Unable to parse cid. Skipped.", "b", block.Height(), "idx", idx, "err", err)
			continue
		}
		res = append(res, convertCid(c))
	}
	return res
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
