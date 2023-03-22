package tests

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func Test_ContractTxBalanceUpdates(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	balancesCache := make(map[common.Address]*big.Int)

	addressToUpdateBalance, unknownAddressToUpdateBalance := tests.GetRandAddr(), tests.GetRandAddr()
	// Add dest address to state
	appState.State.SetBalance(addressToUpdateBalance, big.NewInt(1))

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	appState.State.DeployContract(contractAddress, common.Hash{0x1}, big.NewInt(100))
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Send tx 1
	senderKey, _ := crypto.GenerateKey()
	senderAddress := crypto.PubkeyToAddress(senderKey.PublicKey)

	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress, Type: types.CallContractTx}, senderKey)
	statsCollector.BeginApplyingTx(tx, appState)

	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(senderAddress, big.NewInt(100))
	appState.State.SetBalance(contractAddress, big.NewInt(200))
	statsCollector.CompleteBalanceUpdate(appState)

	statsCollector.AddContractBalanceUpdate(nil, addressToUpdateBalance, func(address common.Address) *big.Int {
		return big.NewInt(300)
	}, big.NewInt(400), appState, &balancesCache)
	statsCollector.AddContractBalanceUpdate(nil, senderAddress, func(address common.Address) *big.Int {
		return big.NewInt(500)
	}, big.NewInt(600), appState, &balancesCache)

	statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5}, nil, 1, nil, nil, nil, nil, false)
	statsCollector.ApplyContractBalanceUpdates(&balancesCache, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send tx 2
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress, Type: types.CallContractTx}, senderKey)
	statsCollector.BeginApplyingTx(tx, appState)

	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(senderAddress, big.NewInt(100))
	appState.State.SetBalance(contractAddress, big.NewInt(200))
	statsCollector.CompleteBalanceUpdate(appState)

	statsCollector.AddContractBalanceUpdate(nil, unknownAddressToUpdateBalance, func(address common.Address) *big.Int {
		return big.NewInt(700)
	}, big.NewInt(800), appState, &balancesCache)
	statsCollector.AddContractBalanceUpdate(nil, senderAddress, func(address common.Address) *big.Int {
		return big.NewInt(900)
	}, big.NewInt(1000), appState, &balancesCache)

	statsCollector.AddOracleVotingTermination(big.NewInt(0), nil, nil)
	statsCollector.ApplyContractBalanceUpdates(&balancesCache, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send fail tx
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress, Type: types.CallContractTx}, senderKey)
	statsCollector.BeginApplyingTx(tx, appState)

	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(senderAddress, big.NewInt(100))
	appState.State.SetBalance(contractAddress, big.NewInt(200))
	statsCollector.CompleteBalanceUpdate(appState)

	statsCollector.AddContractBalanceUpdate(nil, senderAddress, func(address common.Address) *big.Int {
		return big.NewInt(900)
	}, big.NewInt(1000), appState, &balancesCache)

	statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5}, nil, 1, nil, nil, nil, nil, false)

	// Rollback tx balance updates
	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(senderAddress, big.NewInt(0))
	appState.State.SetBalance(contractAddress, big.NewInt(0))
	statsCollector.CompleteBalanceUpdate(appState)
	statsCollector.ApplyContractBalanceUpdates(&balancesCache, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send SendTx to contract address
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 7, To: &contractAddress}, senderKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send SendTx to non-contract address
	rndAddress := tests.GetRandAddr()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 8, To: &rndAddress}, senderKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	balanceUpdates, err := testCommon.GetContractTxBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 8, len(balanceUpdates))

	require.Equal(t, 4, balanceUpdates[2].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[2].ContractAddress)
	require.Equal(t, addressToUpdateBalance.Hex(), balanceUpdates[2].Address)
	require.Equal(t, 2, balanceUpdates[2].ContractType)
	require.Equal(t, 4, balanceUpdates[2].TxId)
	require.Equal(t, 2, *balanceUpdates[2].CallMethod)
	require.Equal(t, "0.0000000000000003", balanceUpdates[2].BalanceOld.String())
	require.Equal(t, "0.0000000000000004", balanceUpdates[2].BalanceNew.String())

	require.Equal(t, 3, balanceUpdates[3].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[3].ContractAddress)
	require.Equal(t, senderAddress.Hex(), balanceUpdates[3].Address)
	require.Equal(t, 2, balanceUpdates[3].ContractType)
	require.Equal(t, 4, balanceUpdates[3].TxId)
	require.Equal(t, 2, *balanceUpdates[3].CallMethod)
	require.Equal(t, "0.0000000000000005", balanceUpdates[3].BalanceOld.String())
	require.Equal(t, "0.0000000000000006", balanceUpdates[3].BalanceNew.String())

	require.Equal(t, 5, balanceUpdates[4].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[4].ContractAddress)
	require.Equal(t, senderAddress.Hex(), balanceUpdates[4].Address)
	require.Equal(t, 2, balanceUpdates[4].ContractType)
	require.Equal(t, 5, balanceUpdates[4].TxId)
	require.Nil(t, balanceUpdates[4].CallMethod)
	require.Equal(t, "0.0000000000000009", balanceUpdates[4].BalanceOld.String())
	require.Equal(t, "0.000000000000001", balanceUpdates[4].BalanceNew.String())

	require.Equal(t, 6, balanceUpdates[5].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[5].ContractAddress)
	require.Equal(t, unknownAddressToUpdateBalance.Hex(), balanceUpdates[5].Address)
	require.Equal(t, 2, balanceUpdates[5].ContractType)
	require.Equal(t, 5, balanceUpdates[5].TxId)
	require.Nil(t, balanceUpdates[5].CallMethod)
	require.Equal(t, "0.0000000000000007", balanceUpdates[5].BalanceOld.String())
	require.Equal(t, "0.0000000000000008", balanceUpdates[5].BalanceNew.String())

	require.Equal(t, 7, balanceUpdates[6].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[6].ContractAddress)
	require.Equal(t, senderAddress.Hex(), balanceUpdates[6].Address)
	require.Equal(t, 2, balanceUpdates[6].ContractType)
	require.Equal(t, 6, balanceUpdates[6].TxId)
	require.Nil(t, balanceUpdates[6].CallMethod)
	require.Nil(t, balanceUpdates[6].BalanceOld)
	require.Nil(t, balanceUpdates[6].BalanceNew)

	require.Equal(t, 8, balanceUpdates[7].Id)
	require.Equal(t, contractAddress.Hex(), balanceUpdates[7].ContractAddress)
	require.Equal(t, senderAddress.Hex(), balanceUpdates[7].Address)
	require.Equal(t, 2, balanceUpdates[7].ContractType)
	require.Equal(t, 7, balanceUpdates[7].TxId)
	require.Nil(t, balanceUpdates[7].CallMethod)
	require.Nil(t, balanceUpdates[7].BalanceOld)
	require.Nil(t, balanceUpdates[7].BalanceNew)

	require.Nil(t, dbAccessor.ResetTo(3))
	balanceUpdates, err = testCommon.GetContractTxBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 8, len(balanceUpdates))

	require.Nil(t, dbAccessor.ResetTo(2))
	balanceUpdates, err = testCommon.GetContractTxBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(balanceUpdates))

	require.Nil(t, dbAccessor.ResetTo(1))
	balanceUpdates, err = testCommon.GetContractTxBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(balanceUpdates))
}
