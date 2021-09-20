package indexer

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/crypto/vrf"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/core/flip"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/events"
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
		types.NewGenesis:              "NewGenesis",
	}
)

type Indexer struct {
	enabled                     bool
	listener                    incoming.Listener
	memPoolIndexer              *mempool.Indexer
	db                          db.Accessor
	restorer                    *restore.Restorer
	state                       *indexerState
	secondaryStorage            *runtime.SecondaryStorage
	restore                     bool
	pm                          monitoring.PerformanceMonitor
	flipLoader                  flip.Loader
	firstBlockHeight            uint64
	firstBlockHeightInitialized bool
	upgradeVotingHistoryCtx     *upgradeVotingHistoryCtx
	eventBus                    eventbus.Bus
}

type upgradeVotingHistoryCtx struct {
	holder               upgrade.UpgradesVotingHolder
	shortHistoryItems    int
	shortHistoryMinShift int
	queue                chan *upgradesVotesWrapper
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

type upgradesVotesWrapper struct {
	upgradesVotes []*db.UpgradeVotes
	height        uint64
}

func NewIndexer(
	enabled bool,
	listener incoming.Listener,
	mempoolIndexer *mempool.Indexer,
	dbAccessor db.Accessor,
	restorer *restore.Restorer,
	secondaryStorage *runtime.SecondaryStorage,
	restoreInitially bool,
	pm monitoring.PerformanceMonitor,
	flipLoader flip.Loader,
	upgradesVotingHolder upgrade.UpgradesVotingHolder,
	upgradeVotingShortHistoryItems int,
	upgradeVotingShortHistoryMinShift int,
	eventBus eventbus.Bus,
) *Indexer {
	return &Indexer{
		enabled:          enabled,
		listener:         listener,
		memPoolIndexer:   mempoolIndexer,
		db:               dbAccessor,
		restorer:         restorer,
		secondaryStorage: secondaryStorage,
		restore:          restoreInitially,
		pm:               pm,
		flipLoader:       flipLoader,
		eventBus:         eventBus,
		upgradeVotingHistoryCtx: &upgradeVotingHistoryCtx{
			holder:               upgradesVotingHolder,
			shortHistoryItems:    upgradeVotingShortHistoryItems,
			shortHistoryMinShift: upgradeVotingShortHistoryMinShift,
			queue:                make(chan *upgradesVotesWrapper, 5),
		},
	}
}

func (indexer *Indexer) Start() {
	indexer.memPoolIndexer.Initialize(indexer.listener.NodeEventBus())
	go indexer.loopRefreshUpgradeVotingHistorySummaries()
	go indexer.updateUpgradesInfo()
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

	if !indexer.enabled {
		for {
			log.Warn("Indexing is disabled")
			time.Sleep(time.Minute)
		}
	}

	indexer.initFirstBlockHeight()

	var genesisBlock *types.Block
	for {
		heightToIndex := indexer.getHeightToIndex()
		if genesisBlock != nil {
			heightToIndex++
		}
		if block.Height() > heightToIndex {
			genesisBlock = indexer.getGenesisBlock()
			if genesisBlock == nil {
				log.Error("Unable to get genesis block")
				indexer.waitForRetry()
				continue
			}
			if block.Height() != genesisBlock.Height()+1 {
				panic(fmt.Sprintf("Incoming block height=%d is greater than expected %d", block.Height(), heightToIndex))
			}
		}
		if block.Height() < heightToIndex {
			log.Info(fmt.Sprintf("Incoming block height=%d is less than expected %d, start resetting indexer db...", block.Height(), heightToIndex))
			heightToReset := block.Height() - 1

			if !indexer.isFirstBlock(block) && block.Header.ParentHash() == indexer.listener.NodeCtx().Blockchain.GenesisInfo().Genesis.Hash() {
				log.Info(fmt.Sprintf("Block %d is first after new genesis", block.Height()))
				heightToReset--
				genesisBlock = indexer.getGenesisBlock()
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

		err := indexer.initializeStateIfNeeded(block)
		if err != nil {
			panic(err)
		}

		if genesisBlock != nil {
			indexer.statsHolder().Disable()
			res, err := indexer.convertIncomingData(genesisBlock)
			if err != nil {
				panic(err)
			}
			indexer.saveData(res.dbData)
			log.Info(fmt.Sprintf("Processed genesis block %d", genesisBlock.Height()))
			genesisBlock = nil
		}

		if indexer.restore {
			log.Info("Start restoring DB data...")
			indexer.restorer.Restore()
			log.Info("DB data has been restored")
			indexer.restore = false
		}

		indexer.statsHolder().Enable()
		indexer.pm.Start("Convert")
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

		if block.Header.Flags().HasFlag(types.ValidationFinished) {
			indexer.eventBus.Publish(&events.NewEpochEvent{Epoch: uint16(res.dbData.Epoch) + 1})
		}

		log.Info(fmt.Sprintf("Processed block %d", block.Height()))

		indexer.refreshUpgradeVotingHistorySummaries(res.dbData.UpgradesVotes, block.Height())

		return
	}
}

func (indexer *Indexer) initFirstBlockHeight() {
	if indexer.firstBlockHeightInitialized {
		return
	}
	if indexer.state.lastIndexedHeight == 0 {
		indexer.firstBlockHeight = indexer.getGenesisBlock().Height()
	} else {
		indexer.firstBlockHeight = 1
	}
	indexer.eventBus.Publish(&events.CurrentEpochEvent{Epoch: indexer.listener.NodeCtx().AppState.State.Epoch()})
	indexer.firstBlockHeightInitialized = true
}

func (indexer *Indexer) getGenesisBlock() *types.Block {
	return indexer.listener.NodeCtx().Blockchain.GetBlock(
		indexer.listener.NodeCtx().Blockchain.GenesisInfo().Genesis.Hash(),
	)
}

func (indexer *Indexer) resetTo(height uint64) error {
	err := indexer.db.ResetTo(height)
	if err != nil {
		return err
	}
	indexer.state = indexer.loadState()
	indexer.firstBlockHeightInitialized = false
	indexer.initFirstBlockHeight()
	return nil
}

func (indexer *Indexer) getHeightToIndex() uint64 {
	if indexer.state == nil {
		indexer.state = indexer.loadState()
	}
	return indexer.state.lastIndexedHeight + 1
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
			lastIndexedHeight: lastHeight,
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
	isGenesisBlock := incomingBlock.Hash() == indexer.listener.NodeCtx().Blockchain.GenesisInfo().Genesis.Hash()
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
		addresses:   make(map[string]*db.Address),
		killedAddrs: make(map[common.Address]struct{}),
	}
	collectorStats := indexer.statsHolder().GetStats()
	epoch := uint64(prevState.State.Epoch())

	indexer.pm.Complete("InitCtx")
	indexer.pm.Start("ConvertBlock")
	isFirstEpochBlock := incomingBlock.Height() == prevState.State.EpochBlock()+1
	block, err := indexer.convertBlock(incomingBlock, ctx, collector, prevState.ValidatorsCache)
	if err != nil {
		return nil, err
	}
	indexer.pm.Complete("ConvertBlock")
	epochResult := indexer.detectEpochResult(incomingBlock, ctx, collector)

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

	delegationSwitches := detectDelegationSwitches(incomingBlock, ctx.prevStateReadOnly, ctx.newStateReadOnly, collector.killedAddrs, collector.switchDelegationTxs)

	for _, removedTransitiveDelegation := range collectorStats.RemovedTransitiveDelegations {
		delegationSwitches = append(delegationSwitches, &db.DelegationSwitch{
			Delegator: removedTransitiveDelegation.Delegator,
		})
	}

	upgradesVotes := detectUpgradeVotes(indexer.upgradeVotingHistoryCtx.holder.Get(), indexer.listener.Config().Consensus.Version)

	poolSizes := detectPoolSizeUpdates(delegationSwitches, collector.getAddresses(), func() []db.EpochIdentity {
		if epochResult == nil {
			return nil
		}
		return epochResult.Identities
	}(), ctx.prevStateReadOnly, ctx.newStateReadOnly)

	dbData := &db.Data{
		Epoch:                                    epoch,
		ValidationTime:                           *big.NewInt(ctx.newStateReadOnly.State.NextValidationTime().Unix()),
		Block:                                    block,
		ActivationTxTransfers:                    collectorStats.ActivationTxTransfers,
		KillTxTransfers:                          collectorStats.KillTxTransfers,
		KillInviteeTxTransfers:                   collectorStats.KillInviteeTxTransfers,
		ActivationTxs:                            collectorStats.ActivationTxs,
		KillInviteeTxs:                           collector.killInviteeTxs,
		BecomeOnlineTxs:                          collector.becomeOnlineTxs,
		BecomeOfflineTxs:                         collector.becomeOfflineTxs,
		SubmittedFlips:                           collector.submittedFlips,
		DeletedFlips:                             collector.deletedFlips,
		FlipKeys:                                 collector.flipKeys,
		FlipsWords:                               collector.flipsWords,
		Addresses:                                collector.getAddresses(),
		ChangedBalances:                          balanceUpdates,
		Coins:                                    coins,
		Penalties:                                convertChargedPenalties(collectorStats.ChargedPenaltiesByAddr),
		MiningRewards:                            convertMiningRewards(collectorStats.MiningRewards),
		BurntCoinsPerAddr:                        collectorStats.BurntCoinsByAddr,
		BalanceUpdates:                           collectorStats.BalanceUpdates,
		CommitteeRewardShare:                     collectorStats.CommitteeRewardShare,
		OracleVotingContracts:                    collectorStats.OracleVotingContracts,
		OracleVotingContractCallStarts:           collectorStats.OracleVotingContractCallStarts,
		OracleVotingContractCallVoteProofs:       collectorStats.OracleVotingContractCallVoteProofs,
		OracleVotingContractCallVotes:            collectorStats.OracleVotingContractCallVotes,
		OracleVotingContractCallFinishes:         collectorStats.OracleVotingContractCallFinishes,
		OracleVotingContractCallProlongations:    collectorStats.OracleVotingContractCallProlongations,
		OracleVotingContractCallAddStakes:        collectorStats.OracleVotingContractCallAddStakes,
		OracleVotingContractTerminations:         collectorStats.OracleVotingContractTerminations,
		OracleLockContracts:                      collectorStats.OracleLockContracts,
		OracleLockContractCallCheckOracleVotings: collectorStats.OracleLockContractCallCheckOracleVotings,
		OracleLockContractCallPushes:             collectorStats.OracleLockContractCallPushes,
		OracleLockContractTerminations:           collectorStats.OracleLockContractTerminations,
		ClearOldEpochCommittees:                  isFirstEpochBlock,
		RefundableOracleLockContracts:            collectorStats.RefundableOracleLockContracts,
		RefundableOracleLockContractCallDeposits: collectorStats.RefundableOracleLockContractCallDeposits,
		RefundableOracleLockContractCallPushes:   collectorStats.RefundableOracleLockContractCallPushes,
		RefundableOracleLockContractCallRefunds:  collectorStats.RefundableOracleLockContractCallRefunds,
		RefundableOracleLockContractTerminations: collectorStats.RefundableOracleLockContractTerminations,
		MultisigContracts:                        collectorStats.MultisigContracts,
		MultisigContractCallAdds:                 collectorStats.MultisigContractCallAdds,
		MultisigContractCallSends:                collectorStats.MultisigContractCallSends,
		MultisigContractCallPushes:               collectorStats.MultisigContractCallPushes,
		MultisigContractTerminations:             collectorStats.MultisigContractTerminations,
		TimeLockContracts:                        collectorStats.TimeLockContracts,
		TimeLockContractCallTransfers:            collectorStats.TimeLockContractCallTransfers,
		TimeLockContractTerminations:             collectorStats.TimeLockContractTerminations,
		TxReceipts:                               collectorStats.TxReceipts,
		ContractTxsBalanceUpdates:                collectorStats.ContractTxsBalanceUpdates,
		EpochResult:                              epochResult,
		DelegationSwitches:                       delegationSwitches,
		UpgradesVotes:                            upgradesVotes,
		PoolSizes:                                poolSizes,
		MinersHistoryItem:                        detectMinersHistoryItem(ctx.prevStateReadOnly, ctx.newStateReadOnly),
		RemovedTransitiveDelegations:             collectorStats.RemovedTransitiveDelegations,
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
	return height == indexer.firstBlockHeight
}

func (indexer *Indexer) detectFirstAddresses(incomingBlock *types.Block, ctx *conversionContext) []*db.Address {
	if !indexer.isFirstBlock(incomingBlock) {
		return nil
	}
	stateChangeByAddr := make(map[common.Address]db.AddressStateChange)

	ctx.newStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		stateChangeByAddr[addr] = db.AddressStateChange{
			PrevState: convertIdentityState(state.Undefined),
			NewState:  convertIdentityState(identity.State),
		}
	})
	var addresses []*db.Address
	var withZeroWallet bool
	ctx.newStateReadOnly.State.IterateAccounts(func(key []byte, _ []byte) bool {
		if key == nil {
			return true
		}
		addr := conversion.BytesToAddr(key)
		if !withZeroWallet && addr == (common.Address{}) {
			withZeroWallet = true
		}
		address := &db.Address{
			Address: conversion.ConvertAddress(addr),
		}
		if stateChange, ok := stateChangeByAddr[addr]; ok {
			address.StateChanges = []db.AddressStateChange{
				stateChange,
			}
			delete(stateChangeByAddr, addr)
		}
		addresses = append(addresses, address)
		return false
	})
	for addr, stateChange := range stateChangeByAddr {
		address := &db.Address{
			Address: conversion.ConvertAddress(addr),
			StateChanges: []db.AddressStateChange{
				stateChange,
			},
		}
		addresses = append(addresses, address)
	}
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
	prevValidatorsCache *validators.ValidatorsCache,
) (db.Block, error) {
	var txs []db.Transaction
	if len(incomingBlock.Body.Transactions) > 0 {
		txs = indexer.convertTransactions(incomingBlock.Body.Transactions, ctx, collector)
	}

	incomingBlock.Header.Flags()
	proposerVrfScore, _ := getProposerVrfScore(
		incomingBlock,
		indexer.listener.NodeCtx().ProposerByRound,
		indexer.listener.NodeCtx().PendingProofs,
		indexer.secondaryStorage,
		prevValidatorsCache,
	)
	encodedBlock, _ := incomingBlock.ToBytes()
	var upgrade *uint32
	if incomingBlock.Header.ProposedHeader != nil {
		upgrade = &incomingBlock.Header.ProposedHeader.Upgrade
	}
	return db.Block{
		Height:                  incomingBlock.Height(),
		Hash:                    conversion.ConvertHash(incomingBlock.Hash()),
		Time:                    incomingBlock.Header.Time(),
		Transactions:            txs,
		Proposer:                getProposer(incomingBlock),
		Flags:                   convertFlags(incomingBlock.Header.Flags()),
		IsEmpty:                 incomingBlock.IsEmpty(),
		BodySize:                len(incomingBlock.Body.ToBytes()),
		FullSize:                len(encodedBlock),
		OriginalValidatorsCount: len(indexer.statsHolder().GetStats().OriginalFinalCommittee),
		PoolValidatorsCount:     len(indexer.statsHolder().GetStats().PoolFinalCommittee),
		VrfProposerThreshold:    ctx.prevStateReadOnly.State.VrfProposerThreshold(),
		ProposerVrfScore:        proposerVrfScore,
		FeeRate:                 blockchain.ConvertToFloat(ctx.prevStateReadOnly.State.FeePerGas()),
		Upgrade:                 upgrade,
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
	collector *conversionCollector,
) []db.Transaction {
	if len(incomingTxs) == 0 {
		return nil
	}
	var txs []db.Transaction
	for _, incomingTx := range incomingTxs {
		txs = append(txs, indexer.convertTransaction(incomingTx, ctx, collector))
	}
	return txs
}

func (indexer *Indexer) convertTransaction(
	incomingTx *types.Transaction,
	ctx *conversionContext,
	collector *conversionCollector,
) db.Transaction {
	if f, h := detectSubmittedFlip(incomingTx); f != nil {
		collector.submittedFlips = append(collector.submittedFlips, *f)
		collector.flipTxs = append(collector.flipTxs, *h)
	}

	if deletedFlip := detectDeletedFlip(incomingTx); deletedFlip != nil {
		collector.deletedFlips = append(collector.deletedFlips, *deletedFlip)
	}

	if killInviteeTx := detectKillInviteeTx(incomingTx, ctx.prevStateReadOnly); killInviteeTx != nil {
		collector.killInviteeTxs = append(collector.killInviteeTxs, *killInviteeTx)
	}

	if becomeOnlineTxHash, becomeOfflineTxHash := detectOnlineStatusTx(incomingTx); becomeOnlineTxHash != nil {
		collector.becomeOnlineTxs = append(collector.becomeOnlineTxs, *becomeOnlineTxHash)
	} else if becomeOfflineTxHash != nil {
		collector.becomeOfflineTxs = append(collector.becomeOfflineTxs, *becomeOfflineTxHash)
	}

	if incomingTx.Type == types.DelegateTx || incomingTx.Type == types.UndelegateTx {
		collector.switchDelegationTxs = append(collector.switchDelegationTxs, incomingTx)
	}

	indexer.handleLongAnswers(incomingTx, ctx, collector)
	txHash := conversion.ConvertHash(incomingTx.Hash())

	sender, _ := types.Sender(incomingTx)
	from := conversion.ConvertAddress(sender)
	if _, present := collector.addresses[from]; !present {
		collector.addresses[from] = &db.Address{
			Address: from,
		}
	}

	getIdentityStateChange := func(address common.Address) *stats.IdentityStateChange {
		if indexer.statsHolder().GetStats().IdentityStateChangesByTxHashAndAddress == nil {
			return nil
		}
		txChanges, ok := indexer.statsHolder().GetStats().IdentityStateChangesByTxHashAndAddress[incomingTx.Hash()]
		if !ok {
			return nil
		}
		change, ok := txChanges[address]
		if !ok {
			return nil
		}
		return change
	}

	senderStateChange := getIdentityStateChange(sender)
	if senderStateChange != nil {
		if incomingTx.Type == types.ActivationTx && senderStateChange.NewState == state.Killed {
			collector.addresses[from].IsTemporary = true
		}
		collector.addresses[from].StateChanges = append(collector.addresses[from].StateChanges,
			db.AddressStateChange{
				PrevState: convertIdentityState(senderStateChange.PrevState),
				NewState:  convertIdentityState(senderStateChange.NewState),
				TxHash:    txHash,
			})

		if senderStateChange.NewState == state.Killed {
			collector.killedAddrs[sender] = struct{}{}
		} else {
			delete(collector.killedAddrs, sender)
		}
	}

	var to string
	if incomingTx.To != nil {
		to = conversion.ConvertAddress(*incomingTx.To)
		if _, present := collector.addresses[to]; !present {
			collector.addresses[to] = &db.Address{
				Address: to,
			}
		}
		if *incomingTx.To != sender {
			recipientStateChange := getIdentityStateChange(*incomingTx.To)
			if recipientStateChange != nil {
				collector.addresses[to].StateChanges = append(collector.addresses[to].StateChanges,
					db.AddressStateChange{
						PrevState: convertIdentityState(recipientStateChange.PrevState),
						NewState:  convertIdentityState(recipientStateChange.NewState),
						TxHash:    txHash,
					})

				if recipientStateChange.NewState == state.Killed {
					collector.killedAddrs[*incomingTx.To] = struct{}{}
				} else {
					delete(collector.killedAddrs, *incomingTx.To)
				}
			}
		}
	}

	txRaw, err := incomingTx.ToBytes()
	if err != nil {
		log.Error("Unable to convert tx to bytes", "tx", txHash, "err", err)
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
		Fee:     blockchain.ConvertToFloat(indexer.statsHolder().GetStats().FeesByTxHash[incomingTx.Hash()]),
		Size:    incomingTx.Size(),
		Raw:     hex.EncodeToString(txRaw),
	}

	return tx
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
		Address: conversion.ConvertAddress(incomingAnswer.Respondent),
		Answer:  convertAnswer(incomingAnswer.Answer),
		Point:   incomingAnswer.Point,
		Grade:   byte(incomingAnswer.Grade),
	}
}

func convertCid(cid cid.Cid) string {
	return cid.String()
}

func (indexer *Indexer) detectEpochResult(block *types.Block, ctx *conversionContext, collector *conversionCollector) *db.EpochResult {
	if !block.Header.Flags().HasFlag(types.ValidationFinished) {
		return nil
	}

	var birthdays []db.Birthday
	var identities []db.EpochIdentity
	var validationRewardsSummaries []db.ValidationRewardSummaries
	var vrsCalculator *validationRewardSummariesCalculator
	if indexer.statsHolder().GetStats().RewardsStats != nil {
		vrsCalculator = newValidationRewardSummariesCalculator(
			indexer.statsHolder().GetStats().RewardsStats,
		)
	}
	memPoolFlipKeysToMigrate := indexer.getMemPoolFlipKeysToMigrate(ctx.prevStateReadOnly.State.Epoch())
	memPoolFlipKeys := memPoolFlipKeysToMigrate
	validationStats := indexer.statsHolder().GetStats().ValidationStats

	authorAddressesByFlipCid := make(map[string]string)

	rewardsBounds := &rewardsBounds{}

	var totalRewardsByAddr map[common.Address]*big.Int
	if indexer.statsHolder().GetStats().RewardsStats != nil {
		totalRewardsByAddr = indexer.statsHolder().GetStats().RewardsStats.TotalRewardsByAddr
	}

	godAddress := ctx.prevStateReadOnly.State.GodAddress()
	newEpoch := ctx.newStateReadOnly.State.Epoch()
	epochRewards, validationRewardsAddresses, delegateesEpochRewards := indexer.detectEpochRewards(block)
	ctx.prevStateReadOnly.State.IterateOverIdentities(func(addr common.Address, identity state.Identity) {
		shardId := identity.ShiftedShardId()
		convertedAddress := conversion.ConvertAddress(addr)
		convertedIdentity := db.EpochIdentity{}
		convertedIdentity.ShardId = shardId
		convertedIdentity.NewShardId = ctx.newStateReadOnly.State.ShardId(addr)
		convertedIdentity.Address = convertedAddress
		newIdentityState := ctx.newStateReadOnly.State.GetIdentityState(addr)
		convertedIdentity.State = convertIdentityState(newIdentityState)
		convertedIdentity.RequiredFlips = ctx.prevStateReadOnly.State.GetRequiredFlips(addr)
		identityPrevState := ctx.prevStateReadOnly.State.GetIdentity(addr)
		convertedIdentity.AvailableFlips = identityPrevState.GetMaximumAvailableFlips()
		convertedIdentity.MadeFlips = ctx.prevStateReadOnly.State.GetMadeFlips(addr)
		convertedIdentity.NextEpochInvites = ctx.newStateReadOnly.State.GetInvites(addr)
		if identity.Delegatee != nil {
			convertedIdentity.DelegateeAddress = conversion.ConvertAddress(*identity.Delegatee)
		}
		validationShardStats := validationStats.Shards[shardId]
		var identityStats *statsTypes.IdentityStats
		var present bool
		if validationShardStats != nil {
			identityStats, present = validationShardStats.IdentitiesPerAddr[addr]
		}
		if validationShardStats != nil && present && identityStats != nil {
			convertedIdentity.ShortPoint = identityStats.ShortPoint
			convertedIdentity.ShortFlips = identityStats.ShortFlips

			if newIdentityState.NewbieOrBetter() || newIdentityState == state.Suspended || newIdentityState == state.Zombie {
				convertedIdentity.TotalShortPoint, convertedIdentity.TotalShortFlips = common.CalculateIdentityScores(ctx.newStateReadOnly.State.GetScores(addr),
					ctx.newStateReadOnly.State.GetShortFlipPoints(addr), ctx.newStateReadOnly.State.GetQualifiedFlipsCount(addr))
			} else if identityStats.Missed {
				convertedIdentity.TotalShortPoint, convertedIdentity.TotalShortFlips = common.CalculateIdentityScores(identity.Scores,
					ctx.prevStateReadOnly.State.GetShortFlipPoints(addr), ctx.prevStateReadOnly.State.GetQualifiedFlipsCount(addr))
			} else {
				convertedIdentity.TotalShortPoint, convertedIdentity.TotalShortFlips = calculateNewTotalScore(identity.Scores,
					identityStats.ShortPoint, identityStats.ShortFlips, ctx.prevStateReadOnly.State.GetShortFlipPoints(addr),
					ctx.prevStateReadOnly.State.GetQualifiedFlipsCount(addr))
			}

			convertedIdentity.LongPoint = identityStats.LongPoint
			convertedIdentity.LongFlips = identityStats.LongFlips
			convertedIdentity.Approved = identityStats.Approved
			convertedIdentity.Missed = identityStats.Missed
			convertedIdentity.ShortFlipCidsToSolve = convertCids(identityStats.ShortFlipsToSolve, validationShardStats.FlipCids, block)
			convertedIdentity.LongFlipCidsToSolve = convertCids(identityStats.LongFlipsToSolve, validationShardStats.FlipCids, block)
		} else {
			convertedIdentity.Approved = false
			convertedIdentity.Missed = true
			convertedIdentity.TotalShortPoint, convertedIdentity.TotalShortFlips = common.CalculateIdentityScores(identity.Scores,
				ctx.prevStateReadOnly.State.GetShortFlipPoints(addr), ctx.prevStateReadOnly.State.GetQualifiedFlipsCount(addr))
		}

		if identityPrevState.State == state.Invite || identityPrevState.State == state.Candidate {
			convertedIdentity.BirthEpoch = uint64(ctx.prevStateReadOnly.State.Epoch())
		} else {
			convertedIdentity.BirthEpoch = uint64(identityPrevState.Birthday)
		}

		identities = append(identities, convertedIdentity)
		delete(validationRewardsAddresses, addr)

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

		if totalRewardsByAddr != nil && addr != godAddress {
			reward, ok := totalRewardsByAddr[addr]
			if ok {
				age := uint64(newEpoch) - convertedIdentity.BirthEpoch
				rewardsBounds.addIfBound(addr, age, reward)
			}
		}

		if identity.State != state.Undefined && identity.State != state.Killed && (newIdentityState == state.Killed || newIdentityState == state.Undefined) {
			collector.killedAddrs[addr] = struct{}{}
		}

		if vrsCalculator != nil {
			var potentialValidationRewardAge uint16
			if identity.State == state.Undefined || identity.State == state.Killed || identity.State == state.Invite || identity.State == state.Candidate {
				potentialValidationRewardAge = 1
			} else {
				potentialValidationRewardAge = newEpoch - uint16(convertedIdentity.BirthEpoch)
			}
			validationRewardSummaries := vrsCalculator.calculateValidationRewardSummaries(
				addr,
				shardId,
				potentialValidationRewardAge,
				identity.Flips,
				newIdentityState,
				convertedIdentity.AvailableFlips,
			)
			validationRewardsSummaries = append(validationRewardsSummaries, validationRewardSummaries)
		}
	})

	for addr := range validationRewardsAddresses {
		identities = append(identities, db.EpochIdentity{
			Address: conversion.ConvertAddress(addr),
			State:   convertIdentityState(state.Undefined),
			Missed:  true,
		})
	}

	var flipsStats []db.FlipStats
	flipStatusesMap := make(map[byte]uint64)
	var reportedFlips uint32
	for _, validationShardStats := range validationStats.Shards {
		for flipIdx, flipStats := range validationShardStats.FlipsPerIdx {
			flipCid, err := cid.Parse(validationShardStats.FlipCids[flipIdx])
			if err != nil {
				log.Error("Unable to parse flip cid. Skipped.", "b", block.Height(), "idx", flipIdx, "err", err)
				continue
			}
			if flipStats.Grade == types.GradeReported {
				reportedFlips++
			}
			flipCidStr := convertCid(flipCid)
			flipStats := db.FlipStats{
				Author:       authorAddressesByFlipCid[flipCidStr],
				Cid:          flipCidStr,
				ShortAnswers: convertStatsAnswers(flipStats.ShortAnswers),
				LongAnswers:  convertStatsAnswers(flipStats.LongAnswers),
				Status:       convertFlipStatus(ceremony.FlipStatus(flipStats.Status)),
				Answer:       convertAnswer(flipStats.Answer),
				Grade:        byte(flipStats.Grade),
			}
			flipsStats = append(flipsStats, flipStats)
			flipStatusesMap[flipStats.Status]++
		}
	}
	flipStatuses := make([]db.FlipStatusCount, 0, len(flipStatusesMap))
	for status, count := range flipStatusesMap {
		flipStatuses = append(flipStatuses, db.FlipStatusCount{
			Status: status,
			Count:  count,
		})
	}

	collectorStats := indexer.statsHolder().GetStats()
	var minScoreForInvite float32 = 0
	if collectorStats.MinScoreForInvite != nil {
		minScoreForInvite = *collectorStats.MinScoreForInvite
	}

	return &db.EpochResult{
		Identities:                identities,
		FlipStats:                 flipsStats,
		FlipStatuses:              flipStatuses,
		Birthdays:                 birthdays,
		MemPoolFlipKeys:           memPoolFlipKeys,
		FailedValidation:          validationStats.Failed,
		EpochRewards:              epochRewards,
		MinScoreForInvite:         minScoreForInvite,
		RewardsBounds:             rewardsBounds.getResult(),
		ReportedFlips:             reportedFlips,
		DelegateesEpochRewards:    delegateesEpochRewards,
		ValidationRewardSummaries: validationRewardsSummaries,
	}
}

func calculateNewTotalScore(scores []byte, shortPoints float32, shortFlipsCount uint32, totalShortPoints float32, totalShortFlipsCount uint32) (totalPoints float32, totalFlips uint32) {
	newScores := make([]byte, len(scores))
	copy(newScores, scores)
	newScores = append(newScores, common.EncodeScore(shortPoints, shortFlipsCount))
	if len(newScores) > common.LastScoresCount {
		newScores = newScores[len(newScores)-common.LastScoresCount:]
	}
	return common.CalculateIdentityScores(newScores, totalShortPoints, totalShortFlipsCount)
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

func (indexer *Indexer) handleLongAnswers(
	tx *types.Transaction,
	ctx *conversionContext,
	collector *conversionCollector,
) {
	if tx.Type != types.SubmitLongAnswersTx {
		return
	}
	attachment := attachments.ParseLongAnswerAttachment(tx)
	if attachment == nil {
		log.Error("Unable to parse long answers payload. Skipped.", "tx", tx.Hash())
		return
	}

	sender, _ := types.Sender(tx)

	if len(attachment.Key) > 0 {
		collector.flipKeys = append(collector.flipKeys, db.FlipKey{
			TxHash: conversion.ConvertHash(tx.Hash()),
			Key:    hex.EncodeToString(attachment.Key),
		})
	}

	if len(attachment.Proof) > 0 {
		for _, f := range ctx.prevStateReadOnly.State.GetIdentity(sender).Flips {
			firstIndex, dictionarySize := indexer.listener.NodeCtx().Ceremony.GetWordDictionaryRange()
			word1, word2, err := getFlipWords(sender, attachment, firstIndex, dictionarySize, int(f.Pair), ctx.prevStateReadOnly)
			flipCid, _ := cid.Parse(f.Cid)
			cidStr := convertCid(flipCid)
			if err != nil {
				log.Error("Unable to get flip words. Skipped.", "tx", tx.Hash(), "cid", cidStr, "err", err)
				continue
			}

			collector.flipsWords = append(collector.flipsWords, db.FlipWords{
				Cid:    cidStr,
				TxHash: conversion.ConvertHash(tx.Hash()),
				Word1:  uint16(word1),
				Word2:  uint16(word2),
			})
		}
	} else {
		log.Error("Empty proof for flip words. Skipped.", "tx", tx.Hash())
	}
}

func getFlipWords(addr common.Address, attachment *attachments.LongAnswerAttachment, firstIndex, dictionarySize, pairId int, appState *appstate.AppState) (word1, word2 int, err error) {
	identity := appState.State.GetIdentity(addr)
	hash, _ := vrf.HashFromProof(attachment.Proof)
	rnd := binary.LittleEndian.Uint64(hash[:])
	return ceremony.GetWords(rnd, firstIndex, dictionarySize, identity.GetTotalWordPairsCount(), pairId)
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
			log.Error(fmt.Sprintf("Unable to save block %d data: %v", data.Block.Height, err))
			indexer.waitForRetry()
			continue
		}
		return
	}
}

func (indexer *Indexer) applyOnState(data *result) {
	indexer.state.lastIndexedHeight = data.dbData.Block.Height
	indexer.state.totalBalance = data.resData.totalBalance
	indexer.state.totalStake = data.resData.totalStake
}

func (indexer *Indexer) waitForRetry() {
	time.Sleep(requestRetryInterval)
}
