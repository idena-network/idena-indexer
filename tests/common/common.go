package common

import (
	"database/sql"
	"github.com/idena-network/idena-go/blockchain"
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

var defaultScriptsPath = filepath.Join("resources", "scripts", "indexer")

func initLog() {
	handler := log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stdout, log.TerminalFormat(false)))
	log.Root().SetHandler(handler)
}

func InitIndexer(
	clearDb bool,
	committeeRewardBlocksCount int,
	schema string,
	scriptsPathPrefix string,
) (*sql.DB, *indexer.Indexer, incoming.Listener, db.Accessor, eventbus.Bus) {
	initLog()
	pm := monitoring.NewEmptyPerformanceMonitor()
	dbConnector, dbAccessor := InitPostgres(clearDb, committeeRewardBlocksCount, schema, scriptsPathPrefix, pm)
	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))
	memDb := db2.NewMemDB()
	appState := appstate.NewAppState(memDb, eventbus.New())
	bus := eventbus.New()
	//chain := blockchain.NewBlockchain(&config.Config{
	//	GenesisConf: &config.GenesisConf{},
	//}, memDb, nil, appState, nil, nil, bus, nil, nil)
	//chain.GenerateGenesis(111)

	chain, _, _, _ := blockchain.NewTestBlockchain(true, nil)
	nodeCtx := &node.NodeCtx{
		ProofsByRound: &sync.Map{},
		PendingProofs: &sync.Map{},
		AppState:      appState,
		Blockchain:    chain.Blockchain,
	}
	listener := NewTestListener(bus, stats.NewStatsCollector(), appState, nodeCtx)
	testIndexer := indexer.NewIndexer(
		listener,
		memPoolIndexer,
		dbAccessor,
		nil,
		nil,
		false,
		pm,
		&TestFlipLoader{},
	)
	testIndexer.Start()
	return dbConnector, testIndexer, listener, dbAccessor, bus
}

func InitPostgres(
	clearDb bool,
	committeeRewardBlocksCount int,
	schema string,
	scriptsPathPrefix string,
	pm monitoring.PerformanceMonitor) (*sql.DB, db.Accessor) {
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
	dbAccessor := db.NewPostgresAccessor(
		PostgresConnStr+"&search_path="+schema,
		filepath.Join(scriptsPathPrefix, defaultScriptsPath),
		&TestWordsLoader{},
		pm,
		committeeRewardBlocksCount,
	)
	return dbConnector, dbAccessor
}

func InitDefaultPostgres(scriptsPathPrefix string) (*sql.DB, db.Accessor) {
	return InitPostgres(true, 0, PostgresSchema, scriptsPathPrefix,
		monitoring.NewEmptyPerformanceMonitor())
}
