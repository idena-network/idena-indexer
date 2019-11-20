package explorer

import (
	"github.com/idena-network/idena-indexer/core/server"
	"github.com/idena-network/idena-indexer/explorer/api"
	"github.com/idena-network/idena-indexer/explorer/config"
	"github.com/idena-network/idena-indexer/explorer/db"
	"github.com/idena-network/idena-indexer/explorer/db/postgres"
	"github.com/idena-network/idena-indexer/log"
	"os"
	"runtime"
)

type Explorer interface {
	Start()
	RouterInitializer() server.RouterInitializer
	Logger() log.Logger
	Destroy()
}

func NewExplorer(c *config.Config) Explorer {
	logger := initLog(c.Verbosity)
	accessor := postgres.NewPostgresAccessor(c.PostgresConnStr, c.ScriptsDir, logger)
	e := &explorer{
		server: api.NewServer(c.Port, c.LatestHours, accessor, logger),
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

func initLog(verbosity int) log.Logger {
	l := log.New()
	logLvl := log.Lvl(verbosity)
	if runtime.GOOS == "windows" {
		l.SetHandler(log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stdout, log.LogfmtFormat())))
	} else {
		l.SetHandler(log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	}
	return l
}
