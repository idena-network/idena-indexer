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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_RefundableOracleLockContractDeployWithExistingAddresses(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	// Add dest address to state
	appState := listener.NodeCtx().AppState
	appState.State.SetBalance(oracleVotingAddress, big.NewInt(1))
	appState.State.SetBalance(successAddress, big.NewInt(1))
	appState.State.SetBalance(failAddress, big.NewInt(1))
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress1, contractAddress2, oracleVotingAddress, &successAddress, &failAddress)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 5, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	olContracts, err := testCommon.GetRefundableOracleLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(olContracts))
	require.Equal(t, 1, olContracts[0].TxId)
	require.Equal(t, oracleVotingAddress.Hex(), olContracts[0].OracleVotingAddress)
	require.Equal(t, 2, olContracts[0].Value)
	require.Equal(t, successAddress.Hex(), *olContracts[0].SuccessAddress)
	require.Equal(t, failAddress.Hex(), *olContracts[0].FailAddress)
	require.Equal(t, 7, olContracts[0].RefundDelay)
	require.Equal(t, 8, olContracts[0].DepositDeadline)
	require.Equal(t, 9, olContracts[0].OracleVotingFee)

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

func Test_RefundableOracleLockContractDeployWithNotExistingAddresses(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2, oracleVotingAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress1, contractAddress2, oracleVotingAddress, &successAddress, &failAddress)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 5, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	olContracts, err := testCommon.GetRefundableOracleLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(olContracts))
	require.Equal(t, 1, olContracts[0].TxId)
	require.Equal(t, oracleVotingAddress.Hex(), olContracts[0].OracleVotingAddress)
	require.Equal(t, 2, olContracts[0].Value)
	require.Equal(t, successAddress.Hex(), *olContracts[0].SuccessAddress)
	require.Equal(t, failAddress.Hex(), *olContracts[0].FailAddress)
	require.Equal(t, 7, olContracts[0].RefundDelay)
	require.Equal(t, 8, olContracts[0].DepositDeadline)
	require.Equal(t, 9, olContracts[0].OracleVotingFee)

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

func Test_RefundableOracleLockContractDeployWithoutAddresses(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2, oracleVotingAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress1, contractAddress2, oracleVotingAddress, nil, nil)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 5, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	olContracts, err := testCommon.GetRefundableOracleLockContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(olContracts))
	require.Equal(t, 1, olContracts[0].TxId)
	require.Equal(t, oracleVotingAddress.Hex(), olContracts[0].OracleVotingAddress)
	require.Equal(t, 2, olContracts[0].Value)
	require.Nil(t, olContracts[0].SuccessAddress)
	require.Nil(t, olContracts[0].FailAddress)
	require.Equal(t, 7, olContracts[0].RefundDelay)
	require.Equal(t, 8, olContracts[0].DepositDeadline)
	require.Equal(t, 9, olContracts[0].OracleVotingFee)

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

func deployRefundableOracleLockContracts(t *testing.T, listener incoming.Listener, bus eventbus.Bus, addr1, addr2, oracleVotingAddress common.Address, successAddress, failAddress *common.Address) {
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

	var sa, fa common.Address
	var saErr, faErr error
	if successAddress != nil {
		sa = *successAddress
	} else {
		saErr = errors.New("error")
	}
	if failAddress != nil {
		fa = *failAddress
	} else {
		faErr = errors.New("error")
	}
	// Success
	tx := &types.Transaction{AccountNonce: 1}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockDeploy(addr1, oracleVotingAddress, 2, sa, saErr,
		fa, faErr, 7, 8, 9, 10, big.NewInt(789000))
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100), ContractAddress: addr1}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Failed tx receipt
	tx = &types.Transaction{AccountNonce: 2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockDeploy(addr2, oracleVotingAddress, 3, sa, saErr,
		fa, faErr, 17, 18, 19, 110, big.NewInt(1789000))
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), Error: errors.New("error message")}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
}

func Test_RefundableOracleLockContractCallDeposit(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call deposits
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallDeposit(big.NewInt(200), big.NewInt(300), big.NewInt(4000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Fail
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallDeposit(big.NewInt(1200), big.NewInt(1300), big.NewInt(14000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash()}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetRefundableOracleLockContractCallDeposits(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, "0.0000000000000002", calls[0].OwnSum.String())
	require.Equal(t, "0.0000000000000003", calls[0].Sum.String())
	require.Equal(t, "0.000000000000004", calls[0].Fee.String())

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(txReceipts))
}

func Test_RefundableOracleLockContractCallPush(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call pushes
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallPush(7, true, 5, nil, big.NewInt(5000), 120)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Fail
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallPush(0, false, 0, nil, new(big.Int), 0)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetRefundableOracleLockContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.True(t, calls[0].OracleVotingExists)
	require.Equal(t, 5, *calls[0].OracleVotingResult)
	require.Equal(t, "0.000000000000005", calls[0].Transfer.String())
	require.Equal(t, 120, *calls[0].RefundBlock)

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(txReceipts))
}

func Test_RefundableOracleLockContractCallPushNilResult(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call success
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallPush(7, true, 0, errors.New("no value"), big.NewInt(5000), 120)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetRefundableOracleLockContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.True(t, calls[0].OracleVotingExists)
	require.Nil(t, calls[0].OracleVotingResult)
	require.Equal(t, "0.000000000000005", calls[0].Transfer.String())
	require.Equal(t, 120, *calls[0].RefundBlock)
}

func Test_RefundableOracleLockContractCallPushZeroRefundBlock(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call success
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallPush(7, true, 5, nil, big.NewInt(5000), 0)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetRefundableOracleLockContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.True(t, calls[0].OracleVotingExists)
	require.Equal(t, 5, *calls[0].OracleVotingResult)
	require.Equal(t, "0.000000000000005", calls[0].Transfer.String())
	require.Nil(t, calls[0].RefundBlock)
}

func Test_RefundableOracleLockContractCallRefund(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call success
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Success
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallRefund(big.NewInt(200), decimal.RequireFromString("0.12345"))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Fail
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallRefund(big.NewInt(1200), decimal.RequireFromString("1.12345"))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetRefundableOracleLockContractCallRefunds(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, "0.0000000000000002", calls[0].Balance.String())
	require.Equal(t, 0.12345, calls[0].Coef)

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(txReceipts))
}

func Test_RefundableOracleLockContractTerminationToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(destAddr, big.NewInt(1))

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetRefundableOracleLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_RefundableOracleLockContractTerminationToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()

	// Deploy contract
	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetRefundableOracleLockContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test__RefundableOracleLockContractReset(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	contractAddress, successAddress, failAddress := tests.GetRandAddr(), tests.GetRandAddr(), tests.GetRandAddr()
	deployRefundableOracleLockContracts(t, listener, bus, contractAddress, tests.GetRandAddr(), tests.GetRandAddr(), &successAddress, &failAddress)

	// Call deposit
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallDeposit(big.NewInt(200), big.NewInt(300), big.NewInt(4000))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Call push
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallPush(7, true, 5, nil, big.NewInt(5000), 120)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Call refund
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(5)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockCallRefund(big.NewInt(200), decimal.RequireFromString("0.12345"))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Terminate
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(6)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddRefundableOracleLockTermination(tests.GetRandAddr())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	assertLens := func(receiptsLen, contractsLen, olContractsLen, callDepositsLen, callPushesLen, callRefundsLen, terminationsLen, addressesLen int) {
		txReceipts, err := testCommon.GetTxReceipts(db)
		require.Nil(t, err)
		require.Equal(t, receiptsLen, len(txReceipts))

		contracts, err := testCommon.GetContracts(db)
		require.Nil(t, err)
		require.Equal(t, contractsLen, len(contracts))

		olContracts, err := testCommon.GetRefundableOracleLockContracts(db)
		require.Nil(t, err)
		require.Equal(t, olContractsLen, len(olContracts))

		callDeposits, err := testCommon.GetRefundableOracleLockContractCallDeposits(db)
		require.Nil(t, err)
		require.Equal(t, callDepositsLen, len(callDeposits))

		callPushes, err := testCommon.GetRefundableOracleLockContractCallPushes(db)
		require.Nil(t, err)
		require.Equal(t, callPushesLen, len(callPushes))

		callRefunds, err := testCommon.GetRefundableOracleLockContractCallRefunds(db)
		require.Nil(t, err)
		require.Equal(t, callRefundsLen, len(callRefunds))

		terminations, err := testCommon.GetRefundableOracleLockContractTerminations(db)
		require.Nil(t, err)
		require.Equal(t, terminationsLen, len(terminations))

		addresses, err := testCommon.GetAddresses(db)
		require.Nil(t, err)
		require.Equal(t, addressesLen, len(addresses))
	}

	assertLens(6, 1, 1, 1, 1, 1, 1, 10)

	require.Nil(t, dbAccessor.ResetTo(6))
	assertLens(6, 1, 1, 1, 1, 1, 1, 10)

	require.Nil(t, dbAccessor.ResetTo(5))
	assertLens(5, 1, 1, 1, 1, 1, 0, 8)

	require.Nil(t, dbAccessor.ResetTo(4))
	assertLens(4, 1, 1, 1, 1, 0, 0, 7)

	require.Nil(t, dbAccessor.ResetTo(3))
	assertLens(3, 1, 1, 1, 0, 0, 0, 6)

	require.Nil(t, dbAccessor.ResetTo(2))
	assertLens(2, 1, 1, 0, 0, 0, 0, 5)

	require.Nil(t, dbAccessor.ResetTo(1))
	assertLens(0, 0, 0, 0, 0, 0, 0, 1)

	require.Nil(t, dbAccessor.ResetTo(0))
	assertLens(0, 0, 0, 0, 0, 0, 0, 0)
}
