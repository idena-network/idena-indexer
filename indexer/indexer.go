package indexer

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/ecies"
	"github.com/idena-network/idena-go/rlp"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/migration/flip"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
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
	identityStates = map[state.IdentityState]string{
		state.Invite:    "Invite",
		state.Candidate: "Candidate",
		state.Newbie:    "Newbie",
		state.Verified:  "Verified",
		state.Suspended: "Suspended",
		state.Zombie:    "Zombie",
		state.Killed:    "Killed",
		state.Undefined: "Undefined",
	}

	txTypes = map[types.TxType]string{
		types.SendTx:               "SendTx",
		types.ActivationTx:         "ActivationTx",
		types.InviteTx:             "InviteTx",
		types.KillTx:               "KillTx",
		types.KillInviteeTx:        "KillInviteeTx",
		types.SubmitFlipTx:         "SubmitFlipTx",
		types.SubmitAnswersHashTx:  "SubmitAnswersHashTx",
		types.SubmitShortAnswersTx: "SubmitShortAnswersTx",
		types.SubmitLongAnswersTx:  "SubmitLongAnswersTx",
		types.EvidenceTx:           "EvidenceTx",
		types.OnlineStatusTx:       "OnlineStatusTx",
	}

	flipStatuses = map[ceremony.FlipStatus]string{
		ceremony.NotQualified:    "NotQualified",
		ceremony.Qualified:       "Qualified",
		ceremony.WeaklyQualified: "WeaklyQualified",
		ceremony.QualifiedByNone: "QualifiedByNone",
	}

	answers = map[types.Answer]string{
		0: "None",
		1: "Left",
		2: "Right",
		3: "Inappropriate",
	}

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
	db                 db.Accessor
	restorer           *restore.Restorer
	state              *indexerState
	sfs                *flip.SecondaryFlipStorage
	genesisBlockHeight uint64
	restore            bool
}

func NewIndexer(listener incoming.Listener,
	db db.Accessor,
	restorer *restore.Restorer,
	sfs *flip.SecondaryFlipStorage,
	genesisBlockHeight uint64,
	restoreInitially bool) *Indexer {
	return &Indexer{
		listener:           listener,
		db:                 db,
		restorer:           restorer,
		sfs:                sfs,
		genesisBlockHeight: genesisBlockHeight,
		restore:            restoreInitially,
	}
}

func (indexer *Indexer) Start() {
	indexer.listener.Listen(indexer.indexBlock, indexer.getHeightToIndex()-1)
}

func (indexer *Indexer) WaitForNodeStop() {
	indexer.listener.WaitForStop()
}

func (indexer *Indexer) Destroy() {
	indexer.listener.Destroy()
	indexer.db.Destroy()
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
				indexer.restore = !indexer.isFirstBlockHeight(block.Height())
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

		data := indexer.convertIncomingData(block)
		indexer.saveData(data)
		indexer.applyOnState(data)

		if data.Block.Height%1000 == 0 {
			log.Info(fmt.Sprintf("Processed block %d", data.Block.Height))
		} else {
			log.Debug(fmt.Sprintf("Processed block %d", data.Block.Height))
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
		totalBalance, totalStake, err := indexer.db.GetTotalCoins()
		if err != nil {
			log.Error(fmt.Sprintf("Unable to get current total coins: %v", err))
			indexer.waitForRetry()
			continue
		}
		return &indexerState{
			lastHeight:   lastHeight,
			totalBalance: totalBalance,
			totalStake:   totalStake,
		}
	}
}

func (indexer *Indexer) convertIncomingData(incomingBlock *types.Block) *db.Data {
	prevState := indexer.listener.AppStateReadonly(incomingBlock.Height() - 1)
	newState := indexer.listener.AppStateReadonly(incomingBlock.Height())
	ctx := &conversionContext{
		blockHeight:       incomingBlock.Height(),
		prevStateReadOnly: prevState,
		newStateReadOnly:  newState,
		addresses:         make(map[string]*db.Address),
	}
	epoch := uint64(prevState.State.Epoch())

	block := indexer.convertBlock(incomingBlock, ctx)
	identities, flipStats, flipsMemPoolData := indexer.determineEpochResult(incomingBlock, ctx)

	firstAddresses := indexer.determineFirstAddresses(incomingBlock, ctx)
	for _, addr := range firstAddresses {
		if curAddr, present := ctx.addresses[addr.Address]; present {
			curAddr.StateChanges = append(curAddr.StateChanges, addr.StateChanges...)
		} else {
			ctx.addresses[addr.Address] = addr
		}
	}

	return &db.Data{
		Epoch:               epoch,
		ValidationTime:      *big.NewInt(ctx.newStateReadOnly.State.NextValidationTime().Unix()),
		Block:               block,
		Identities:          identities,
		SubmittedFlips:      ctx.submittedFlips,
		FlipKeys:            ctx.flipKeys,
		FlipsData:           append(ctx.flipsData, flipsMemPoolData...),
		FlipSizeUpdates:     ctx.flipSizeUpdates,
		FlipStats:           flipStats,
		Addresses:           ctx.getAddresses(),
		BalanceUpdates:      ctx.balanceUpdates,
		BalanceCoins:        indexer.getBalanceCoins(ctx),
		StakeCoins:          indexer.getStakeCoins(ctx),
		SaveEpochSummary:    incomingBlock.Header.Flags().HasFlag(types.ValidationFinished),
		PrevBlockValidators: indexer.convertPrevBlockValidators(incomingBlock, ctx),
		Penalty:             detectChargedPenalty(incomingBlock, ctx.newStateReadOnly),
		BurntPenalties: detectBurntPenalties(incomingBlock, ctx.prevStateReadOnly, ctx.newStateReadOnly,
			indexer.listener.Blockchain()),
	}
}

func (indexer *Indexer) getBalanceCoins(ctx *conversionContext) db.Coins {
	burnt := decimal.Zero
	if ctx.totalFee != nil {
		burnt = decimal.NewFromBigInt(ctx.totalFee, 0)
		burnt = burnt.Mul(decimal.NewFromFloat32(indexer.listener.Config().Consensus.FeeBurnRate))
	}
	diff := decimal.Zero
	if ctx.totalBalanceDiff != nil {
		diff = decimal.NewFromBigInt(ctx.totalBalanceDiff.balance, 0)
	}
	return getCoins(indexer.state.totalBalance, burnt, diff)
}

func (indexer *Indexer) getStakeCoins(ctx *conversionContext) db.Coins {
	burnt := decimal.Zero
	diff := decimal.Zero
	if ctx.totalBalanceDiff != nil {
		diff = decimal.NewFromBigInt(ctx.totalBalanceDiff.stake, 0)
		burnt = decimal.NewFromBigInt(ctx.totalBalanceDiff.burntStake, 0)
	}
	return getCoins(indexer.state.totalStake, burnt, diff)
}

func getCoins(prevTotal decimal.Decimal, burnt decimal.Decimal, diff decimal.Decimal) db.Coins {
	total := prevTotal.Add(blockchain.ConvertToFloat(math.ToInt(diff)))
	minted := burnt.Add(diff)
	res := db.Coins{
		Minted: blockchain.ConvertToFloat(math.ToInt(minted)),
		Burnt:  blockchain.ConvertToFloat(math.ToInt(burnt)),
		Total:  total,
	}
	return res
}

func (indexer *Indexer) isFirstBlock(incomingBlock *types.Block) bool {
	return indexer.isFirstBlockHeight(incomingBlock.Height())
}

func (indexer *Indexer) isFirstBlockHeight(height uint64) bool {
	return height == indexer.genesisBlockHeight+1
}

func (indexer *Indexer) determineFirstAddresses(incomingBlock *types.Block, ctx *conversionContext) []*db.Address {
	if !indexer.isFirstBlock(incomingBlock) {
		return nil
	}
	var addresses []*db.Address
	ctx.newStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
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
	return addresses
}

func (indexer *Indexer) convertBlock(incomingBlock *types.Block, ctx *conversionContext) db.Block {
	stateToApply := ctx.newStateReadOnly.Readonly(ctx.blockHeight - 1)
	txs := indexer.convertTransactions(incomingBlock.Body.Transactions, stateToApply, ctx)

	blockBalanceUpdateDetector := NewBlockBalanceUpdateDetector(incomingBlock,
		indexer.isFirstBlock(incomingBlock),
		stateToApply,
		indexer.listener.Blockchain(),
		ctx)
	balanceUpdates, diff := blockBalanceUpdateDetector.GetUpdates(ctx.newStateReadOnly)
	if len(balanceUpdates) > 0 {
		ctx.balanceUpdates = append(ctx.balanceUpdates, balanceUpdates...)
		if ctx.totalBalanceDiff == nil {
			ctx.totalBalanceDiff = diff
		} else {
			ctx.totalBalanceDiff.Add(diff)
		}
	}

	incomingBlock.Header.Flags()
	return db.Block{
		Height:       incomingBlock.Height(),
		Hash:         convertHash(incomingBlock.Hash()),
		Time:         *incomingBlock.Header.Time(),
		Transactions: txs,
		Proposer:     getProposer(incomingBlock),
		Flags:        convertFlags(incomingBlock.Header.Flags()),
		IsEmpty:      incomingBlock.IsEmpty(),
		Size:         len(incomingBlock.Body.Bytes()),
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

func getProposer(block *types.Block) string {
	if block.IsEmpty() {
		return ""
	}
	return conversion.ConvertAddress(block.Header.ProposedHeader.Coinbase)
}

func (indexer *Indexer) convertPrevBlockValidators(block *types.Block, ctx *conversionContext) []string {
	prevBlock := indexer.listener.Blockchain().GetBlockByHeight(block.Height() - 1)
	if prevBlock == nil {
		return nil
	}
	cert := indexer.listener.Blockchain().GetCertificate(prevBlock.Hash())
	if cert == nil {
		return nil
	}
	var res []string
	for _, vote := range cert.Votes {
		res = append(res, conversion.ConvertAddress(vote.VoterAddr()))
	}
	return res
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
	if f := indexer.determineSubmittedFlip(incomingTx, ctx); f != nil {
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

	txBalanceUpdateDetector := NewTxBalanceUpdateDetector(incomingTx, stateToApply)
	senderPrevState := stateToApply.State.GetIdentityState(sender)
	fee, err := indexer.listener.Blockchain().ApplyTxOnState(stateToApply, incomingTx)
	if err != nil {
		log.Error("Unable to apply tx on state", "tx", incomingTx.Hash(), "err", err)
	}

	if ctx.totalFee == nil {
		ctx.totalFee = new(big.Int)
	}
	ctx.totalFee = new(big.Int).Add(ctx.totalFee, fee)

	balanceUpdates, diff := txBalanceUpdateDetector.GetUpdates(stateToApply)
	if len(balanceUpdates) > 0 {
		ctx.balanceUpdates = append(ctx.balanceUpdates, balanceUpdates...)
		if ctx.totalBalanceDiff == nil {
			ctx.totalBalanceDiff = diff
		} else {
			ctx.totalBalanceDiff.Add(diff)
		}
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
		Fee:     blockchain.ConvertToFloat(fee),
		Size:    incomingTx.Size(),
	}

	return tx
}

func convertTxType(txType types.TxType) string {
	if res, ok := txTypes[txType]; ok {
		return res
	}
	return fmt.Sprintf("Unknown tx type %d", txType)
}

func convertIdentityState(state state.IdentityState) string {
	if res, ok := identityStates[state]; ok {
		return res
	}
	return fmt.Sprintf("Unknown state %d", state)
}

func convertFlipStatus(status ceremony.FlipStatus) string {
	if res, ok := flipStatuses[status]; ok {
		return res
	}
	return fmt.Sprintf("Unknown status %d", status)
}

func convertAnswer(answer types.Answer) string {
	if res, ok := answers[answer]; ok {
		return res
	}
	return fmt.Sprintf("Unknown answer %d", answer)
}

func convertStatsAnswers(incomingAnswers []ceremony.FlipAnswerStats) []db.Answer {
	var answers []db.Answer
	for _, answer := range incomingAnswers {
		answers = append(answers, convertStatsAnswer(answer))
	}
	return answers
}

func convertStatsAnswer(incomingAnswer ceremony.FlipAnswerStats) db.Answer {
	return db.Answer{
		Address: conversion.ConvertAddress(incomingAnswer.Respondent),
		Answer:  convertAnswer(incomingAnswer.Answer),
		Point:   incomingAnswer.Point,
	}
}

func convertHash(hash common.Hash) string {
	return hash.Hex()
}

func convertCid(cid cid.Cid) string {
	return cid.String()
}

func (indexer *Indexer) determineEpochResult(block *types.Block, ctx *conversionContext) ([]db.EpochIdentity, []db.FlipStats, []db.FlipData) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil, nil
	}

	var identities []db.EpochIdentity
	validationStats := indexer.listener.Ceremony().GetValidationStats()

	ctx.prevStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.Address = conversion.ConvertAddress(addr)
		convertedIdentity.State = convertIdentityState(ctx.newStateReadOnly.State.GetIdentityState(addr))
		convertedIdentity.TotalShortPoint = ctx.newStateReadOnly.State.GetShortFlipPoints(addr)
		convertedIdentity.TotalShortFlips = ctx.newStateReadOnly.State.GetQualifiedFlipsCount(addr)
		convertedIdentity.RequiredFlips = ctx.prevStateReadOnly.State.GetRequiredFlips(addr)
		convertedIdentity.MadeFlips = ctx.prevStateReadOnly.State.GetMadeFlips(addr)
		if stats, present := validationStats.IdentitiesPerAddr[addr]; present {
			convertedIdentity.ShortPoint = stats.ShortPoint
			convertedIdentity.ShortFlips = stats.ShortFlips
			convertedIdentity.LongPoint = stats.LongPoint
			convertedIdentity.LongFlips = stats.LongFlips
			convertedIdentity.Approved = stats.Approved
			convertedIdentity.Missed = stats.Missed
			convertedIdentity.ShortFlipCidsToSolve = convertCids(stats.ShortFlipsToSolve, validationStats.FlipCids, block)
			convertedIdentity.LongFlipCidsToSolve = convertCids(stats.LongFlipsToSolve, validationStats.FlipCids, block)
		} else {
			convertedIdentity.Approved = false
			convertedIdentity.Missed = true
		}
		identities = append(identities, convertedIdentity)
	})

	var flipsStats []db.FlipStats
	for flipIdx, stats := range validationStats.FlipsPerIdx {
		flipCid, err := cid.Parse(validationStats.FlipCids[flipIdx])
		if err != nil {
			log.Error("Unable to parse flip cid. Skipped.", "b", block.Height(), "idx", flipIdx, "err", err)
			continue
		}
		flipStats := db.FlipStats{
			Cid:          convertCid(flipCid),
			ShortAnswers: convertStatsAnswers(stats.ShortAnswers),
			LongAnswers:  convertStatsAnswers(stats.LongAnswers),
			Status:       convertFlipStatus(stats.Status),
			Answer:       convertAnswer(stats.Answer),
		}
		flipsStats = append(flipsStats, flipStats)
	}

	return identities, flipsStats, indexer.getFlipsMemPoolKeyData(ctx)
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
	if indexer.sfs != nil && ctx.blockHeight > indexer.sfs.GetLastBlockHeight() {
		indexer.sfs.Destroy()
		indexer.sfs = nil
		log.Info("Completed flip migration")
	}
	var flipsMemPoolKeyData []db.FlipData
	var parsedData db.FlipContent
	for _, addrFlipCid := range flipCidsWithoutData {
		if indexer.sfs == nil {
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
			parsedData, err = indexer.sfs.GetFlipContent(addrFlipCid.Cid)
			if err != nil {
				log.Error("Unable to get flip data from previous db. Skipped.", "cid", addrFlipCid, "err", err)
				continue
			}
			log.Info("Migrated flip content from previous db", "cid", addrFlipCid)
		}
		flipsMemPoolKeyData = append(flipsMemPoolKeyData, db.FlipData{
			Cid:     addrFlipCid.Cid,
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

func (indexer *Indexer) determineSubmittedFlip(tx *types.Transaction, ctx *conversionContext) *db.Flip {
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
	if len(attachment.Key) == 0 {
		return
	}

	flipsData, err := indexer.getFlipsData(tx, attachment, ctx)
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

func (indexer *Indexer) getFlipsData(tx *types.Transaction, attachment *attachments.ShortAnswerAttachment, ctx *conversionContext) ([]db.FlipData, error) {
	sender, _ := types.Sender(tx)
	from := conversion.ConvertAddress(sender)
	keyAuthorFlips, err := indexer.db.GetCurrentFlipCids(from)
	if err != nil {
		return nil, err
	}
	if len(keyAuthorFlips) == 0 {
		return nil, nil
	}
	var flipsData []db.FlipData
	for _, flipCidStr := range keyAuthorFlips {
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
			Cid:     flipCidStr,
			TxHash:  convertHash(tx.Hash()),
			Content: parsedData,
		})
	}
	return flipsData, nil
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

func (indexer *Indexer) applyOnState(data *db.Data) {
	indexer.state.lastHeight = data.Block.Height
	indexer.state.totalBalance = data.BalanceCoins.Total
	indexer.state.totalStake = data.StakeCoins.Total
}

func (indexer *Indexer) waitForRetry() {
	time.Sleep(requestRetryInterval)
}
