package indexer

import (
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/ecies"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"math/big"
	"time"
)

const (
	requestRetryInterval      = time.Second * 5
	firstBlockHeight          = 2
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
	listener   incoming.Listener
	db         db.Accessor
	lastHeight uint64
}

func NewIndexer(listener incoming.Listener, db db.Accessor) *Indexer {
	return &Indexer{
		listener: listener,
		db:       db,
	}
}

func (indexer *Indexer) Start() {
	indexer.listener.Listen(indexer.indexBlock, indexer.getHeightToIndex()-1)
}

func (indexer *Indexer) Destroy() {
	indexer.listener.Destroy()
	indexer.db.Destroy()
}

func (indexer *Indexer) indexBlock(block *types.Block) {
	for {
		heightToIndex := indexer.getHeightToIndex()

		if !isFirstBlock(block) && block.Height() > heightToIndex {
			panic(fmt.Sprintf("Incoming block height=%d is greater than expected %d", block.Height(), heightToIndex))
		}

		if block.Height() < heightToIndex {
			height := block.Height() - 1
			if err := indexer.resetTo(height); err != nil {
				log.Error(fmt.Sprintf("Unable to reset to height=%d", height), "err", err)
				indexer.waitForRetry()
			} else {
				log.Info(fmt.Sprintf("Indexer db has been reset to height=%d", height))
			}
			// retry in any case to ensure incoming height equals to expected height to index after reset
			continue
		}

		data := indexer.convertIncomingData(block)
		indexer.saveData(data)
		indexer.lastHeight = data.Block.Height
		log.Debug(fmt.Sprintf("Processed block %d", data.Block.Height))
		return
	}
}

func (indexer *Indexer) resetTo(height uint64) error {
	err := indexer.db.ResetTo(height)
	if err != nil {
		return err
	}
	indexer.lastHeight = indexer.loadHeightToIndex()
	return nil
}

func (indexer *Indexer) getHeightToIndex() uint64 {
	if indexer.lastHeight == 0 {
		return indexer.loadHeightToIndex()
	}
	return indexer.lastHeight + 1
}

func (indexer *Indexer) loadHeightToIndex() uint64 {
	for {
		lastHeight, err := indexer.db.GetLastHeight()
		if err != nil {
			log.Error(fmt.Sprintf("Unable to get last saved height: %v", err))
			indexer.waitForRetry()
			continue
		}
		return lastHeight + 1
	}
}

type conversionContext struct {
	blockHeight         uint64
	submittedFlips      []db.Flip
	flipKeys            []db.FlipKey
	flipsData           []db.FlipData
	addresses           map[string]*db.Address
	prevStateReadOnly   *appstate.AppState
	newStateReadOnly    *appstate.AppState
	chain               *blockchain.Blockchain
	c                   *ceremony.ValidationCeremony
	fp                  *flip.Flipper
	getFlips            func(string) ([]string, error)
	getFlipsWithoutData func(uint32) ([]string, error)
}

func (ctx *conversionContext) getAddresses() []db.Address {
	var addresses []db.Address
	for _, addr := range ctx.addresses {
		addresses = append(addresses, *addr)
	}
	return addresses
}

func (indexer *Indexer) convertIncomingData(incomingBlock *types.Block) *db.Data {
	prevState := indexer.listener.AppStateReadonly(incomingBlock.Height() - 1)
	newState := indexer.listener.AppStateReadonly(incomingBlock.Height())
	ctx := conversionContext{
		blockHeight:         incomingBlock.Height(),
		prevStateReadOnly:   prevState,
		newStateReadOnly:    newState,
		chain:               indexer.listener.Blockchain(),
		c:                   indexer.listener.Ceremony(),
		fp:                  indexer.listener.Flipper(),
		getFlips:            indexer.db.GetCurrentFlipCids,
		getFlipsWithoutData: indexer.db.GetCurrentFlipsWithoutData,
		addresses:           make(map[string]*db.Address),
	}
	epoch := uint64(prevState.State.Epoch())

	block := convertBlock(incomingBlock, &ctx)
	identities, flipStats, flipsMemPoolData := determineEpochResult(incomingBlock, &ctx)

	firstAddresses := determineFirstAddresses(incomingBlock, &ctx)
	for _, addr := range firstAddresses {
		if curAddr, present := ctx.addresses[addr.Address]; present {
			curAddr.StateChanges = append(curAddr.StateChanges, addr.StateChanges...)
		} else {
			ctx.addresses[addr.Address] = addr
		}
	}

	return &db.Data{
		Epoch:            epoch,
		ValidationTime:   *big.NewInt(ctx.newStateReadOnly.State.NextValidationTime().Unix()),
		Block:            block,
		Identities:       identities,
		SubmittedFlips:   ctx.submittedFlips,
		FlipKeys:         ctx.flipKeys,
		FlipsData:        ctx.flipsData,
		FlipsMemPoolData: flipsMemPoolData,
		FlipStats:        flipStats,
		Addresses:        ctx.getAddresses(),
		Balances:         determineBalanceChanges(&ctx),
	}
}

func isFirstBlock(incomingBlock *types.Block) bool {
	return incomingBlock.Height() == firstBlockHeight
}

func determineFirstAddresses(incomingBlock *types.Block, ctx *conversionContext) []*db.Address {
	if !isFirstBlock(incomingBlock) {
		return nil
	}
	var addresses []*db.Address
	ctx.newStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		addresses = append(addresses, &db.Address{
			Address: convertAddress(addr),
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

func bytesToAddr(bytes []byte) common.Address {
	addr := common.Address{}
	addr.SetBytes(bytes[1:])
	return addr
}

func determineBalanceChanges(ctx *conversionContext) []db.Balance {
	var balances []db.Balance

	callback := func(addr common.Address) bool {
		prevBalance := ctx.prevStateReadOnly.State.GetBalance(addr)
		prevStake := ctx.prevStateReadOnly.State.GetStakeBalance(addr)
		balance := ctx.newStateReadOnly.State.GetBalance(addr)
		stake := ctx.newStateReadOnly.State.GetStakeBalance(addr)
		if balance.Cmp(prevBalance) != 0 || stake.Cmp(prevStake) != 0 {
			balances = append(balances, db.Balance{
				Address: convertAddress(addr),
				Balance: blockchain.ConvertToFloat(balance),
				Stake:   blockchain.ConvertToFloat(stake),
			})
		}
		return false
	}

	prevAddrs := mapset.NewSet()
	ctx.prevStateReadOnly.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		addr := bytesToAddr(key)
		callback(addr)
		prevAddrs.Add(addr)
		return false
	})

	ctx.newStateReadOnly.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		addr := bytesToAddr(key)
		if !prevAddrs.Contains(addr) {
			callback(addr)
		}
		return false
	})
	return balances
}

func convertBlock(incomingBlock *types.Block, ctx *conversionContext) db.Block {
	txs := convertTransactions(incomingBlock.Body.Transactions, ctx)
	incomingBlock.Header.Flags()
	return db.Block{
		Height:       incomingBlock.Height(),
		Hash:         convertHash(incomingBlock.Hash()),
		Time:         *incomingBlock.Header.Time(),
		Transactions: txs,
		Proposer:     getProposer(incomingBlock),
		Flags:        convertFlags(incomingBlock.Header.Flags()),
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
	return convertAddress(block.Header.ProposedHeader.Coinbase)
}

func convertTransactions(incomingTxs []*types.Transaction, ctx *conversionContext) []db.Transaction {
	if len(incomingTxs) == 0 {
		return nil
	}
	stateToApply := ctx.prevStateReadOnly.Readonly(ctx.blockHeight - 1)
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, convertTransaction(incomingTx, ctx, stateToApply))
	}
	return txs
}

func convertTransaction(incomingTx *types.Transaction, ctx *conversionContext, stateToApply *appstate.AppState) db.Transaction {
	if f := determineSubmittedFlip(incomingTx, ctx); f != nil {
		ctx.submittedFlips = append(ctx.submittedFlips, *f)
	}

	convertShortAnswers(incomingTx, ctx)
	txHash := convertHash(incomingTx.Hash())

	sender, _ := types.Sender(incomingTx)
	from := convertAddress(sender)
	if _, present := ctx.addresses[from]; !present {
		ctx.addresses[from] = &db.Address{
			Address: from,
		}
	}

	var to string
	var recipientPrevState *state.IdentityState
	if incomingTx.To != nil {
		to = convertAddress(*incomingTx.To)
		if _, present := ctx.addresses[to]; !present {
			ctx.addresses[to] = &db.Address{
				Address: to,
			}
		}
		st := stateToApply.State.GetIdentityState(*incomingTx.To)
		recipientPrevState = &st
	}

	senderPrevState := stateToApply.State.GetIdentityState(sender)
	fee, err := ctx.chain.ApplyTxOnState(stateToApply, incomingTx)
	if err != nil {
		log.Error("Unable to calculate tx fee", "tx", incomingTx.Hash(), "err", err)
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
		Address: convertAddress(incomingAnswer.Respondent),
		Answer:  convertAnswer(incomingAnswer.Answer),
	}
}

func convertAddress(address common.Address) string {
	return address.Hex()
}

func convertHash(hash common.Hash) string {
	return hash.Hex()
}

func convertCid(cid cid.Cid) string {
	return cid.String()
}

func determineEpochResult(block *types.Block, ctx *conversionContext) ([]db.EpochIdentity, []db.FlipStats, []db.FlipData) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil, nil
	}

	var identities []db.EpochIdentity
	validationStats := ctx.c.GetValidationStats()

	ctx.prevStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.Address = convertAddress(addr)
		convertedIdentity.State = convertIdentityState(ctx.newStateReadOnly.State.GetIdentityState(addr))
		convertedIdentity.TotalShortPoint = ctx.newStateReadOnly.State.GetShortFlipPoints(addr)
		convertedIdentity.TotalShortFlips = ctx.newStateReadOnly.State.GetQualifiedFlipsCount(addr)
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

	return identities, flipsStats, getFlipsMemPoolKeyData(ctx)
}

func getFlipsMemPoolKeyData(ctx *conversionContext) []db.FlipData {
	flipCidsWithoutData, err := ctx.getFlipsWithoutData(flipLimitToGetMemPoolData)
	if err != nil {
		log.Error("Unable to get cids without data to try to load it with key from mem pool. Skipped.", "err", err)
		return nil
	}
	if len(flipCidsWithoutData) == 0 {
		return nil
	}
	var flipsMemPoolKeyData []db.FlipData
	log.Info(fmt.Sprintf("Flip count for loading data with key from mem pool: %d", len(flipCidsWithoutData)))
	for _, flipCidStr := range flipCidsWithoutData {
		flipCid, _ := cid.Decode(flipCidStr)
		data, err := ctx.fp.GetFlip(flipCid.Bytes())
		if err != nil {
			log.Error("Unable to get flip data with key from mem pool. Skipped.", "cid", flipCidStr, "err", err)
			continue
		}
		flipsMemPoolKeyData = append(flipsMemPoolKeyData, db.FlipData{
			Cid:    flipCidStr,
			TxHash: "",
			Data:   data,
		})
	}
	return flipsMemPoolKeyData
}

func determineSubmittedFlip(tx *types.Transaction, ctx *conversionContext) *db.Flip {
	if tx.Type != types.SubmitFlipTx {
		return nil
	}
	flipCid, err := cid.Parse(tx.Payload)
	if err != nil {
		log.Error("Unable to parse flip cid. Skipped.", "tx", tx.Hash(), "err", err)
		return nil
	}
	ipfsFlip, err := ctx.fp.GetRawFlip(tx.Payload)
	var size uint32
	if err != nil {
		log.Error("Unable to get flips data to define flip size.", "tx", tx.Hash(), "err", err)
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

func convertShortAnswers(tx *types.Transaction, ctx *conversionContext) {
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

	flipsData, err := getFlipsData(tx, attachment, ctx)
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

func getFlipsData(tx *types.Transaction, attachment *attachments.ShortAnswerAttachment, ctx *conversionContext) ([]db.FlipData, error) {
	sender, _ := types.Sender(tx)
	from := convertAddress(sender)
	keyAuthorFlips, err := ctx.getFlips(from)
	if err != nil {
		return nil, err
	}
	if len(keyAuthorFlips) == 0 {
		return nil, nil
	}
	var flipsData []db.FlipData
	for _, flipCidStr := range keyAuthorFlips {
		flipCid, _ := cid.Decode(flipCidStr)
		flipData, err := getFlipData(flipCid.Bytes(), attachment.Key, ctx)
		if err != nil {
			log.Error("Unable to get flip data. Skipped.", "tx", tx.Hash(), "cid", flipCidStr, "err", err)
			continue
		}
		flipsData = append(flipsData, db.FlipData{
			Cid:    flipCidStr,
			TxHash: convertHash(tx.Hash()),
			Data:   flipData,
		})
	}
	return flipsData, nil
}

func getFlipData(cid []byte, key []byte, ctx *conversionContext) ([]byte, error) {
	ipfsFlip, err := ctx.fp.GetRawFlip(cid)
	if err != nil {
		return nil, err
	}
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

func (indexer *Indexer) waitForRetry() {
	time.Sleep(requestRetryInterval)
}
