package transaction

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common/hexutil"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	types2 "github.com/idena-network/idena-indexer/explorer/types"
	"github.com/idena-network/idena-indexer/log"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func Test_MemPool(t *testing.T) {
	memPool := NewMemPool(log.New())

	key1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)

	key2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)

	addr3 := tests.GetRandAddr()

	var tx1, tx2, tx3 *types.Transaction
	var apiTxs []*types2.TransactionSummary
	var apiTx *types2.TransactionDetail
	var bytes hexutil.Bytes
	var err error

	tx1 = &types.Transaction{
		Payload: []byte{0x1, 0x2},
	}
	tx1, _ = types.SignTx(tx1, key1)

	tx2 = &types.Transaction{
		Payload: []byte{0x2, 0x3},
		To:      &addr3,
	}
	tx2, _ = types.SignTx(tx2, key2)

	tx3 = &types.Transaction{
		Payload: []byte{0x2, 0x3},
		To:      &addr1,
	}
	tx3, _ = types.SignTx(tx3, key2)

	// When
	require.Nil(t, memPool.AddTransaction(tx1))

	// Then
	apiTx, err = memPool.GetTransaction(tx1.Hash().Hex())
	require.Nil(t, err)
	require.NotNil(t, apiTx)
	require.Equal(t, addr1.Hex(), apiTx.From)
	apiTx, err = memPool.GetTransaction(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, apiTx)

	bytes, err = memPool.GetTransactionRaw(tx1.Hash().Hex())
	require.Nil(t, err)
	require.Equal(t, hexutil.Bytes([]byte{0x1, 0x2}), bytes)
	bytes, err = memPool.GetTransactionRaw(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, bytes)

	apiTxs, err = memPool.GetTransactions(0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetTransactions(10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr1.Hex(), apiTxs[0].From)

	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr1.Hex(), apiTxs[0].From)
	apiTxs, err = memPool.GetAddressTransactions(addr2.Hex(), 10)
	require.Nil(t, err)
	require.Empty(t, apiTxs)

	// When
	require.Nil(t, memPool.AddTransaction(tx2))

	// Then
	apiTx, err = memPool.GetTransaction(tx1.Hash().Hex())
	require.Nil(t, err)
	require.NotNil(t, apiTx)
	require.Equal(t, addr1.Hex(), apiTx.From)
	apiTx, err = memPool.GetTransaction(tx2.Hash().Hex())
	require.Nil(t, err)
	require.NotNil(t, apiTx)
	require.Equal(t, addr2.Hex(), apiTx.From)

	bytes, err = memPool.GetTransactionRaw(tx1.Hash().Hex())
	require.Nil(t, err)
	require.Equal(t, hexutil.Bytes([]byte{0x1, 0x2}), bytes)
	bytes, err = memPool.GetTransactionRaw(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Equal(t, hexutil.Bytes([]byte{0x2, 0x3}), bytes)

	apiTxs, err = memPool.GetTransactions(0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetTransactions(10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 2)
	require.True(t, addr1.Hex() == apiTxs[0].From && addr2.Hex() == apiTxs[1].From ||
		addr1.Hex() == apiTxs[1].From && addr2.Hex() == apiTxs[0].From)

	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr1.Hex(), apiTxs[0].From)
	apiTxs, err = memPool.GetAddressTransactions(addr2.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr2.Hex(), apiTxs[0].From)
	apiTxs, err = memPool.GetAddressTransactions(addr3.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr3.Hex(), apiTxs[0].To)
	require.Equal(t, tx2.Hash().Hex(), apiTxs[0].Hash)

	// When
	require.Nil(t, memPool.AddTransaction(tx3))

	// Then
	apiTxs, err = memPool.GetTransactions(10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 3)

	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 2)

	// When
	require.Nil(t, memPool.RemoveTransaction(tx3))
	require.Nil(t, memPool.RemoveTransaction(tx2))

	// Then
	apiTx, err = memPool.GetTransaction(tx1.Hash().Hex())
	require.Nil(t, err)
	require.NotNil(t, apiTx)
	require.Equal(t, addr1.Hex(), apiTx.From)
	apiTx, err = memPool.GetTransaction(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, apiTx)

	bytes, err = memPool.GetTransactionRaw(tx1.Hash().Hex())
	require.Nil(t, err)
	require.Equal(t, hexutil.Bytes([]byte{0x1, 0x2}), bytes)
	bytes, err = memPool.GetTransactionRaw(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, bytes)

	apiTxs, err = memPool.GetTransactions(0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetTransactions(10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr1.Hex(), apiTxs[0].From)

	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 10)
	require.Nil(t, err)
	require.Len(t, apiTxs, 1)
	require.Equal(t, addr1.Hex(), apiTxs[0].From)
	apiTxs, err = memPool.GetAddressTransactions(addr2.Hex(), 10)
	require.Nil(t, err)
	require.Empty(t, apiTxs)

	// When
	require.Nil(t, memPool.RemoveTransaction(tx1))

	// Then
	apiTx, err = memPool.GetTransaction(tx1.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, apiTx)
	apiTx, err = memPool.GetTransaction(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, apiTx)

	bytes, err = memPool.GetTransactionRaw(tx1.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, bytes)
	bytes, err = memPool.GetTransactionRaw(tx2.Hash().Hex())
	require.Nil(t, err)
	require.Nil(t, bytes)

	apiTxs, err = memPool.GetTransactions(0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetTransactions(10)
	require.Nil(t, err)
	require.Empty(t, apiTxs)

	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 0)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetAddressTransactions(addr1.Hex(), 10)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
	apiTxs, err = memPool.GetAddressTransactions(addr2.Hex(), 10)
	require.Nil(t, err)
	require.Empty(t, apiTxs)
}

func Test_MemPoolConcurrency(t *testing.T) {
	memPool := NewMemPool(log.New())

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	addrHex := addr.Hex()
	to := tests.GetRandAddr()
	tx := &types.Transaction{
		To:      &to,
		Payload: []byte{0x1, 0x2},
	}
	tx, _ = types.SignTx(tx, key)
	txHashHex := tx.Hash().Hex()

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		for i := 0; i < 9999; i++ {
			require.Nil(t, memPool.AddTransaction(tx))
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 9999; i++ {
			require.Nil(t, memPool.RemoveTransaction(tx))
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 9999; i++ {
			_, err := memPool.GetTransactions(100)
			require.Nil(t, err)
			_, err = memPool.GetAddressTransactions(addrHex, 100)
			require.Nil(t, err)
			_, err = memPool.GetTransaction(txHashHex)
			require.Nil(t, err)
			_, err = memPool.GetTransactionRaw(txHashHex)
			require.Nil(t, err)
		}
		wg.Done()
	}()

	wg.Wait()
}
