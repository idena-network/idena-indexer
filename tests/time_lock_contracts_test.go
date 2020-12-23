package tests

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/incoming"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func Test_TimeLockContractDeploy(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	// When
	timestamp := time.Now()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress1, contractAddress2)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 1, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	tlContracts, err := testCommon.GetTimeLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(tlContracts))
	require.Equal(t, 1, tlContracts[0].TxId)
	require.Equal(t, timestamp.UTC().Unix(), tlContracts[0].Timestamp)

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(txReceipts))

	require.True(t, txReceipts[0].Success)
	require.Equal(t, 1, txReceipts[0].TxId)
	require.Equal(t, "0.000000000000001100", txReceipts[0].GasCost)
	require.Equal(t, 11, txReceipts[0].GasUsed)
	require.Empty(t, txReceipts[0].Error)

	require.False(t, txReceipts[1].Success)
	require.Equal(t, 2, txReceipts[1].TxId)
	require.Equal(t, "0.000000000000000000", txReceipts[1].GasCost)
	require.Equal(t, 0, txReceipts[1].GasUsed)
	require.Equal(t, "error message", txReceipts[1].Error)
}

func deployTimeLockContracts(t *testing.T, listener incoming.Listener, bus eventbus.Bus, timestamp time.Time, addr1, addr2 common.Address) {
	appState := listener.NodeCtx().AppState

	var height uint64
	height++
	appState.Precommit()
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)

	// Success
	tx := &types.Transaction{AccountNonce: 1}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockDeploy(addr1, uint64(timestamp.Unix()))
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100), ContractAddress: addr1}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Failed tx receipt
	tx = &types.Transaction{AccountNonce: 2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockDeploy(addr2, uint64(timestamp.Unix()))
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), Error: errors.New("error message")}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
}

func Test_TimeLockContractCallTransferToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(destAddr, big.NewInt(1))

	// Deploy contract
	timestamp := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress, tests.GetRandAddr())

	// Transfer
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockCallTransfer(destAddr, big.NewInt(54321000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	transfers, err := testCommon.GetTimeLockContractCallTransfers(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(transfers))
	require.Equal(t, 1, transfers[0].ContractTxId)
	require.Equal(t, 3, transfers[0].TxId)
	require.Equal(t, destAddr.Hex(), transfers[0].Dest)
	require.Equal(t, "0.000000000054321", transfers[0].Amount.String())
}

func Test_TimeLockContractCallTransferToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()

	// Deploy contract
	timestamp := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress, tests.GetRandAddr())

	// Transfer
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockCallTransfer(destAddr, big.NewInt(54321000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	transfers, err := testCommon.GetTimeLockContractCallTransfers(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(transfers))
	require.Equal(t, 1, transfers[0].ContractTxId)
	require.Equal(t, 3, transfers[0].TxId)
	require.Equal(t, destAddr.Hex(), transfers[0].Dest)
	require.Equal(t, "0.000000000054321", transfers[0].Amount.String())
}

func Test_TimeLockContractTerminationToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(destAddr, big.NewInt(1))

	// Deploy contract
	timestamp := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress, tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetTimeLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_TimeLockContractTerminationToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()

	// Deploy contract
	timestamp := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress, tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetTimeLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_TimeLockContractReset(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	timestamp := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployTimeLockContracts(t, listener, bus, timestamp, contractAddress, tests.GetRandAddr())

	// Call transfer
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockCallTransfer(tests.GetRandAddr(), big.NewInt(54321000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Terminate
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddTimeLockTermination(tests.GetRandAddr())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	assertLens := func(receiptsLen, contractsLen, timelockContractsLen, callTransfersLen, terminationsLen, addressesLen int) {
		txReceipts, err := testCommon.GetTxReceipts(db)
		require.Nil(t, err)
		require.Equal(t, receiptsLen, len(txReceipts))

		contracts, err := testCommon.GetContracts(db)
		require.Nil(t, err)
		require.Equal(t, contractsLen, len(contracts))

		tlContracts, err := testCommon.GetTimeLockContracts(db)
		require.Nil(t, err)
		require.Equal(t, timelockContractsLen, len(tlContracts))

		callTransfers, err := testCommon.GetTimeLockContractCallTransfers(db)
		require.Nil(t, err)
		require.Equal(t, callTransfersLen, len(callTransfers))

		terminations, err := testCommon.GetTimeLockContractTerminations(db)
		require.Nil(t, err)
		require.Equal(t, terminationsLen, len(terminations))

		addresses, err := testCommon.GetAddresses(db)
		require.Nil(t, err)
		require.Equal(t, addressesLen, len(addresses))
	}

	assertLens(4, 1, 1, 1, 1, 6)

	require.Nil(t, dbAccessor.ResetTo(4))
	assertLens(4, 1, 1, 1, 1, 6)

	require.Nil(t, dbAccessor.ResetTo(3))
	assertLens(3, 1, 1, 1, 0, 4)

	require.Nil(t, dbAccessor.ResetTo(2))
	assertLens(2, 1, 1, 0, 0, 2)

	require.Nil(t, dbAccessor.ResetTo(1))
	assertLens(0, 0, 0, 0, 0, 1)

	require.Nil(t, dbAccessor.ResetTo(0))
	assertLens(0, 0, 0, 0, 0, 0)
}
