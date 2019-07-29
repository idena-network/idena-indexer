package main

import (
	config3 "github.com/idena-network/idena-go/config"
	nodeLog "github.com/idena-network/idena-go/log"
	"github.com/idena-network/idena-indexer/config"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/explorer"
	config2 "github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/idena-network/idena-indexer/log"
	"gopkg.in/urfave/cli.v1"
	"os"
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
			Value: "explorerConfig.json",
		},
		cli.StringFlag{
			Name:  "indexerConfig",
			Usage: "Indexer config file",
			Value: "indexerConfig.json",
		},
		config3.TcpPortFlag,
		config3.RpcPortFlag,
		config3.IpfsPortFlag,
	}

	app.Action = func(context *cli.Context) error {

		e := explorer.NewExplorer(config2.LoadConfig(context.String("explorerConfig")))
		defer e.Destroy()
		go e.Start()

		conf := config.LoadConfig(context.String("indexerConfig"))
		initLog(conf.Verbosity, conf.NodeVerbosity)
		indexer := initIndexer(context, conf)
		defer indexer.Destroy()
		indexer.Start()
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

func initIndexer(context *cli.Context, config *config.Config) *indexer.Indexer {
	listener := incoming.NewListener(context)
	dbAccessor := db.NewPostgresAccessor(config.PostgresConnStr, config.ScriptsDir)
	return indexer.NewIndexer(listener, dbAccessor)
}
