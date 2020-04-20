package tests

import (
	"database/sql"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/incoming"
	"github.com/idena-network/idena-indexer/indexer"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/monitoring"
	db2 "github.com/tendermint/tm-db"
	"os"
	"path/filepath"
	"sync"
)

const (
	PostgresConnStr = "postgres://postgres@localhost?sslmode=disable"
	PostgresSchema  = "auto_test_schema"
)

func initLog() {
	handler := log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stdout, log.TerminalFormat(false)))
	log.Root().SetHandler(handler)
}

func InitIndexer(
	clearDb bool,
	committeeRewardBlocksCount int,
	schema string,
) (*sql.DB, *indexer.Indexer, incoming.Listener, db.Accessor, eventbus.Bus) {
	initLog()
	dbConnector, err := sql.Open("postgres", PostgresConnStr)
	if err != nil {
		panic(err)
	}
	if clearDb {
		_, err = dbConnector.Exec("DROP SCHEMA IF EXISTS " + schema + " CASCADE")
		if err != nil {
			panic(err)
		}
		_, err = dbConnector.Exec("CREATE SCHEMA " + schema)
		if err != nil {
			panic(err)
		}
	}
	_, err = dbConnector.Exec("SET SEARCH_PATH TO " + schema)
	if err != nil {
		panic(err)
	}
	pm := monitoring.NewEmptyPerformanceMonitor()
	dbAccessor := db.NewPostgresAccessor(
		PostgresConnStr+"&search_path="+schema,
		filepath.Join("..", "resources", "scripts", "indexer"),
		&TestWordsLoader{},
		pm,
		committeeRewardBlocksCount,
	)
	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))
	appState := appstate.NewAppState(db2.NewMemDB(), eventbus.New())
	bus := eventbus.New()
	nodeCtx := &node.NodeCtx{
		ProofsByRound: &sync.Map{},
		PendingProofs: &sync.Map{},
		AppState:      appState,
	}
	listener := NewTestListener(bus, stats.NewStatsCollector(), appState, nodeCtx)
	testIndexer := indexer.NewIndexer(
		listener,
		memPoolIndexer,
		dbAccessor,
		nil,
		nil,
		1,
		false,
		pm,
		nil,
	)
	testIndexer.Start()
	return dbConnector, testIndexer, listener, dbAccessor, bus
}
