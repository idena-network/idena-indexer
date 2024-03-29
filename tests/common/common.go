package common

import (
	"database/sql"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	config2 "github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/validators"
	"github.com/idena-network/idena-go/node"
	"github.com/idena-network/idena-indexer/core/mempool"
	"github.com/idena-network/idena-indexer/core/restore"
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
	changesHistoryBlocksCount int,
	schema string,
	scriptsPathPrefix string,
) (*sql.DB, *indexer.Indexer, incoming.Listener, db.Accessor, eventbus.Bus) {
	initLog()
	pm := monitoring.NewEmptyPerformanceMonitor()
	dbConnector, dbAccessor := InitPostgres(clearDb, changesHistoryBlocksCount, schema, scriptsPathPrefix, pm)
	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))
	memDb := db2.NewMemDB()
	appState, _ := appstate.NewAppState(memDb, eventbus.New())
	appState.ValidatorsCache = validators.NewValidatorsCache(appState.IdentityState, common.Address{})
	appState.ValidatorsCache.Load()

	nodeEventBus := eventbus.New()
	collectorEventBus := eventbus.New()

	chain, _, _, _ := blockchain.NewTestBlockchain(true, nil)
	nodeCtx := &node.NodeCtx{
		PendingProofs: &sync.Map{},
		ProposerByRound: func(round uint64) (hash common.Hash, proposer []byte, ok bool) {
			return common.Hash{}, nil, true
		},
		AppState:   appState,
		Blockchain: chain.Blockchain,
	}
	nodeConfig := &config2.Config{
		Consensus: &config2.ConsensusConf{},
	}
	tokenContractHolder := new(stats.TokenContractHolderImpl)
	listener := NewTestListener(nodeEventBus, stats.NewStatsCollector(collectorEventBus, chain.Config().Consensus, tokenContractHolder), appState, nodeCtx, chain.SecStore(), nodeConfig)
	tokenContractHolder.ProvideNodeCtx(nodeCtx, nodeConfig)
	restorer := restore.NewRestorer(dbAccessor, appState, chain.Blockchain)
	upgradesVotingHolder := &TestUpgradesVotingHolder{}
	indexerEventBus := eventbus.New()
	testIndexer := indexer.NewIndexer(
		true,
		listener,
		memPoolIndexer,
		dbAccessor,
		restorer,
		nil,
		false,
		pm,
		&TestFlipLoader{},
		upgradesVotingHolder,
		20,
		5,
		indexerEventBus,
		"",
		dbAccessor,
		indexer.NewOracleVotingToProlongDetector(),
		false,
		false,
	)
	testIndexer.Start()
	return dbConnector, testIndexer, listener, dbAccessor, nodeEventBus
}

type Options struct {
	RestoreInitially                  bool
	ScriptsPathPrefix                 string
	Schema                            string
	ClearDb                           bool
	ChangesHistoryBlocksCount         int
	AppState                          *appstate.AppState
	TestBlockchain                    *blockchain.TestBlockchain
	UpgradeVotingShortHistoryItems    *int
	UpgradeVotingShortHistoryMinShift *int
	NodeConfig                        *config2.Config
	OracleVotingToProlongDetector     indexer.OracleVotingToProlongDetector
	TokenContractHolder               stats.TokenContractHolder
}

type IndexerCtx struct {
	DbConnector          *sql.DB
	Indexer              *indexer.Indexer
	Listener             incoming.Listener
	DbAccessor           db.Accessor
	EventBus             eventbus.Bus
	TestBlockchain       *blockchain.TestBlockchain
	UpgradesVotingHolder *TestUpgradesVotingHolder
}

func InitIndexer2(opt Options) *IndexerCtx {

	if opt.UpgradeVotingShortHistoryItems == nil {
		opt.UpgradeVotingShortHistoryItems = Pint(20)
	}
	if opt.UpgradeVotingShortHistoryMinShift == nil {
		opt.UpgradeVotingShortHistoryMinShift = Pint(5)
	}
	if opt.NodeConfig == nil {
		opt.NodeConfig = &config2.Config{}
	}
	if opt.NodeConfig.Consensus == nil {
		opt.NodeConfig.Consensus = &config2.ConsensusConf{}
	}
	if opt.OracleVotingToProlongDetector == nil {
		opt.OracleVotingToProlongDetector = indexer.NewOracleVotingToProlongDetector()
	}

	initLog()
	pm := monitoring.NewEmptyPerformanceMonitor()
	dbConnector, dbAccessor := InitPostgres(opt.ClearDb, opt.ChangesHistoryBlocksCount, opt.Schema, opt.ScriptsPathPrefix, pm)
	memPoolIndexer := mempool.NewIndexer(dbAccessor, log.New("component", "mpi"))
	memDb := db2.NewMemDB()
	appState := opt.AppState
	if appState == nil {
		appState, _ = appstate.NewAppState(memDb, eventbus.New())
		appState.ValidatorsCache = validators.NewValidatorsCache(appState.IdentityState, common.Address{})
		appState.ValidatorsCache.Load()
	}

	chain := opt.TestBlockchain
	if chain == nil {
		chain, _, _, _ = blockchain.NewTestBlockchain(true, nil)
	}
	nodeCtx := &node.NodeCtx{
		PendingProofs: &sync.Map{},
		ProposerByRound: func(round uint64) (hash common.Hash, proposer []byte, ok bool) {
			return common.Hash{}, nil, true
		},
		AppState:   appState,
		Blockchain: chain.Blockchain,
	}
	if opt.TokenContractHolder == nil {
		tokenContractHolder := new(stats.TokenContractHolderImpl)
		tokenContractHolder.ProvideNodeCtx(nodeCtx, chain.Config())
		opt.TokenContractHolder = tokenContractHolder
	}
	nodeEventBus := eventbus.New()
	collectorEventBus := eventbus.New()
	listener := NewTestListener(nodeEventBus, stats.NewStatsCollector(collectorEventBus, chain.Config().Consensus, opt.TokenContractHolder), appState, nodeCtx, chain.SecStore(), opt.NodeConfig)
	restorer := restore.NewRestorer(dbAccessor, appState, chain.Blockchain)
	upgradesVotingHolder := &TestUpgradesVotingHolder{}
	indexerEventBus := eventbus.New()
	testIndexer := indexer.NewIndexer(
		true,
		listener,
		memPoolIndexer,
		dbAccessor,
		restorer,
		nil,
		opt.RestoreInitially,
		pm,
		&TestFlipLoader{},
		upgradesVotingHolder,
		*opt.UpgradeVotingShortHistoryItems,
		*opt.UpgradeVotingShortHistoryMinShift,
		indexerEventBus,
		"",
		dbAccessor,
		opt.OracleVotingToProlongDetector,
		false,
		false,
	)
	testIndexer.Start()
	return &IndexerCtx{
		dbConnector, testIndexer, listener, dbAccessor, nodeEventBus, chain, upgradesVotingHolder,
	}
}

func InitPostgres(
	clearDb bool,
	changesHistoryBlocksCount int,
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
		changesHistoryBlocksCount,
		false,
		"",
		"",
	)
	return dbConnector, dbAccessor
}

func InitDefaultPostgres(scriptsPathPrefix string) (*sql.DB, db.Accessor) {
	return InitPostgres(true, 0, PostgresSchema, scriptsPathPrefix,
		monitoring.NewEmptyPerformanceMonitor())
}

func Pint(v int) *int {
	return &v
}
