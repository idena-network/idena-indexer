package transaction

import (
	"fmt"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-indexer/core/conversion"
	types2 "github.com/idena-network/idena-indexer/core/types"
	"github.com/idena-network/idena-indexer/log"
	"sync"
	"time"
)

type MemPool interface {
	GetTransaction(hash string) (*types2.TransactionDetail, error)
	GetTransactionRaw(hash string) (hexutil.Bytes, error)
	GetAddressTransactions(address string, count int) ([]*types2.TransactionSummary, error)
	GetTransactions(count int) ([]*types2.TransactionSummary, error)
	GetTransactionsCount() (int, error)

	AddTransaction(tx *types.Transaction) error
	RemoveTransaction(tx *types.Transaction) error
}

func NewMemPool(logger log.Logger) MemPool {
	res := &memPool{
		txsByHash:    newTxsByHashWrapper(),
		txsByAddress: newTxsByAddressWrapper(),
		logger:       logger,
	}
	go res.loopLog()
	return res
}

type memPool struct {
	logger       log.Logger
	txsByHash    *txsByHashWrapper
	txsByAddress *txsByAddressWrapper
	mutex        sync.Mutex
}

func (pool *memPool) loopLog() {
	for {
		time.Sleep(time.Minute * 5)
		txs := pool.txsByHash.len()
		addresses := pool.txsByAddress.len()
		pool.logger.Info(fmt.Sprintf("txs: %v, addresses: %v", txs, addresses))
	}
}

type txsByHashWrapper struct {
	mutex     sync.RWMutex
	txsByHash map[common.Hash]*types.Transaction
}

func newTxsByHashWrapper() *txsByHashWrapper {
	return &txsByHashWrapper{
		txsByHash: make(map[common.Hash]*types.Transaction),
	}
}

func (w *txsByHashWrapper) isEmpty() bool {
	return w.len() == 0
}

func (w *txsByHashWrapper) len() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return len(w.txsByHash)
}

func (w *txsByHashWrapper) add(tx *types.Transaction) {
	hash := tx.Hash()
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.txsByHash[hash] = tx
}

func (w *txsByHashWrapper) contains(hash common.Hash) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	_, ok := w.txsByHash[hash]
	return ok
}

func (w *txsByHashWrapper) remove(hash common.Hash) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	delete(w.txsByHash, hash)
}

func (w *txsByHashWrapper) get(hash common.Hash) *types.Transaction {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.txsByHash[hash]
}

func (w *txsByHashWrapper) all(count int) []*types.Transaction {
	if count <= 0 {
		return nil
	}
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	size := int(math.Min(uint64(count), uint64(len(w.txsByHash))))
	if size == 0 {
		return nil
	}
	res := make([]*types.Transaction, size)
	i := 0
	for _, tx := range w.txsByHash {
		res[i] = tx
		i++
		if i == size {
			break
		}
	}
	return res
}

type txsByAddressWrapper struct {
	mutex        sync.RWMutex
	txsByAddress map[common.Address]*txsByHashWrapper
}

func newTxsByAddressWrapper() *txsByAddressWrapper {
	return &txsByAddressWrapper{
		txsByAddress: make(map[common.Address]*txsByHashWrapper),
	}
}

func (w *txsByAddressWrapper) len() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return len(w.txsByAddress)
}

func (w *txsByAddressWrapper) add(tx *types.Transaction) {
	sender, _ := types.Sender(tx)
	w.addAddressTx(sender, tx)
	if tx.To != nil {
		w.addAddressTx(*tx.To, tx)
	}
}

func (w *txsByAddressWrapper) addAddressTx(address common.Address, tx *types.Transaction) {
	w.mutex.RLock()
	addressTxs, ok := w.txsByAddress[address]
	if !ok {
		addressTxs = newTxsByHashWrapper()
	}
	w.mutex.RUnlock()
	addressTxs.add(tx)
	if !ok {
		w.mutex.Lock()
		w.txsByAddress[address] = addressTxs
		w.mutex.Unlock()
	}
}

func (w *txsByAddressWrapper) remove(tx *types.Transaction) {
	sender, _ := types.Sender(tx)
	hash := tx.Hash()
	w.removeByAddressAndHash(sender, hash)
	if tx.To != nil {
		w.removeByAddressAndHash(*tx.To, hash)
	}
}

func (w *txsByAddressWrapper) removeByAddressAndHash(address common.Address, hash common.Hash) {
	w.mutex.RLock()
	txsByHashWrapper, ok := w.txsByAddress[address]
	w.mutex.RUnlock()
	if !ok {
		return
	}
	txsByHashWrapper.remove(hash)
	if txsByHashWrapper.isEmpty() {
		w.mutex.Lock()
		delete(w.txsByAddress, address)
		w.mutex.Unlock()
	}
}

func (w *txsByAddressWrapper) get(address common.Address, count int) []*types.Transaction {
	w.mutex.RLock()
	txsByHash, ok := w.txsByAddress[address]
	w.mutex.RUnlock()
	if !ok {
		return nil
	}
	return txsByHash.all(count)
}

func (pool *memPool) AddTransaction(tx *types.Transaction) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.addTx(tx)
	return nil
}

func (pool *memPool) RemoveTransaction(tx *types.Transaction) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.removeTx(tx)
	return nil
}

func (pool *memPool) GetTransaction(hash string) (*types2.TransactionDetail, error) {
	txHash := common.HexToHash(hash)
	tx := pool.txsByHash.get(txHash)
	if tx == nil {
		return nil, nil
	}
	return toTransactionDetail(tx), nil
}

func (pool *memPool) GetTransactionRaw(hash string) (hexutil.Bytes, error) {
	txHash := common.HexToHash(hash)
	tx := pool.txsByHash.get(txHash)
	if tx == nil {
		return nil, nil
	}
	return tx.Payload, nil
}

func (pool *memPool) GetAddressTransactions(address string, count int) ([]*types2.TransactionSummary, error) {
	txAddress := common.HexToAddress(address)
	txs := pool.txsByAddress.get(txAddress, count)
	return toTransactionSummaries(txs), nil
}

func (pool *memPool) GetTransactions(count int) ([]*types2.TransactionSummary, error) {
	txs := pool.txsByHash.all(count)
	return toTransactionSummaries(txs), nil
}

func (pool *memPool) GetTransactionsCount() (int, error) {
	return pool.txsByHash.len(), nil
}

func (pool *memPool) addTx(tx *types.Transaction) {
	pool.txsByHash.add(tx)
	pool.txsByAddress.add(tx)
}

func (pool *memPool) removeTx(tx *types.Transaction) {
	if !pool.txsByHash.contains(tx.Hash()) {
		return
	}
	pool.txsByHash.remove(tx.Hash())
	pool.txsByAddress.remove(tx)
}

func toTransactionDetail(tx *types.Transaction) *types2.TransactionDetail {
	var from, to string
	sender, _ := types.Sender(tx)
	from = conversion.ConvertAddress(sender)
	if tx.To != nil {
		to = conversion.ConvertAddress(*tx.To)
	}
	return &types2.TransactionDetail{
		Epoch:  uint64(tx.Epoch),
		Hash:   conversion.ConvertHash(tx.Hash()),
		Type:   conversion.ConvertTxType(tx.Type),
		From:   from,
		To:     to,
		Amount: blockchain.ConvertToFloat(tx.Amount),
		Tips:   blockchain.ConvertToFloat(tx.Tips),
		MaxFee: blockchain.ConvertToFloat(tx.MaxFee),
		Size:   uint32(tx.Size()),
	}
}

func toTransactionSummary(tx *types.Transaction) *types2.TransactionSummary {
	var from, to string
	sender, _ := types.Sender(tx)
	from = conversion.ConvertAddress(sender)
	if tx.To != nil {
		to = conversion.ConvertAddress(*tx.To)
	}
	amount := blockchain.ConvertToFloat(tx.Amount)
	tips := blockchain.ConvertToFloat(tx.Tips)
	maxFee := blockchain.ConvertToFloat(tx.MaxFee)
	return &types2.TransactionSummary{
		Hash:   conversion.ConvertHash(tx.Hash()),
		Type:   conversion.ConvertTxType(tx.Type),
		From:   from,
		To:     to,
		Amount: &amount,
		Tips:   &tips,
		MaxFee: &maxFee,
		Size:   uint32(tx.Size()),
	}
}

func toTransactionSummaries(txs []*types.Transaction) []*types2.TransactionSummary {
	if len(txs) == 0 {
		return nil
	}
	res := make([]*types2.TransactionSummary, 0, len(txs))
	for _, tx := range txs {
		res = append(res, toTransactionSummary(tx))
	}
	return res
}
