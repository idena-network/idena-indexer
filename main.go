package main

import (
	"fmt"
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/core/holder/online"
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
	"github.com/idena-network/idena-indexer/migration/flip"
	flipMigrationDb "github.com/idena-network/idena-indexer/migration/flip/db"
	"github.com/idena-network/idena-indexer/monitoring"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	app := cli.NewApp()
	app.Name = "github.com/idena-network/idena-indexer"
	app.Version = "0.0.1"

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
			listener.Blockchain(),
			listener.OfflineDetector())

		explorerRi := e.RouterInitializer()
		indexerApi := api.NewApi(currentOnlineIdentitiesHolder)
		ownRi := server.NewRouterInitializer(indexerApi, e.Logger())
		apiServer := server.NewServer(explorerConf.Port, e.Logger())
		go apiServer.Start(explorerRi, ownRi)

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
		nodeLog.Root().SetHandler(nodeLog.LvlFilterHandler(nodeLogLvl, nodeLog.StreamHandler(os.Stdout, nodeLog.LogfmtFormat())))
	} else {
		log.Root().SetHandler(log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
		nodeLog.Root().SetHandler(nodeLog.LvlFilterHandler(nodeLogLvl, nodeLog.StreamHandler(os.Stderr, nodeLog.TerminalFormat(true))))
	}
}

func initIndexer(config *config.Config) (*indexer.Indexer, incoming.Listener) {
	performanceMonitor := initPerformanceMonitor(config.PerformanceMonitor)
	wordsLoader := words.NewLoader(config.WordsFile)
	listener := incoming.NewListener(config.NodeConfigFile, performanceMonitor)
	dbAccessor := db.NewPostgresAccessor(config.Postgres.ConnStr, config.Postgres.ScriptsDir, wordsLoader, performanceMonitor)
	restorer := restore.NewRestorer(dbAccessor, listener.AppState(), listener.Blockchain())
	var sfs *flip.SecondaryFlipStorage
	if config.FlipMigrationPostgres != nil {
		sfs = flip.NewSecondaryFlipStorage(flipMigrationDb.NewPostgresAccessor(config.FlipMigrationPostgres.ConnStr, config.FlipMigrationPostgres.ScriptsDir))
	}
	restoreInitially := config.RestoreInitially
	if migrated, err := migrateDataIfNeeded(config); err != nil {
		panic(fmt.Sprintf("Unable to migrate data: %v", err))
	} else {
		restoreInitially = restoreInitially || migrated
	}

	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))

	return indexer.NewIndexer(listener,
			memPoolIndexer,
			dbAccessor,
			restorer,
			sfs,
			uint64(config.GenesisBlockHeight),
			restoreInitially,
			performanceMonitor),
		listener
}

func initPerformanceMonitor(config *config.PerformanceMonitorConfig) monitoring.PerformanceMonitor {
	if config == nil || !config.Enabled {
		return monitoring.NewEmptyPerformanceMonitor()
	}
	return monitoring.NewPerformanceMonitor(config.BlocksToLog, log.New("component", "pm"))
}

func migrateDataIfNeeded(config *config.Config) (bool, error) {
	if config.Migration == nil {
		return false, nil
	}
	dbAccessor := migrationDb.NewPostgresAccessor(config.Postgres.ConnStr, config.Migration.OldSchema, config.Migration.ScriptsDir)
	defer dbAccessor.Destroy()
	log.Info("Start migrating data...")
	if err := dbAccessor.MigrateTo(config.Migration.Height); err != nil {
		return false, err
	}
	log.Info("Data migration has been completed")
	return true, nil
}
