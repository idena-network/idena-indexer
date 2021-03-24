package main

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/events"
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/core/flip"
	"github.com/idena-network/idena-indexer/core/holder/online"
	"github.com/idena-network/idena-indexer/core/holder/transaction"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/explorer"
	explorerConfig "github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/import/words"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/idena-network/idena-indexer/log"
	migrationDb "github.com/idena-network/idena-indexer/migration/db"
	runtimeMigration "github.com/idena-network/idena-indexer/migration/runtime"
	runtimeMigrationDb "github.com/idena-network/idena-indexer/migration/runtime/db"
	"github.com/idena-network/idena-indexer/monitoring"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
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
			Name:  "explorerConfig",
			Usage: "Explorer config file",
			Value: filepath.Join("conf", "explorer.json"),
		},
		cli.StringFlag{
			Name:  "indexerConfig",
			Usage: "Indexer config file",
			Value: filepath.Join("conf", "indexer.json"),
		},
	}

	app.Action = func(context *cli.Context) error {

		conf := config.LoadConfig(context.String("indexerConfig"))
		initLog(conf.Verbosity, conf.NodeVerbosity)
		log.Info("Starting app...")

		txMemPool := transaction.NewMemPool(log.New("component", "txMemPool"))

		// Indexer
		indxr, listener, contractsMemPool := initIndexer(conf, txMemPool)
		defer indxr.Destroy()

		// Explorer
		explorerConf := explorerConfig.LoadConfig(context.String("explorerConfig"))
		networkSizeLoader := &networkSizeLoader{}
		e := explorer.NewExplorer(explorerConf, txMemPool, contractsMemPool, networkSizeLoader)
		defer e.Destroy()

		// Start indexer
		indxr.Start()

		// Server for explorer & indexer api
		currentOnlineIdentitiesHolder := online.NewCurrentOnlineIdentitiesCache(listener.AppState(),
			listener.NodeCtx().Blockchain,
			listener.NodeCtx().OfflineDetector)
		networkSizeLoader.holder = currentOnlineIdentitiesHolder

		upgradesVoting := upgrade.NewUpgradesVotingHolder(listener.NodeCtx().Upgrader)

		explorerRi := e.RouterInitializer()
		indexerApi := api.NewApi(currentOnlineIdentitiesHolder, upgradesVoting)
		ownRi := server.NewRouterInitializer(indexerApi, e.Logger())

		description, err := ioutil.ReadFile(filepath.Join(explorerConf.HtmlDir, "api.html"))
		if err != nil {
			panic(fmt.Sprintf("Unable to initialize api description: %v", err))
		}

		apiServer := server.NewServer(
			explorerConf.Port,
			explorerConf.MaxReqCount,
			time.Second*time.Duration(explorerConf.ReqTimeoutSec),
			e.Logger(),
			explorerConf.ReqsPerMinuteLimit,
			description,
		)
		go apiServer.Start(explorerConf.Swagger, explorerRi, ownRi)

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

func initIndexer(config *config.Config, txMemPool transaction.MemPool) (*indexer.Indexer, incoming.Listener, mempool.Contracts) {
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
			contractsMemPoolBus.Publish(newTxEvent)
			tx := newTxEvent.Tx
			if err := txMemPool.AddTransaction(tx); err != nil {
				log.Warn("Unable to add new tx to tx mem pool", "hash", tx.Hash().Hex())
			}
		})

	performanceMonitor := initPerformanceMonitor(config.PerformanceMonitor)
	wordsLoader := words.NewLoader(config.WordsFile)
	statsCollector := stats.NewStatsCollector(statsCollectorEventBus)
	listener := incoming.NewListener(config.NodeConfigFile, nodeEventBus, statsCollector, performanceMonitor)
	dbAccessor := db.NewPostgresAccessor(config.Postgres.ConnStr, config.Postgres.ScriptsDir, wordsLoader,
		performanceMonitor, config.CommitteeRewardBlocksCount, config.MiningRewards)
	restorer := restore.NewRestorer(dbAccessor, listener.AppState(), listener.NodeCtx().Blockchain)
	var secondaryStorage *runtimeMigration.SecondaryStorage
	if config.RuntimeMigration.Enabled {
		secondaryStorage = runtimeMigration.NewSecondaryStorage(runtimeMigrationDb.NewPostgresAccessor(
			config.RuntimeMigration.Postgres.ConnStr, config.RuntimeMigration.Postgres.ScriptsDir))
	}
	restoreInitially := config.RestoreInitially
	if migrated, err := migrateDataIfNeeded(config); err != nil {
		panic(fmt.Sprintf("Unable to migrate data: %v", err))
	} else {
		restoreInitially = restoreInitially || migrated
	}

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
	contractsMemPool := mempool.NewContracts(listener.NodeCtx().AppState, listener.NodeCtx().Blockchain, listener.Config(), contractsMemPoolLogger)
	contractsMemPoolBus.Subscribe(events.NewTxEventID, func(e eventbus.Event) {
		newTxEvent := e.(*events.NewTxEvent)
		if err := contractsMemPool.ProcessTx(newTxEvent.Tx); err != nil {
			contractsMemPoolLogger.Error(fmt.Sprintf("Unable to process tx: %v", err))
		}
	})
	contractsMemPoolBus.Subscribe(removedMemPoolTxEventId, func(e eventbus.Event) {
		contractsMemPool.RemoveTx(e.(*removedMemPoolTxEvent).tx)
	})

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
		),
		listener, contractsMemPool
}

func initPerformanceMonitor(config config.PerformanceMonitorConfig) monitoring.PerformanceMonitor {
	if !config.Enabled {
		return monitoring.NewEmptyPerformanceMonitor()
	}
	return monitoring.NewPerformanceMonitor(config.BlocksToLog, log.New("component", "pm"))
}

func migrateDataIfNeeded(config *config.Config) (bool, error) {
	if !config.Migration.Enabled {
		return false, nil
	}
	dbAccessor := migrationDb.NewPostgresAccessor(config.Postgres.ConnStr, config.Migration.OldSchema,
		config.Migration.ScriptsDir)
	defer dbAccessor.Destroy()
	log.Info("Start migrating data...")
	if err := dbAccessor.MigrateTo(config.Migration.Height); err != nil {
		return false, err
	}
	log.Info("Data migration has been completed")
	return true, nil
}

type networkSizeLoader struct {
	holder online.CurrentOnlineIdentitiesHolder
}

func (l *networkSizeLoader) Load() (uint64, error) {
	if l.holder == nil {
		return 0, errors.New("loader has not been initialized")
	}
	return uint64(len(l.holder.GetAll())), nil
}
