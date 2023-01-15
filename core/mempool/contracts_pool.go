package mempool

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-go/vm"
	"github.com/idena-network/idena-indexer/core/stats"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

type Contracts interface {
	GetOracleVotingContractDeploys(author common.Address) ([]db.OracleVotingContract, error)
	GetAddressContractTxs(address, contractAddress string) ([]db.Transaction, error)
	ProcessTx(tx *types.Transaction) error
	RemoveTx(tx *types.Transaction)
}

func NewContracts(appState *appstate.AppState, chain *blockchain.Blockchain, nodeConfig *config.Config, logger log.Logger) Contracts {
	c := &contractsImpl{
		appState:            appState,
		chain:               chain,
		nodeConfig:          nodeConfig,
		logger:              logger,
		statsCollector:      stats.NewStatsCollector(eventbus.New(), nodeConfig.Consensus),
		txChan:              make(chan *types.Transaction, 10000),
		oracleVotingDeploys: make(map[common.Address]map[common.Hash]*db.OracleVotingContract),
		addressContractTxs:  newAddressContractTxs(),
	}
	c.statsCollector.(stats.StatsHolder).Enable()
	go c.startListening()
	go c.startSizeLogging()
	return c
}

type contractsImpl struct {
	appState   *appstate.AppState
	chain      *blockchain.Blockchain
	nodeConfig *config.Config
	logger     log.Logger

	statsCollector collector.StatsCollector
	appStateCache  *appStateCache

	txChan chan *types.Transaction

	oracleVotingDeploysMutex sync.RWMutex
	oracleVotingDeploys      map[common.Address]map[common.Hash]*db.OracleVotingContract // author -> txHash -> Contract

	addressContractTxs *addressContractTxs
}

type appStateCache struct {
	height   uint64
	appState *appstate.AppState
}

type addressContractTxs struct {
	mutex                     sync.RWMutex
	txsByAddressAndContract   map[string]map[string]*sync.Map // sender -> contract address -> tx hash -> tx
	contractAddressesByTxHash map[string]string
}

func newAddressContractTxs() *addressContractTxs {
	return &addressContractTxs{
		txsByAddressAndContract:   make(map[string]map[string]*sync.Map),
		contractAddressesByTxHash: make(map[string]string),
	}
}

func (t *addressContractTxs) add(tx *db.Transaction, contractAddress string) {
	txHashKey := strings.ToLower(tx.Hash)
	txAddressKey := strings.ToLower(tx.From)
	contractAddressKey := strings.ToLower(contractAddress)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.contractAddressesByTxHash[txHashKey] = contractAddress
	txsByContract, ok := t.txsByAddressAndContract[txAddressKey]
	if !ok {
		txsByContract = make(map[string]*sync.Map)
		t.txsByAddressAndContract[txAddressKey] = txsByContract
	}
	txs, ok := txsByContract[contractAddressKey]
	if !ok {
		txs = &sync.Map{}
		txsByContract[contractAddressKey] = txs
	}
	txs.Store(tx.Hash, tx)
}

func (t *addressContractTxs) remove(txHash, sender string) {

	txHashKey := strings.ToLower(txHash)
	txAddressKey := strings.ToLower(sender)

	t.mutex.Lock()
	defer t.mutex.Unlock()
	contractAddress, ok := t.contractAddressesByTxHash[txHashKey]
	if !ok {
		return
	}
	contractAddressKey := strings.ToLower(contractAddress)
	delete(t.contractAddressesByTxHash, txHashKey)
	txsByContract, ok := t.txsByAddressAndContract[txAddressKey]
	if !ok {
		return
	}
	txs, ok := txsByContract[contractAddressKey]
	if !ok {
		return
	}
	txs.Delete(txHashKey)
	isEmpty := true
	txs.Range(func(key, value interface{}) bool {
		isEmpty = false
		return false
	})
	if isEmpty {
		delete(txsByContract, contractAddressKey)
	}
	if len(txsByContract) == 0 {
		delete(t.txsByAddressAndContract, txAddressKey)
	}
}

func (t *addressContractTxs) get(address, contractAddress string) []db.Transaction {
	txAddressKey := strings.ToLower(address)
	contractAddressKey := strings.ToLower(contractAddress)
	t.mutex.RLock()
	txsByContract, ok := t.txsByAddressAndContract[txAddressKey]
	if !ok {
		t.mutex.RUnlock()
		return nil
	}
	txs, ok := txsByContract[contractAddressKey]
	if !ok {
		t.mutex.RUnlock()
		return nil
	}
	t.mutex.RUnlock()
	var res []db.Transaction
	txs.Range(func(key, value interface{}) bool {
		res = append(res, *(value.(*db.Transaction)))
		return true
	})
	return res
}

func (c *contractsImpl) startSizeLogging() {
	for {
		time.Sleep(time.Minute * 5)
		var deploys, txs int
		c.oracleVotingDeploysMutex.RLock()
		deploys = len(c.oracleVotingDeploys)
		c.oracleVotingDeploysMutex.RUnlock()

		c.addressContractTxs.mutex.RLock()
		txs = len(c.addressContractTxs.contractAddressesByTxHash)
		addresses := len(c.addressContractTxs.txsByAddressAndContract)
		c.addressContractTxs.mutex.RUnlock()

		c.logger.Debug(fmt.Sprintf("txs: %v, addresses: %v, deploys: %v", txs, addresses, deploys))
	}
}

func (c *contractsImpl) GetOracleVotingContractDeploys(author common.Address) ([]db.OracleVotingContract, error) {
	c.oracleVotingDeploysMutex.RLock()
	defer c.oracleVotingDeploysMutex.RUnlock()
	deploys, ok := c.oracleVotingDeploys[author]
	if !ok || len(deploys) == 0 {
		return nil, nil
	}
	res := make([]db.OracleVotingContract, len(deploys))
	i := 0
	for _, deploy := range deploys {
		res[i] = *deploy
		i++
	}
	return res, nil
}

func (c *contractsImpl) GetAddressContractTxs(address, contractAddress string) ([]db.Transaction, error) {
	return c.addressContractTxs.get(address, contractAddress), nil
}

func (c *contractsImpl) ProcessTx(tx *types.Transaction) error {
	select {
	case c.txChan <- tx:
	default:
		return errors.New("tx chan size limit reached")
	}
	return nil
}

func (c *contractsImpl) startListening() {
	for {
		tx := <-c.txChan
		if err := c.processTx(tx); err != nil {
			c.logger.Warn(errors.Wrapf(err, "Unable to process tx %v", tx.Hash().Hex()).Error())
		}
	}
}

func (c *contractsImpl) processTx(tx *types.Transaction) error {
	if isContractTx(tx) {
		return c.processContractTx(tx)
	}
	if tx.To == nil {
		return nil
	}
	appState, err := c.readonlyAppState()
	if err != nil {
		return errors.Wrap(err, "unable to get readonly app state to process tx")
	}
	to := *tx.To
	if !isContractAddress(to, appState) {
		return nil
	}
	c.applyTxSentToContract(tx, to)
	return nil
}

func isContractTx(tx *types.Transaction) bool {
	return tx.Type == types.DeployContractTx || tx.Type == types.CallContractTx || tx.Type == types.TerminateContractTx
}

func isContractAddress(address common.Address, appState *appstate.AppState) bool {
	return appState.State.GetCodeHash(address) != nil
}

func (c *contractsImpl) processContractTx(tx *types.Transaction) error {
	appState, err := c.readonlyAppState()
	if err != nil {
		return errors.Wrap(err, "unable to get readonly app state to process tx")
	}
	statsCollector := c.statsCollector
	statsCollector.EnableCollecting()
	defer statsCollector.CompleteCollecting()
	statsCollector.BeginApplyingTx(tx, appState)
	defer statsCollector.CompleteApplyingTx(appState)
	cvm := vm.NewVmImpl(appState, c.chain, c.chain.Head, statsCollector, c.nodeConfig)
	txReceipt := cvm.Run(tx, nil, -1)
	c.applyDeployTx(tx.Hash(), txReceipt, appState)
	c.applyContractTx(tx, txReceipt)
	return nil
}

func (c *contractsImpl) applyDeployTx(txHash common.Hash, txReceipt *types.TxReceipt, appState *appstate.AppState) {
	statsCollector := c.statsCollector
	statsCollector.AddTxReceipt(txReceipt, appState)

	statsHolder := statsCollector.(stats.StatsHolder)
	collectedStats := statsHolder.GetStats()
	if collectedStats == nil {
		return
	}

	if len(collectedStats.OracleVotingContracts) == 0 {
		return
	}

	statsDeploy := collectedStats.OracleVotingContracts[0]
	author := txReceipt.From

	c.oracleVotingDeploysMutex.Lock()
	c.getOrCreateAuthorContracts(author)[txHash] = statsDeploy
	c.oracleVotingDeploysMutex.Unlock()
}

func (c *contractsImpl) applyContractTx(tx *types.Transaction, txReceipt *types.TxReceipt) {
	if !txReceipt.Success {
		c.logger.Warn(fmt.Sprintf("contract tx receipt is not success: %v", txReceipt.Error))
		return
	}
	c.applyTxSentToContract(tx, txReceipt.ContractAddress)
}

func (c *contractsImpl) applyTxSentToContract(tx *types.Transaction, contractAddress common.Address) {
	sender, _ := types.Sender(tx)
	var to string
	if tx.To != nil {
		to = tx.To.Hex()
	}
	c.addressContractTxs.add(&db.Transaction{
		Hash:   tx.Hash().Hex(),
		Type:   tx.Type,
		From:   sender.Hex(),
		To:     to,
		Amount: blockchain.ConvertToFloat(tx.Amount),
		Tips:   blockchain.ConvertToFloat(tx.Tips),
		MaxFee: blockchain.ConvertToFloat(tx.MaxFee),
		Nonce:  tx.AccountNonce,
	}, contractAddress.Hex())
}

func (c *contractsImpl) RemoveTx(tx *types.Transaction) {
	c.removeDeployTx(tx)
	c.removeAddressContractTx(tx)
}

func (c *contractsImpl) removeDeployTx(tx *types.Transaction) {
	if tx.Type != types.DeployContractTx {
		return
	}
	sender, _ := types.Sender(tx)
	c.oracleVotingDeploysMutex.Lock()
	defer c.oracleVotingDeploysMutex.Unlock()
	if txs, ok := c.oracleVotingDeploys[sender]; ok {
		delete(txs, tx.Hash())
		if len(txs) == 0 {
			delete(c.oracleVotingDeploys, sender)
		}
	}
}

func (c *contractsImpl) removeAddressContractTx(tx *types.Transaction) {
	if !isContractTx(tx) && tx.To == nil {
		return
	}
	sender, _ := types.Sender(tx)
	c.addressContractTxs.remove(tx.Hash().Hex(), sender.Hex())
}

func (c *contractsImpl) readonlyAppState() (*appstate.AppState, error) {
	currentHeight := c.chain.Head.Height()
	if c.appStateCache != nil && c.appStateCache.height == currentHeight {
		return c.appStateCache.appState, nil
	}
	s, err := c.appState.Readonly(currentHeight)
	if err != nil {
		return nil, err
	}
	c.appStateCache = &appStateCache{
		height:   currentHeight,
		appState: s,
	}
	return s, nil
}

func (c *contractsImpl) getOrCreateAuthorContracts(address common.Address) map[common.Hash]*db.OracleVotingContract {
	res := c.oracleVotingDeploys[address]
	if res == nil {
		res = make(map[common.Hash]*db.OracleVotingContract)
		c.oracleVotingDeploys[address] = res
	}
	return res
}
