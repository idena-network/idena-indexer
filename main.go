package main

import (
	"fmt"
	"github.com/idena-network/idena-go/core/state"
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/core/flip"
	"github.com/idena-network/idena-indexer/core/holder/online"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
	"github.com/idena-network/idena-indexer/core/server"
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

		// Explorer
		explorerConf := explorerConfig.LoadConfig(context.String("explorerConfig"))
		e := explorer.NewExplorer(explorerConf)
		defer e.Destroy()

		// Indexer
		indxr, listener := initIndexer(conf)
		defer indxr.Destroy()
		indxr.Start()

		// Server for explorer & indexer api
		currentOnlineIdentitiesHolder := online.NewCurrentOnlineIdentitiesCache(listener.AppState(),
			listener.NodeCtx().Blockchain,
			listener.NodeCtx().OfflineDetector)

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

func initIndexer(config *config.Config) (*indexer.Indexer, incoming.Listener) {
	performanceMonitor := initPerformanceMonitor(config.PerformanceMonitor)
	wordsLoader := words.NewLoader(config.WordsFile)
	listener := incoming.NewListener(config.NodeConfigFile, performanceMonitor)
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

	return indexer.NewIndexer(
			listener,
			memPoolIndexer,
			dbAccessor,
			restorer,
			secondaryStorage,
			restoreInitially,
			performanceMonitor,
			flipLoader,
		),
		listener
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
