package main

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	config2 "github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/database"
	"github.com/idena-network/idena-go/events"
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/core/flip"
	"github.com/idena-network/idena-indexer/core/holder/contract"
	"github.com/idena-network/idena-indexer/core/holder/online"
	state2 "github.com/idena-network/idena-indexer/core/holder/state"
	"github.com/idena-network/idena-indexer/core/holder/transaction"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	logUtil "github.com/idena-network/idena-indexer/core/log"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/data"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/import/words"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/idena-network/idena-indexer/log"
	runtimeMigration "github.com/idena-network/idena-indexer/migration/runtime"
	runtimeMigrationDb "github.com/idena-network/idena-indexer/migration/runtime/db"
	"github.com/idena-network/idena-indexer/migration/tokens"
	"github.com/idena-network/idena-indexer/monitoring"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// @license.name Apache 2.0
func main() {
	app := cli.NewApp()
	app.Name = "github.com/idena-network/idena-indexer"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "Config file",
			Value: filepath.Join("conf", "config.json"),
		},
	}

	app.Action = func(context *cli.Context) error {

		conf := config.LoadConfig(context.String("config"))
		initLog(conf.Verbosity, conf.NodeVerbosity)
		log.Info("Starting app...")

		txMemPool := transaction.NewMemPool(log.New("component", "txMemPool"))

		// Indexer
		indxr, listener, contractsMemPool, upgradesVoting := initIndexer(conf, txMemPool)
		defer indxr.Destroy()

		// Start indexer
		indxr.Start()

		// Server for explorer & indexer api
		currentOnlineIdentitiesHolder := online.NewCurrentOnlineIdentitiesCache(listener.AppState(),
			listener.NodeCtx().Blockchain,
			listener.NodeCtx().OfflineDetector)

		apiLogger, err := logUtil.NewFileLogger("api.log", conf.Api.LogFileSize)
		if err != nil {
			panic(err)
		}
		appStateHolder := state2.NewAppStateHolder(listener.NodeCtx().AppState, listener.NodeCtx().Blockchain)
		contractHolder := contract.NewHolder(appStateHolder)
		indexerApi := api.NewApi(currentOnlineIdentitiesHolder, upgradesVoting, txMemPool, contractsMemPool,
			state2.NewHolder(conf.TreeSnapshotDir, log.New("component", "stateHolder")), contractHolder)
		ownRi := server.NewRouterInitializer(indexerApi, apiLogger)

		apiServer := server.NewServer(conf.Api.Port, apiLogger)
		go apiServer.Start(ownRi)

		indxr.WaitForNodeStop()

		return nil
	}

	app.Run(os.Args)
}

func initLog(verbosity int, nodeVerbosity int) {
	logLvl := log.Lvl(verbosity)
	nodeLogLvl := nodeLog.Lvl(nodeVerbosity)
	if runtime.GOOS == "windows" {
		log.Root().SetHandler(log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stdout, log.LogfmtFormat())))
		nodeLog.Root().SetHandler(nodeLog.LvlFilterHandler(nodeLogLvl, nodeLog.StreamHandler(os.Stdout,
			nodeLog.LogfmtFormat())))
	} else {
		log.Root().SetHandler(log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stderr,
			log.TerminalFormat(true))))
		nodeLog.Root().SetHandler(nodeLog.LvlFilterHandler(nodeLogLvl, nodeLog.StreamHandler(os.Stderr,
			nodeLog.TerminalFormat(true))))
	}
}

const removedMemPoolTxEventId = eventbus.EventID("removed-mem-pool-tx")

type removedMemPoolTxEvent struct {
	tx *types.Transaction
}

func (removedMemPoolTxEvent) EventID() eventbus.EventID {
	return removedMemPoolTxEventId
}

func initIndexer(config *config.Config, txMemPool transaction.MemPool) (*indexer.Indexer, incoming.Listener, mempool.Contracts, upgrade.UpgradesVotingHolder) {
	indexerEventBus := eventbus.New()
	contractsMemPoolBus := eventbus.New()
	statsCollectorEventBus := eventbus.New()
	statsCollectorEventBus.Subscribe(stats.RemovedMemPoolTxEventID, func(e eventbus.Event) {
		tx := e.(*stats.RemovedMemPoolTxEvent).Tx
		contractsMemPoolBus.Publish(&removedMemPoolTxEvent{tx: tx})
		if err := txMemPool.RemoveTransaction(tx); err != nil {
			log.Warn("Unable to remove tx from tx mem pool", "hash", tx.Hash().Hex())
		}
	})

	nodeEventBus := eventbus.New()
	nodeEventBus.Subscribe(events.NewTxEventID,
		func(e eventbus.Event) {
			newTxEvent := e.(*events.NewTxEvent)
			if newTxEvent.Deferred {
				return
			}
			contractsMemPoolBus.Publish(newTxEvent)
			tx := newTxEvent.Tx
			if err := txMemPool.AddTransaction(tx); err != nil {
				log.Warn("Unable to add new tx to tx mem pool", "hash", tx.Hash().Hex())
			}
		})

	performanceMonitor := initPerformanceMonitor(config.PerformanceMonitor)
	wordsLoader := words.NewLoader(config.WordsFile)
	nodeConfig := initNodeConfig(config.NodeConfigFile)
	tokenContractHolder := new(stats.TokenContractHolderImpl)
	statsCollector := stats.NewStatsCollector(statsCollectorEventBus, nodeConfig.Consensus, tokenContractHolder)
	listener := incoming.NewListener(nodeConfig, nodeEventBus, statsCollector, performanceMonitor)
	tokenContractHolder.ProvideNodeCtx(listener.NodeCtx(), nodeConfig)
	var dataTable, dataStateTable string
	if config.Data != nil && config.Data.Enabled {
		dataTable, dataStateTable = config.Data.Table, config.Data.StateTable
	}
	dbAccessor := db.NewPostgresAccessor(config.Postgres.ConnStr, config.Postgres.ScriptsDir, wordsLoader,
		performanceMonitor, config.CommitteeRewardBlocksCount, config.MiningRewards, dataTable, dataStateTable)
	restorer := restore.NewRestorer(dbAccessor, listener.AppState(), listener.NodeCtx().Blockchain)
	var secondaryStorage *runtimeMigration.SecondaryStorage
	if config.RuntimeMigration.Enabled {
		secondaryStorage = runtimeMigration.NewSecondaryStorage(runtimeMigrationDb.NewPostgresAccessor(
			config.RuntimeMigration.Postgres.ConnStr, config.RuntimeMigration.Postgres.ScriptsDir))
	}
	restoreInitially := config.RestoreInitially

	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))

	flipLoaderLogger := log.New("component", "flipLoader")
	flipLoader := flip.NewLoader(
		func() uint64 {
			return uint64(listener.AppState().State.Epoch())
		},
		func() bool {
			head := listener.NodeCtx().Blockchain.Head
			if head == nil {
				return false
			}
			prevState, err := listener.AppState().Readonly(head.Height() - 1)
			if err != nil {
				flipLoaderLogger.Error(fmt.Sprintf("Unable to get app state for height %d, err %v", head.Height()-1, err))
				return false
			}
			return prevState.State.ValidationPeriod() < state.FlipLotteryPeriod
		},
		dbAccessor, listener.Flipper(), flipLoaderLogger)

	flip.StartContentLoader(
		dbAccessor,
		config.FlipContentLoader.BatchSize,
		config.FlipContentLoader.AttemptsLimit,
		time.Minute*time.Duration(config.FlipContentLoader.RetryIntervalMin),
		listener.Flipper(),
		log.New("component", "flipContentLoader"),
	)

	contractsMemPoolLogger := log.New("component", "contractsMemPool")
	contractsMemPool := mempool.NewContracts(listener.NodeCtx().AppState, listener.NodeCtx().Blockchain, listener.Config(), contractsMemPoolLogger, tokenContractHolder)
	contractsMemPoolBus.Subscribe(events.NewTxEventID, func(e eventbus.Event) {
		newTxEvent := e.(*events.NewTxEvent)
		if err := contractsMemPool.ProcessTx(newTxEvent.Tx); err != nil {
			contractsMemPoolLogger.Error(fmt.Sprintf("Unable to process tx: %v", err))
		}
	})
	contractsMemPoolBus.Subscribe(removedMemPoolTxEventId, func(e eventbus.Event) {
		contractsMemPool.RemoveTx(e.(*removedMemPoolTxEvent).tx)
	})

	upgradesVoting := upgrade.NewUpgradesVotingHolder(listener.NodeCtx().Upgrader)

	peersTracker := mempool.NewPeersTracker(dbAccessor, log.New("component", "peersTracker"))
	nodeEventBus.Subscribe(events.PeersEventID, func(e eventbus.Event) {
		peersEvent := e.(*events.PeersEvent)
		peersTracker.AddPeersData(peersEvent.PeersData, peersEvent.Time)
	})

	if config.VoteCounting.Enabled {
		voteCountingTracker := mempool.NewVoteCountingTracker(dbAccessor, log.New("component", "voteCountingTracker"))
		statsCollectorEventBus.Subscribe(stats.VoteCountingStepResultEventID, func(e eventbus.Event) {
			voteCountingTracker.SubmitVoteCountingStepResultEvent(e.(*stats.VoteCountingStepResultEvent))
		})
		statsCollectorEventBus.Subscribe(stats.VoteCountingResultEventID, func(e eventbus.Event) {
			voteCountingTracker.SubmitVoteCountingResultEvent(e.(*stats.VoteCountingResultEvent))
		})
		statsCollectorEventBus.Subscribe(stats.ProofProposalEventID, func(e eventbus.Event) {
			voteCountingTracker.SubmitProofProposalEvent(e.(*stats.ProofProposalEvent))
		})
		statsCollectorEventBus.Subscribe(stats.BlockProposalEventID, func(e eventbus.Event) {
			voteCountingTracker.SubmitBlockProposalEvent(e.(*stats.BlockProposalEvent))
		})
	}

	if config.Data != nil && config.Data.Enabled {
		data.StartDataService(indexerEventBus, dbAccessor, log.New("component", "dataEngine"))
	}

	func() {
		if listener.NodeCtx().Blockchain.GetHead() == nil {
			return
		}
		appState, err := listener.AppState().ForCheckWithOverwrite(listener.NodeCtx().Blockchain.GetHead().Height())
		if err != nil {
			panic(errors.Wrap(err, "failed to initialize app state for token balances"))
		}
		if err := tokens.CollectData(config.Postgres.ConnStr, tokenContractHolder, appState); err != nil {
			panic(errors.Wrap(err, "failed to initialize token balances"))
		}
	}()

	enabled := config.Enabled == nil || *config.Enabled
	return indexer.NewIndexer(
			enabled,
			listener,
			memPoolIndexer,
			dbAccessor,
			restorer,
			secondaryStorage,
			restoreInitially,
			performanceMonitor,
			flipLoader,
			upgradesVoting,
			config.UpgradeVotingShortHistoryItems,
			config.UpgradeVotingShortHistoryMinShift,
			indexerEventBus,
			config.TreeSnapshotDir,
			dbAccessor,
			indexer.NewOracleVotingToProlongDetector(),
			config.CheckBalances,
		),
		listener, contractsMemPool, upgradesVoting
}

func initPerformanceMonitor(config config.PerformanceMonitorConfig) monitoring.PerformanceMonitor {
	if !config.Enabled {
		return monitoring.NewEmptyPerformanceMonitor()
	}
	return monitoring.NewPerformanceMonitor(config.BlocksToLog, log.New("component", "pm"))
}

func initNodeConfig(nodeConfigFile string) *config2.Config {
	cfg, err := config2.MakeConfigFromFile(nodeConfigFile)
	if err != nil {
		panic(err)
	}
	cfg.P2P.MaxInboundPeers = config2.LowPowerMaxInboundNotOwnShardPeers
	cfg.P2P.MaxOutboundPeers = config2.LowPowerMaxOutboundNotOwnShardPeers
	cfg.P2P.MaxInboundOwnShardPeers = config2.LowPowerMaxInboundOwnShardPeers
	cfg.P2P.MaxOutboundOwnShardPeers = config2.LowPowerMaxOutboundOwnShardPeers
	cfg.IpfsConf.LowWater = 8
	cfg.IpfsConf.HighWater = 10
	cfg.IpfsConf.GracePeriod = "30s"
	cfg.IpfsConf.ReproviderInterval = "0"
	cfg.IpfsConf.Routing = "dhtclient"
	cfg.IpfsConf.PublishPeers = true
	cfg.Sync.FastSync = false
	cfg.Mempool.ResetInCeremony = true

	cfgTransform(cfg)
	return cfg
}

func cfgTransform(cfg *config2.Config) {
	db, err := node.OpenDatabase(cfg.DataDir, "idenachain", 16, 16, false)
	if err != nil {
		log.Error("Cannot transform consensus config", "err", err)
		return
	}
	defer db.Close()
	repo := database.NewRepo(db)
	consVersion := repo.ReadConsensusVersion()
	if consVersion <= uint32(cfg.Consensus.Version) {
		return
	}
	for v := cfg.Consensus.Version + 1; v <= config2.ConsensusVerson(consVersion); v++ {
		config2.ApplyConsensusVersion(v, cfg.Consensus)
		log.Info("Consensus config transformed to", "ver", v)
	}
}
