package indexer

import (
	"encoding/hex"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/flip"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/ecies"
	"math/big"

	//"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/core/ceremony"
	//"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/log"
	//"github.com/idena-network/idena-go/rlp"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-cid"
	"time"
)

const requestRetryInterval = time.Second * 5

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
		if !indexer.checkBlock(block, heightToIndex) {
			return
		}
		prevState := indexer.listener.Node().AppStateReadonly(block.Height() - 1)
		newState := indexer.listener.Node().AppStateReadonly(block.Height())
		data := convertIncomingData(block, prevState, newState,
			indexer.listener.Node().Blockchain(), indexer.listener.Node().Ceremony(), indexer.listener.Node().Flipper(),
			indexer.db.GetCurrentFlipCids)
		indexer.saveData(data)
		indexer.lastHeight = data.Block.Height
		log.Debug(fmt.Sprintf("Processed block %d", data.Block.Height))
	}
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

func (indexer *Indexer) checkBlock(block *types.Block, fromHeight uint64) bool {
	if block.Height() < fromHeight {
		return false
	}

	if block.Height() > fromHeight {
		log.Warn(fmt.Sprintf("Incoming block height=%d is greater than expected %d", block.Height(), fromHeight))
	}
	return true
}

type conversionContext struct {
	submittedFlips []db.Flip
	flipKeys       []db.FlipKey
	flipsData      []db.FlipData
	addresses      []db.Address
	prevState      *appstate.AppState
	newState       *appstate.AppState
	chain          *blockchain.Blockchain
	c              *ceremony.ValidationCeremony
	fp             *flip.Flipper
	getFlips       func(string) ([]string, error)
}

func convertIncomingData(incomingBlock *types.Block, prevState *appstate.AppState, newState *appstate.AppState,
	chain *blockchain.Blockchain, c *ceremony.ValidationCeremony, fp *flip.Flipper, getFlips func(string) ([]string, error)) *db.Data {

	ctx := conversionContext{
		prevState: prevState,
		newState:  newState,
		chain:     chain,
		c:         c,
		fp:        fp,
		getFlips:  getFlips,
	}
	epoch := uint64(prevState.State.Epoch())

	block := convertBlock(incomingBlock, &ctx)
	identities, flipStats := determineEpochResult(incomingBlock, &ctx)

	ctx.addresses = append(ctx.addresses, determineFirstAddresses(incomingBlock, &ctx)...)

	return &db.Data{
		Epoch:          epoch,
		ValidationTime: *big.NewInt(ctx.newState.State.NextValidationTime().Unix()),
		Block:          block,
		Identities:     identities,
		SubmittedFlips: ctx.submittedFlips,
		FlipKeys:       ctx.flipKeys,
		FlipsData:      ctx.flipsData,
		FlipStats:      flipStats,
		Addresses:      ctx.addresses,
	}
}

func isFirstBlock(incomingBlock *types.Block) bool {
	return incomingBlock.Height() == 2
}

func determineFirstAddresses(incomingBlock *types.Block, ctx *conversionContext) []db.Address {
	if !isFirstBlock(incomingBlock) {
		return nil
	}
	var addresses []db.Address
	ctx.newState.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		addresses = append(addresses, db.Address{
			Address:  convertAddress(addr),
			NewState: convertIdentityState(identity.State),
		})
	})
	return addresses
}

func convertBlock(incomingBlock *types.Block, ctx *conversionContext) db.Block {
	txs := convertTransactions(incomingBlock.Body.Transactions, ctx)
	return db.Block{
		Height:       incomingBlock.Height(),
		Hash:         convertHash(incomingBlock.Hash()),
		Time:         *incomingBlock.Header.Time(),
		Transactions: txs,
	}
}

func convertTransactions(incomingTxs []*types.Transaction, ctx *conversionContext) []db.Transaction {
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, convertTransaction(incomingTx, ctx))
	}
	return txs
}

func convertTransaction(incomingTx *types.Transaction, ctx *conversionContext) db.Transaction {
	if f := determineSubmittedFlip(incomingTx); f != nil {
		ctx.submittedFlips = append(ctx.submittedFlips, *f)
	}

	convertShortAnswers(incomingTx, ctx)
	txHash := convertHash(incomingTx.Hash())

	sender, _ := types.Sender(incomingTx)
	from := convertAddress(sender)
	ctx.addresses = append(ctx.addresses, convertTxAddress(sender, ctx))

	var to string
	if incomingTx.To != nil {
		to = convertAddress(*incomingTx.To)
		if to != from {
			ctx.addresses = append(ctx.addresses, convertTxAddress(*incomingTx.To, ctx))
		}
	}

	fee, err := ctx.chain.ApplyTxOnState(ctx.prevState, incomingTx)
	if err != nil {
		log.Error("Unable to calculate tx fee", "tx", incomingTx.Hash(), "err", err)
	}

	tx := db.Transaction{
		Type:    convertTxType(incomingTx.Type),
		Payload: incomingTx.Payload,
		Hash:    txHash,
		From:    from,
		To:      to,
		Amount:  incomingTx.Amount,
		Fee:     fee,
	}
	return tx
}

func convertTxAddress(address common.Address, ctx *conversionContext) db.Address {
	prevAddrState := ctx.prevState.State.GetIdentityState(address)
	curAddrState := ctx.newState.State.GetIdentityState(address)
	var newAddrState string
	if curAddrState != prevAddrState {
		newAddrState = convertIdentityState(curAddrState)
	}

	return db.Address{
		Address:  convertAddress(address),
		NewState: newAddrState,
	}
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

func determineEpochResult(block *types.Block, ctx *conversionContext) ([]db.EpochIdentity, []db.FlipStats) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil
	}

	var identities []db.EpochIdentity
	validationStats := ctx.c.GetValidationStats()

	ctx.prevState.State.IterateOverIdentities(func(addr common.Address, _ state.Identity) {
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.Address = convertAddress(addr)
		convertedIdentity.State = convertIdentityState(ctx.newState.State.GetIdentityState(addr))
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

	return identities, flipsStats
}

func determineSubmittedFlip(tx *types.Transaction) *db.Flip {
	if tx.Type != types.SubmitFlipTx {
		return nil
	}
	flipCid, err := cid.Parse(tx.Payload)
	if err != nil {
		log.Error("Unable to parse flip cid. Skipped.", "tx", tx.Hash(), "err", err)
		return nil
	}
	f := &db.Flip{
		TxHash: convertHash(tx.Hash()),
		Cid:    convertCid(flipCid),
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
	} else {
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
