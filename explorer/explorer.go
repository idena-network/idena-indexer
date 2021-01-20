package explorer

import (
	"fmt"
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/explorer/api"
	"github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/db/cached"
	"github.com/idena-network/idena-indexer/explorer/db/postgres"
	"github.com/idena-network/idena-indexer/explorer/monitoring"
	service2 "github.com/idena-network/idena-indexer/explorer/service"
	"github.com/idena-network/idena-indexer/log"
	"time"
)

type Explorer interface {
	Start()
	RouterInitializer() server.RouterInitializer
	Logger() log.Logger
	Destroy()
}

func NewExplorer(c *config.Config, memPool api.MemPool, contractsMemPool service2.ContractsMemPool, networkSizeLoader service2.NetworkSizeLoader) Explorer {
	logger, err := initLog(c.Verbosity)
	if err != nil {
		panic(err)
	}
	pm, err := createPerformanceMonitor(c.PerformanceMonitor)
	if err != nil {
		panic(err)
	}
	accessor := cached.NewCachedAccessor(
		postgres.NewPostgresAccessor(
			c.PostgresConnStr,
			c.ScriptsDir,
			networkSizeLoader,
			logger,
		),
		c.DefaultCacheMaxItemCount,
		time.Second*time.Duration(c.DefaultCacheItemLifeTimeSec),
		logger.New("component", "cachedDbAccessor"),
	)
	service := api.NewService(accessor, memPool)
	contractsService := service2.NewContracts(accessor, contractsMemPool)
	dynamicConfigHolder := config.NewDynamicConfigHolder(c.DynamicConfigFile, logger.New("component", "dConfHolder"))
	e := &explorer{
		server: api.NewServer(
			c.Port,
			c.LatestHours,
			c.ActiveAddressHours,
			c.FrozenBalanceAddrs,
			func() string {
				c := dynamicConfigHolder.GetConfig()
				if c == nil || len(c.DumpCid) == 0 {
					return ""
				}
				return fmt.Sprintf("https://ipfs.io/ipfs/%s", c.DumpCid)
			},
			service,
			contractsService,
			logger,
			pm,
		),
		db:     accessor,
		logger: logger,
	}
	return e
}

type explorer struct {
	server api.Server
	db     db.Accessor
	logger log.Logger
}

func (e *explorer) Start() {
	e.server.Start()
}

func (e *explorer) RouterInitializer() server.RouterInitializer {
	return e.server
}

func (e *explorer) Logger() log.Logger {
	return e.logger
}

func (e *explorer) Destroy() {
	e.db.Destroy()
}

func initLog(verbosity int) (log.Logger, error) {
	l := log.New()
	logLvl := log.Lvl(verbosity)
	fileHandler, err := getLogFileHandler("explorer.log")
	if err != nil {
		return nil, err
	}
	l.SetHandler(log.LvlFilterHandler(logLvl, fileHandler))
	return l, nil
}

func getLogFileHandler(fileName string) (log.Handler, error) {
	fileHandler, _ := log.FileHandler(fileName, log.TerminalFormat(false))
	return fileHandler, nil
}

func createPerformanceMonitor(c config.PerformanceMonitorConfig) (monitoring.PerformanceMonitor, error) {
	if !c.Enabled {
		return monitoring.NewEmptyPerformanceMonitor(), nil
	}
	logger, err := createPerformanceMonitorLogger()
	if err != nil {
		return monitoring.NewEmptyPerformanceMonitor(), err
	}
	interval, err := time.ParseDuration(c.Interval)
	if err != nil {
		return monitoring.NewEmptyPerformanceMonitor(), err
	}
	return monitoring.NewPerformanceMonitor(interval, logger), nil
}

func createPerformanceMonitorLogger() (log.Logger, error) {
	l := log.New()
	logLvl := log.LvlInfo
	fileHandler, err := getLogFileHandler("explorerPm.log")
	if err != nil {
		return nil, err
	}
	l.SetHandler(log.LvlFilterHandler(logLvl, fileHandler))
	return l, nil
}
