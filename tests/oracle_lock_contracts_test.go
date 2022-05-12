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
)

func Test_OracleLockContractDeployWithExistingAddresses(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	// Add dest address to state
	appState := listener.NodeCtx().AppState
	appState.State.SetBalance(oracleVotingAddress, big.NewInt(1))
	appState.State.SetBalance(successAddress, big.NewInt(1))
	appState.State.SetBalance(failAddress, big.NewInt(1))
	deployOracleLockContracts(t, listener, bus, contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 3, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	olContracts, err := testCommon.GetOracleLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(olContracts))
	require.Equal(t, 1, olContracts[0].TxId)
	require.Equal(t, oracleVotingAddress.Hex(), olContracts[0].OracleVotingAddress)
	require.Equal(t, 2, olContracts[0].Value)
	require.Equal(t, successAddress.Hex(), olContracts[0].SuccessAddress)
	require.Equal(t, failAddress.Hex(), olContracts[0].FailAddress)

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

func Test_OracleLockContractDeployWithNotExistingAddresses(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 3, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	olContracts, err := testCommon.GetOracleLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(olContracts))
	require.Equal(t, 1, olContracts[0].TxId)
	require.Equal(t, oracleVotingAddress.Hex(), olContracts[0].OracleVotingAddress)
	require.Equal(t, 2, olContracts[0].Value)
	require.Equal(t, successAddress.Hex(), olContracts[0].SuccessAddress)
	require.Equal(t, failAddress.Hex(), olContracts[0].FailAddress)

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

func deployOracleLockContracts(t *testing.T, listener incoming.Listener, bus eventbus.Bus, addr1, addr2, oracleVotingAddress, successAddress, failAddress common.Address) {
	appState := listener.NodeCtx().AppState

	var height uint64
	height++
	appState.Precommit(true)
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)

	// Success
	tx := &types.Transaction{AccountNonce: 1}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockDeploy(addr1, oracleVotingAddress, 2, successAddress, failAddress)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100), ContractAddress: addr1}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Failed tx receipt
	tx = &types.Transaction{AccountNonce: 2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockDeploy(addr2, oracleVotingAddress, 3, successAddress, failAddress)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), Error: errors.New("error message")}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
}

func Test_OracleLockContractCallCheckOracleVoting(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr())

	// Call success
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallCheckOracleVoting(3, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Error
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallCheckOracleVoting(0, errors.New("error"))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetOracleLockContractCallCheckOracleVotings(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, 3, *calls[0].OracleVotingResult)

	require.Equal(t, 1, calls[1].ContractTxId)
	require.Equal(t, 4, calls[1].TxId)
	require.Nil(t, calls[1].OracleVotingResult)
}

func Test_OracleLockContractCallPush(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr())

	// Call success
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallPush(true, 3 /*nil,*/, big.NewInt(135000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Not success
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallPush(false, 4 /*nil,*/, big.NewInt(246000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Zero
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallPush(false, 0 /*errors.New("error"),*/, big.NewInt(357000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetOracleLockContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.True(t, calls[0].Success)
	require.Equal(t, 3, calls[0].OracleVotingResult)
	require.Equal(t, "0.000000000000135", calls[0].Transfer.String())

	require.Equal(t, 1, calls[1].ContractTxId)
	require.Equal(t, 4, calls[1].TxId)
	require.False(t, calls[1].Success)
	require.Equal(t, 4, calls[1].OracleVotingResult)
	require.Equal(t, "0.000000000000246", calls[1].Transfer.String())

	require.Equal(t, 1, calls[2].ContractTxId)
	require.Equal(t, 5, calls[2].TxId)
	require.False(t, calls[2].Success)
	require.Zero(t, calls[2].OracleVotingResult)
	require.Equal(t, "0.000000000000357", calls[2].Transfer.String())
}

func Test_OracleLockContractTerminationToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(destAddr, big.NewInt(1))

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetOracleLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_OracleLockContractTerminationToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetOracleLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_OracleLockContractReset(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	contractAddress := tests.GetRandAddr()
	deployOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr())

	// Call check oracle voting
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallCheckOracleVoting(3, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Call push
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockCallPush(true, 3 /*nil,*/, big.NewInt(135000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Terminate
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(5)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleLockTermination(tests.GetRandAddr())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	assertLens := func(receiptsLen, contractsLen, olContractsLen, callCheckOracleVotingsLen, callPushesLen, terminationsLen, addressesLen int) {
		txReceipts, err := testCommon.GetTxReceipts(db)
		require.Nil(t, err)
		require.Equal(t, receiptsLen, len(txReceipts))

		contracts, err := testCommon.GetContracts(db)
		require.Nil(t, err)
		require.Equal(t, contractsLen, len(contracts))

		olContracts, err := testCommon.GetOracleLockContracts(db)
		require.Nil(t, err)
		require.Equal(t, olContractsLen, len(olContracts))

		callCheckOracleVotings, err := testCommon.GetOracleLockContractCallCheckOracleVotings(db)
		require.Nil(t, err)
		require.Equal(t, callCheckOracleVotingsLen, len(callCheckOracleVotings))

		callPushes, err := testCommon.GetOracleLockContractCallPushes(db)
		require.Nil(t, err)
		require.Equal(t, callPushesLen, len(callPushes))

		terminations, err := testCommon.GetOracleLockContractTerminations(db)
		require.Nil(t, err)
		require.Equal(t, terminationsLen, len(terminations))

		addresses, err := testCommon.GetAddresses(db)
		require.Nil(t, err)
		require.Equal(t, addressesLen, len(addresses))
	}

	assertLens(5, 1, 1, 1, 1, 1, 8)

	require.Nil(t, dbAccessor.ResetTo(5))
	assertLens(5, 1, 1, 1, 1, 1, 8)

	require.Nil(t, dbAccessor.ResetTo(4))
	assertLens(4, 1, 1, 1, 1, 0, 6)

	require.Nil(t, dbAccessor.ResetTo(3))
	assertLens(3, 1, 1, 1, 0, 0, 6)

	require.Nil(t, dbAccessor.ResetTo(2))
	assertLens(2, 1, 1, 0, 0, 0, 5)

	require.Nil(t, dbAccessor.ResetTo(1))
	assertLens(0, 0, 0, 0, 0, 0, 1)

	require.Nil(t, dbAccessor.ResetTo(0))
	assertLens(0, 0, 0, 0, 0, 0, 0)
}
