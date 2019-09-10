package main

import (
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/core/activity"
	"github.com/idena-network/idena-indexer/core/api"
	"github.com/idena-network/idena-indexer/core/penalty"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/explorer"
	explorerConfig "github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/migration/flip"
	migration "github.com/idena-network/idena-indexer/migration/flip/db"
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

		// Explorer
		explorerConf := explorerConfig.LoadConfig(context.String("explorerConfig"))
		e := explorer.NewExplorer(explorerConf)
		defer e.Destroy()

		// Indexer
		conf := config.LoadConfig(context.String("indexerConfig"))
		initLog(conf.Verbosity, conf.NodeVerbosity)
		indxr := initIndexer(conf)
		defer indxr.Destroy()
		indxr.Start()

		// Server for explorer & indexer api
		lastActivities := activity.NewLastActivitiesCache(indxr.OfflineDetector())
		currentPenalties := penalty.NewCurrentPenaltiesCache(indxr.AppState(), indxr.Blockchain())
		explorerRi := e.RouterInitializer()
		indexerApi := api.NewApi(lastActivities, currentPenalties)
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

func initIndexer(config *config.Config) *indexer.Indexer {
	listener := incoming.NewListener(config.NodeConfigFile)
	dbAccessor := db.NewPostgresAccessor(config.Postgres.ConnStr, config.Postgres.ScriptsDir)
	var sfs *flip.SecondaryFlipStorage
	if config.MigrationPostgres != nil {
		sfs = flip.NewSecondaryFlipStorage(migration.NewPostgresAccessor(config.MigrationPostgres.ConnStr, config.MigrationPostgres.ScriptsDir))
	}
	return indexer.NewIndexer(listener, dbAccessor, sfs, uint64(config.GenesisBlockHeight))
}
