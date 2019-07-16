package indexer

import (
	"encoding/hex"
	"github.com/idena-network/idena-go/blockchain"

	//"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/rlp"

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
	indexer.listener.Listen(indexer.indexBlock)
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
		data := convertIncomingData(block, indexer.listener.Node().AppStateReadonly(block.Height()-1),
			indexer.listener.Node().Blockchain(), indexer.listener.Node().Ceremony())
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
}

func convertIncomingData(incomingBlock *types.Block, appState *appstate.AppState, chain *blockchain.Blockchain, c *ceremony.ValidationCeremony) *db.Data {

	ctx := conversionContext{}
	epoch := uint64(appState.State.Epoch())

	block := convertBlock(incomingBlock, appState, chain, &ctx)
	identities, flipStats := determineEpochResult(incomingBlock, appState, c)

	return &db.Data{
		Epoch:          epoch,
		Block:          block,
		Identities:     identities,
		SubmittedFlips: ctx.submittedFlips,
		FlipKeys:       ctx.flipKeys,
		FlipStats:      flipStats,
	}
}

func convertBlock(incomingBlock *types.Block, appState *appstate.AppState, chain *blockchain.Blockchain, ctx *conversionContext) db.Block {
	txs := convertTransactions(incomingBlock.Body.Transactions, appState, chain, ctx)
	return db.Block{
		Height:       incomingBlock.Height(),
		Hash:         incomingBlock.Hash().Hex(),
		Time:         *incomingBlock.Header.Time(),
		Transactions: txs,
	}
}

func convertTransactions(incomingTxs []*types.Transaction, appState *appstate.AppState, chain *blockchain.Blockchain, ctx *conversionContext) []db.Transaction {
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, convertTransaction(incomingTx, appState, chain, ctx))
	}
	return txs
}

func convertTransaction(incomingTx *types.Transaction, appState *appstate.AppState, chain *blockchain.Blockchain, ctx *conversionContext) db.Transaction {
	if flip := determineSubmittedFlip(incomingTx); flip != nil {
		ctx.submittedFlips = append(ctx.submittedFlips, *flip)
	}

	convertShortAnswers(incomingTx, ctx)

	sender, _ := types.Sender(incomingTx)
	from := sender.Hex()
	var to string
	if incomingTx.To != nil {
		to = incomingTx.To.Hex()
	}

	fee, err := chain.ApplyTxOnState(appState, incomingTx)
	if err != nil {
		log.Error("Unable to calculate tx fee", "tx", incomingTx.Hash(), "err", err)
	}

	tx := db.Transaction{
		Type:    convertTxType(incomingTx.Type),
		Payload: incomingTx.Payload,
		Hash:    incomingTx.Hash().Hex(),
		From:    from,
		To:      to,
		Amount:  incomingTx.Amount,
		Fee:     fee,
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

func determineEpochResult(block *types.Block, appState *appstate.AppState, c *ceremony.ValidationCeremony) ([]db.EpochIdentity, []db.FlipStats) {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil, nil
	}

	var identities []db.EpochIdentity
	validationStats := c.GetValidationStats()

	for addr, stats := range validationStats.IdentitiesPerAddr {
		addrHex := addr.Hex()
		identity := db.EpochIdentity{
			Address:    addrHex,
			ShortPoint: stats.ShortPoint,
			ShortFlips: stats.ShortFlips,
			LongPoint:  stats.LongPoint,
			LongFlips:  stats.LongFlips,
			State:      convertIdentityState(stats.State),
		}
		identities = append(identities, identity)
	}

	var flipsStats []db.FlipStats
	for flipIdx, stats := range validationStats.FlipsPerIdx {
		flipCid, err := cid.Parse(validationStats.FlipCids[flipIdx])
		if err != nil {
			log.Error("Unable to parse flip cid. Skipped.", "b", block.Height(), "idx", flipIdx, "err", err)
			continue
		}
		flipStats := db.FlipStats{
			Cid:          flipCid.String(),
			ShortAnswers: stats.ShortAnswers,
			LongAnswers:  stats.LongAnswers,
			Status:       convertFlipStatus(stats.Status),
			Answer:       stats.Answer,
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
	flip := &db.Flip{
		TxHash: tx.Hash().Hex(),
		Cid:    flipCid.String(),
	}
	return flip
}

func convertShortAnswers(tx *types.Transaction, ctx *conversionContext) {
	if tx.Type != types.SubmitShortAnswersTx {
		return
	}
	answer := attachments.ShortAnswerAttachment{}
	if err := rlp.DecodeBytes(tx.Payload, &answer); err != nil {
		log.Error("Unable to parse short answers payload. Skipped.", "tx", tx.Hash(), "err", err)
		return
	}
	if len(answer.Key) > 0 {
		ctx.flipKeys = append(ctx.flipKeys, db.FlipKey{
			TxHash: tx.Hash().Hex(),
			Key:    hex.EncodeToString(answer.Key),
		})
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

func (indexer *Indexer) waitForRetry() {
	time.Sleep(requestRetryInterval)
}
