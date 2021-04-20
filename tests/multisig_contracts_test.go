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

func Test_MultisigContractDeploy(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	// When
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress1, contractAddress2)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 4, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	msContracts, err := testCommon.GetMultisigContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(msContracts))
	require.Equal(t, 1, msContracts[0].TxId)
	require.Equal(t, 3, msContracts[0].MinVotes)
	require.Equal(t, 120, msContracts[0].MaxVotes)
	require.Equal(t, 1, msContracts[0].State)

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

func deployMultisigContracts(t *testing.T, listener incoming.Listener, bus eventbus.Bus, addr1, addr2 common.Address) {
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
	statsCollector.AddMultisigDeploy(addr1, 3, 120, 1)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100), ContractAddress: addr1}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Failed tx receipt
	tx = &types.Transaction{AccountNonce: 2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigDeploy(addr2, 4, 121, 1)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), Error: errors.New("error message")}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
}

func Test_MultisigContractCallAddExistingAddressAndNoNewState(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	addr := tests.GetRandAddr()
	// Add address to state
	appState.State.SetBalance(addr, big.NewInt(1))

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigCallAdd(addr, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallAdds(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Nil(t, calls[0].NewState)
}

func Test_MultisigContractCallAddNotExistingAddressAndNewState(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	addr := tests.GetRandAddr()
	state := byte(2)
	statsCollector.AddMultisigCallAdd(addr, &state)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallAdds(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Equal(t, 2, *calls[0].NewState)
}

func Test_MultisigContractCallSendToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	addr := tests.GetRandAddr()
	// Add address to state
	appState.State.SetBalance(addr, big.NewInt(1))

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigCallSend(addr, big.NewInt(456000).Bytes())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallSends(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Equal(t, "0.000000000000456", calls[0].Amount.String())
}

func Test_MultisigContractCallSendToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	addr := tests.GetRandAddr()
	statsCollector.AddMultisigCallSend(addr, big.NewInt(456000).Bytes())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallSends(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Equal(t, "0.000000000000456", calls[0].Amount.String())
}

func Test_MultisigContractCallPushToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	addr := tests.GetRandAddr()
	// Add address to state
	appState.State.SetBalance(addr, big.NewInt(1))

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigCallPush(addr, big.NewInt(456000).Bytes(), 5, 6)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Equal(t, "0.000000000000456", calls[0].Amount.String())
	require.Equal(t, 5, calls[0].VoteAddressCnt)
	require.Equal(t, 6, calls[0].VoteAmountCnt)
}

func Test_MultisigContractCallPushToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Call
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	addr := tests.GetRandAddr()
	statsCollector.AddMultisigCallPush(addr, big.NewInt(456000).Bytes(), 5, 6)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	calls, err := testCommon.GetMultisigContractCallPushes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(calls))
	require.Equal(t, 1, calls[0].ContractTxId)
	require.Equal(t, 3, calls[0].TxId)
	require.Equal(t, addr.Hex(), calls[0].Address)
	require.Equal(t, "0.000000000000456", calls[0].Amount.String())
	require.Equal(t, 5, calls[0].VoteAddressCnt)
	require.Equal(t, 6, calls[0].VoteAmountCnt)
}

func Test_MultisigContractTerminationToExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(destAddr, big.NewInt(1))

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetMultisigContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_MultisigContractTerminationToNotExistingAddress(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	destAddr := tests.GetRandAddr()

	// Deploy contract
	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())

	// Terminate
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigTermination(destAddr)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetMultisigContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 3, terminations[0].TxId)
	require.Equal(t, destAddr.Hex(), terminations[0].Dest)
}

func Test_MultisigContractReset(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	contractAddress := tests.GetRandAddr()
	deployMultisigContracts(t, listener, bus, contractAddress, tests.GetRandAddr())
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Call add
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 3, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	state := byte(2)
	statsCollector.AddMultisigCallAdd(tests.GetRandAddr(), &state)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Call send
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigCallSend(tests.GetRandAddr(), big.NewInt(456000).Bytes())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Call push
	respondentKey, _ = crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height = uint64(5)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddMultisigCallPush(tests.GetRandAddr(), big.NewInt(456000).Bytes(), 5, 6)
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
	statsCollector.AddMultisigTermination(tests.GetRandAddr())
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	assertLens := func(receiptsLen, contractsLen, multisigContractsLen, callAddsLen, callSendsLen, callPushesLen, terminationsLen, addressesLen int) {
		txReceipts, err := testCommon.GetTxReceipts(db)
		require.Nil(t, err)
		require.Equal(t, receiptsLen, len(txReceipts))

		contracts, err := testCommon.GetContracts(db)
		require.Nil(t, err)
		require.Equal(t, contractsLen, len(contracts))

		msContracts, err := testCommon.GetMultisigContracts(db)
		require.Nil(t, err)
		require.Equal(t, multisigContractsLen, len(msContracts))

		callAdds, err := testCommon.GetMultisigContractCallAdds(db)
		require.Nil(t, err)
		require.Equal(t, callAddsLen, len(callAdds))

		callSends, err := testCommon.GetMultisigContractCallSends(db)
		require.Nil(t, err)
		require.Equal(t, callSendsLen, len(callSends))

		callPushes, err := testCommon.GetMultisigContractCallPushes(db)
		require.Nil(t, err)
		require.Equal(t, callPushesLen, len(callPushes))

		terminations, err := testCommon.GetMultisigContractTerminations(db)
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
	assertLens(4, 1, 1, 1, 1, 0, 0, 6)

	require.Nil(t, dbAccessor.ResetTo(3))
	assertLens(3, 1, 1, 1, 0, 0, 0, 4)

	require.Nil(t, dbAccessor.ResetTo(2))
	assertLens(2, 1, 1, 0, 0, 0, 0, 2)

	require.Nil(t, dbAccessor.ResetTo(1))
	assertLens(0, 0, 0, 0, 0, 0, 0, 1)

	require.Nil(t, dbAccessor.ResetTo(0))
	assertLens(0, 0, 0, 0, 0, 0, 0, 0)
}
