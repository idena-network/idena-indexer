package tests

import (
	"database/sql"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func Test_ResetOracleVotingResultsAndSummariesOld(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 5, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Block 3
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Send vote proof
	respondentKey, _ := crypto.GenerateKey()
	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteProofOld([]byte{0x3, 0x4})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send vote 1
	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteOld(7, []byte{0x4, 0x5})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Block 4
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	// Send votes 2, 3
	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteOld(8, []byte{0x4, 0x5})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 7, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteOld(7, []byte{0x4, 0x5})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	results, err := testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(results))
	require.Equal(t, 2, results[0].Count)
	require.Equal(t, 1, results[1].Count)
	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 3, summaries[0].Votes)
	sChanges, err := testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(sChanges))
	rChanges, err := testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(4))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(results))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 3, summaries[0].Votes)
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(3))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1, results[0].Count)
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 1, summaries[0].Votes)
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(2))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(results))
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(rChanges))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 0, summaries[0].VoteProofs)
	require.Equal(t, 0, summaries[0].Votes)

	require.Nil(t, dbAccessor.ResetTo(1))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(summaries))
}

func Test_ResetOracleVotingResultsAndSummaries(t *testing.T) {
	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 5, testCommon.PostgresSchema, "..")
	defer listener.Destroy()
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Block 3
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	// Send vote proof
	respondentKey, _ := crypto.GenerateKey()
	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteProof([]byte{0x3, 0x4}, pUint64(99))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Send vote 1
	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5}, pUint64(1), 1, pUint64(2), nil, nil, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Block 4
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	// Send votes 2, 3
	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVote(8, []byte{0x4, 0x5}, pUint64(1), 1, pUint64(1), nil, nil, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	respondentKey, _ = crypto.GenerateKey()
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 7, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	delegatee := tests.GetRandAddr()
	statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5}, pUint64(2), 2, pUint64(0), &delegatee, []byte{9}, pUint64(10))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	results, err := testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(results))
	require.Equal(t, 2, results[0].Count)
	require.Equal(t, 1, results[1].Count)
	require.Equal(t, 10, results[2].Count)
	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 3, summaries[0].Votes)
	require.Equal(t, 0, *summaries[0].SecretVotesCount)

	sChanges, err := testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(sChanges))
	rChanges, err := testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(4))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(results))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 3, summaries[0].Votes)
	require.Equal(t, 0, *summaries[0].SecretVotesCount)
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 4, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(3))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1, results[0].Count)
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 1, summaries[0].Votes)
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(rChanges))

	require.Nil(t, dbAccessor.ResetTo(2))
	results, err = testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(results))
	sChanges, err = testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sChanges))
	rChanges, err = testCommon.GetOracleVotingContractResultChanges(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(rChanges))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 0, summaries[0].VoteProofs)
	require.Equal(t, 0, summaries[0].Votes)

	require.Nil(t, dbAccessor.ResetTo(1))
	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(summaries))
}

func Test_ResetOracleVotingContractsOld(t *testing.T) {

	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 1000, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	respondentKey1, _ := crypto.GenerateKey()
	respondentAddress1 := crypto.PubkeyToAddress(respondentKey1.PublicKey)
	respondentKey2, _ := crypto.GenerateKey()
	respondentAddress2 := crypto.PubkeyToAddress(respondentKey2.PublicKey)
	respondentKey3, _ := crypto.GenerateKey()
	respondentAddress3 := crypto.PubkeyToAddress(respondentKey3.PublicKey)
	respondentKey4, _ := crypto.GenerateKey()
	respondentAddress4 := crypto.PubkeyToAddress(respondentKey4.PublicKey)

	indexData := func() {

		statsCollector := listener.StatsCollector()
		appState := listener.NodeCtx().AppState
		appState.State.SetState(respondentAddress1, state.Verified)
		appState.State.SetPubKey(respondentAddress1, []byte{0x1, 0x2})
		appState.State.SetState(respondentAddress2, state.Verified)
		appState.State.SetPubKey(respondentAddress2, []byte{0x2, 0x3})
		appState.State.SetState(respondentAddress3, state.Verified)
		appState.State.SetPubKey(respondentAddress3, []byte{0x3, 0x4})
		appState.State.SetState(respondentAddress4, state.Verified)
		appState.State.SetPubKey(respondentAddress4, []byte{0x5, 0x6})

		// Deploy contracts
		startTime := time.Now().UTC()
		contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
		deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

		// Deploy one more contract
		statsCollector.EnableCollecting()
		height := uint64(3)
		block := buildBlock(height)
		tx := &types.Transaction{AccountNonce: 4}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingDeploy(tests.GetRandAddr(), uint64(startTime.Unix()), new(big.Int).SetUint64(23400), []byte{0x1, 0x2},
			0, 1, 2, 3, 4, 5, 7)
		statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100)}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call Start
		statsCollector.EnableCollecting()
		height = uint64(4)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 5, To: &contractAddress1}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 45, 100)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call Start
		statsCollector.EnableCollecting()
		height = uint64(5)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 6, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallStart(1, 6, 3, new(big.Int).SetUint64(100), []byte{0x3, 0x4}, 50, 100)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote proof
		statsCollector.EnableCollecting()
		height = uint64(6)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 7, To: &contractAddress2}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProofOld([]byte{0x3, 0x4})
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote proofs
		statsCollector.EnableCollecting()
		height = uint64(7)
		block = buildBlock(height)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 8, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProofOld([]byte{0x4, 0x5})
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 9, To: &contractAddress2}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProofOld([]byte{0x5, 0x6})
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 10, To: &contractAddress1}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProofOld([]byte{0x6, 0x7})
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote
		statsCollector.EnableCollecting()
		height = uint64(8)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 11, To: &contractAddress1}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteOld(7, []byte{0x4, 0x5})
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call finish
		statsCollector.EnableCollecting()
		height = uint64(9)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 12, To: &contractAddress1}
		statsCollector.BeginApplyingTx(tx, appState)
		result := byte(5)
		statsCollector.AddOracleVotingCallFinish(2, &result, big.NewInt(500), big.NewInt(700), big.NewInt(800))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Prolong with no startBlock
		statsCollector.EnableCollecting()
		height = uint64(10)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 13, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallProlongationOld(nil, 123, []byte{0x1, 0x2}, 999, 999)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Reach 17 block to switch contract (contractAddress2) to counting state (startBlockHeight (6) + votingDuration (11) = 17)
		for i := uint64(11); i <= 17; i++ {
			statsCollector.EnableCollecting()
			height = i
			block = buildBlock(height)
			require.Nil(t, applyBlock(bus, block, appState))
			statsCollector.CompleteCollecting()
		}

		// Prolong with startBlock and send coins to contract
		statsCollector.EnableCollecting()
		height = uint64(18)
		block = buildBlock(height)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 14, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		startBlock := uint64(19)
		statsCollector.AddOracleVotingCallProlongationOld(&startBlock, 123, []byte{0x1, 0x2}, 999, 999)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx = &types.Transaction{AccountNonce: 14, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.BeginTxBalanceUpdate(tx, appState)
		appState.State.SetBalance(contractAddress2, big.NewInt(12300))
		statsCollector.CompleteBalanceUpdate(appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call finish
		statsCollector.EnableCollecting()
		height = uint64(19)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 15, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallFinish(2, nil, big.NewInt(500), big.NewInt(700), big.NewInt(800))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Send coins to completed contract
		statsCollector.EnableCollecting()
		height = uint64(20)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 16, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.BeginTxBalanceUpdate(tx, appState)
		appState.State.SetBalance(contractAddress2, big.NewInt(12300))
		statsCollector.CompleteBalanceUpdate(appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Terminate
		statsCollector.EnableCollecting()
		height = uint64(21)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 17, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingTermination(big.NewInt(7800), nil, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()
	}

	var sovc []testCommon.SortedOracleVotingContract
	var sovcc []testCommon.SortedOracleVotingContractCommittee
	var sovccOracle *testCommon.SortedOracleVotingContractCommittee

	assertions := []assertion{
		{
			heightToReset: 0,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{})
			},
		},
		{
			heightToReset: 1,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{addresses: 5})
			},
		},
		{
			heightToReset: 2,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 3, contracts: 2, ovContracts: 2, addresses: 7,
					sorted: 2, summaries: 2})
			},
		},
		{
			heightToReset: 3,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 4, contracts: 3, ovContracts: 3, addresses: 8,
					sorted: 3, summaries: 3})
			},
		},
		{
			heightToReset: 4,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 5, contracts: 3, ovContracts: 3, callStarts: 1,
					committees: 1, addresses: 8, sorted: 3, sortedCommittees: 1, sortedChanges: 1,
					sortedCommitteesChanges: 1, summaries: 3})
			},
		},
		{
			heightToReset: 5,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 6, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 4,
					summaries: 3})
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				sorted, _ := testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 124, *sorted[0].CountingBlock)
				require.Equal(t, 1, sorted[0].State)
				require.Equal(t, 17, *sorted[1].CountingBlock)
				require.Equal(t, 1, sorted[1].State)
			},
		},
		{
			heightToReset: 6,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 7, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 5,
					callVoteProofs: 1, summaries: 3, summariesChanges: 1})

				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)

				sovccChanges, _ := testCommon.GetSortedOracleVotingContractCommitteeChanges(db)
				require.False(t, *sovccChanges[4].Deleted)
			},
		},
		{
			heightToReset: 7,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 10, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 7,
					callVoteProofs: 4, summaries: 3, summariesChanges: 3})
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
			},
		},
		{
			heightToReset: 8,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 11, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 7,
					callVoteProofs: 4, callVotes: 1, summaries: 3, summariesChanges: 4})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 1, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sovc[0].SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
			},
		},
		{
			heightToReset: 9,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 12, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 3, sortedCommitteesChanges: 8,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.Equal(t, 2, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
			},
		},
		{
			heightToReset: 10,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 4, sortedCommitteesChanges: 11,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
			},
		},
		{
			heightToReset: 16,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 4, sortedCommitteesChanges: 11,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && (oracle.State == 1 || oracle.State == 5) || oracle.TxId == 1 && oracle.State == 2)
				}
			},
		},
		{
			heightToReset: 17,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 5, sortedCommitteesChanges: 15,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 3, sovc[1].State)
				require.Equal(t, 17, *sovc[1].CountingBlock)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && oracle.State == 3 || oracle.TxId == 1 && oracle.State == 2)
				}
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress3)
				require.False(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress4)
				require.False(t, sovccOracle.Voted)
			},
		},
		{
			heightToReset: 18,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 14, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 7, sortedCommitteesChanges: 27,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 2, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				require.Equal(t, 30, *sovc[1].CountingBlock)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && (oracle.State == 1 && !oracle.Voted || oracle.State == 5 && oracle.Voted) || oracle.TxId == 1 && oracle.State == 2)
				}
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress3)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				require.Equal(t, "000000000000000000000000000000.0000000000000006980000000000000000003", *sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress4)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				require.Equal(t, "000000000000000000000000000000.0000000000000006980000000000000000003", *sovccOracle.SortKey)
			},
		},
		{
			heightToReset: 19,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 15, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 8, sortedCommitteesChanges: 31,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, summaries: 3, summariesChanges: 6})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Nil(t, sovc[0].SortKey)
				require.Equal(t, 2, sovc[1].State)
				require.Nil(t, sovc[1].SortKey)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && oracle.State == 2 || oracle.TxId == 1 && oracle.State == 2)
				}
			},
		},
		{
			heightToReset: 20,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 15, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 8, sortedCommitteesChanges: 31,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, summaries: 3, summariesChanges: 6})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Nil(t, sovc[0].SortKey)
				require.Nil(t, sovc[1].SortKey)
			},
		},
		{
			heightToReset: 21,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 16, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 3, sortedChanges: 9, sortedCommitteesChanges: 35,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, terminations: 1, summaries: 3, summariesChanges: 7})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Nil(t, sovc[0].SortKey)
				require.Nil(t, sovc[1].SortKey)
			},
		}}

	indexData()
	for i := len(assertions) - 1; i >= 0; i-- {
		require.Nil(t, dbAccessor.ResetTo(assertions[i].heightToReset))
		assertions[i].assert()
	}
}

func Test_ResetOracleVotingContracts(t *testing.T) {

	db, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 1000, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

	respondentKey1, _ := crypto.GenerateKey()
	respondentAddress1 := crypto.PubkeyToAddress(respondentKey1.PublicKey)
	respondentKey2, _ := crypto.GenerateKey()
	respondentAddress2 := crypto.PubkeyToAddress(respondentKey2.PublicKey)
	respondentKey3, _ := crypto.GenerateKey()
	respondentAddress3 := crypto.PubkeyToAddress(respondentKey3.PublicKey)
	respondentKey4, _ := crypto.GenerateKey()
	respondentAddress4 := crypto.PubkeyToAddress(respondentKey4.PublicKey)

	indexData := func() {

		statsCollector := listener.StatsCollector()
		appState := listener.NodeCtx().AppState
		appState.State.SetState(respondentAddress1, state.Verified)
		appState.State.SetPubKey(respondentAddress1, []byte{0x1, 0x2})
		appState.State.SetState(respondentAddress2, state.Verified)
		appState.State.SetPubKey(respondentAddress2, []byte{0x2, 0x3})
		appState.State.SetState(respondentAddress3, state.Verified)
		appState.State.SetPubKey(respondentAddress3, []byte{0x3, 0x4})
		appState.State.SetState(respondentAddress4, state.Verified)
		appState.State.SetPubKey(respondentAddress4, []byte{0x5, 0x6})

		// Deploy contracts
		startTime := time.Now().UTC()
		contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
		deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

		// Deploy one more contract
		statsCollector.EnableCollecting()
		height := uint64(3)
		block := buildBlock(height)
		tx := &types.Transaction{AccountNonce: 4}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingDeploy(tests.GetRandAddr(), uint64(startTime.Unix()), new(big.Int).SetUint64(23400), []byte{0x1, 0x2},
			0, 1, 2, 3, 4, 5, 7)
		statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100)}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call Start
		statsCollector.EnableCollecting()
		height = uint64(4)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 5, To: &contractAddress1}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 45, 100)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call Start
		statsCollector.EnableCollecting()
		height = uint64(5)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 6, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallStart(1, 6, 3, new(big.Int).SetUint64(100), []byte{0x3, 0x4}, 50, 100)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote proof
		statsCollector.EnableCollecting()
		height = uint64(6)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 7, To: &contractAddress2}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProof([]byte{0x3, 0x4}, pUint64(1))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote proofs
		statsCollector.EnableCollecting()
		height = uint64(7)
		block = buildBlock(height)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 8, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProof([]byte{0x4, 0x5}, pUint64(2))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 9, To: &contractAddress2}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProof([]byte{0x5, 0x6}, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 10, To: &contractAddress1}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVoteProof([]byte{0x6, 0x7}, pUint64(1))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Vote
		statsCollector.EnableCollecting()
		height = uint64(8)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 11, To: &contractAddress1}, respondentKey1)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5}, pUint64(1), 1, pUint64(0), nil, nil, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call finish
		statsCollector.EnableCollecting()
		height = uint64(9)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 12, To: &contractAddress1}
		statsCollector.BeginApplyingTx(tx, appState)
		result := byte(5)
		statsCollector.AddOracleVotingCallFinish(2, &result, big.NewInt(500), big.NewInt(700), big.NewInt(800))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Prolong with no startBlock
		statsCollector.EnableCollecting()
		height = uint64(10)
		block = buildBlock(height)
		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 13, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallProlongation(nil, 123, []byte{0x1, 0x2}, 999, 999, nil, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Reach 17 block to switch contract (contractAddress2) to counting state (startBlockHeight (6) + votingDuration (11) = 17)
		for i := uint64(11); i <= 17; i++ {
			statsCollector.EnableCollecting()
			height = i
			block = buildBlock(height)
			require.Nil(t, applyBlock(bus, block, appState))
			statsCollector.CompleteCollecting()
		}

		// Prolong with startBlock and newEpochWithoutGrowth and send coins to contract
		statsCollector.EnableCollecting()
		height = uint64(18)
		block = buildBlock(height)

		tx, _ = types.SignTx(&types.Transaction{AccountNonce: 14, To: &contractAddress2}, respondentKey2)
		statsCollector.BeginApplyingTx(tx, appState)
		startBlock := uint64(19)
		newEpochWithoutGrowth := byte(1)
		statsCollector.AddOracleVotingCallProlongation(&startBlock, 123, []byte{0x1, 0x2}, 999, 999, &newEpochWithoutGrowth, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		tx = &types.Transaction{AccountNonce: 14, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.BeginTxBalanceUpdate(tx, appState)
		appState.State.SetBalance(contractAddress2, big.NewInt(12300))
		statsCollector.CompleteBalanceUpdate(appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Call finish
		statsCollector.EnableCollecting()
		height = uint64(19)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 15, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingCallFinish(2, nil, big.NewInt(500), big.NewInt(700), big.NewInt(800))
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Send coins to completed contract
		statsCollector.EnableCollecting()
		height = uint64(20)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 16, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.BeginTxBalanceUpdate(tx, appState)
		appState.State.SetBalance(contractAddress2, big.NewInt(12300))
		statsCollector.CompleteBalanceUpdate(appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		// Terminate
		statsCollector.EnableCollecting()
		height = uint64(21)
		block = buildBlock(height)
		tx = &types.Transaction{AccountNonce: 17, To: &contractAddress2}
		statsCollector.BeginApplyingTx(tx, appState)
		statsCollector.AddOracleVotingTermination(big.NewInt(7800), nil, nil)
		statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()
	}

	var sovc []testCommon.SortedOracleVotingContract
	var sovcc []testCommon.SortedOracleVotingContractCommittee
	var sovccOracle *testCommon.SortedOracleVotingContractCommittee

	assertions := []assertion{
		{
			heightToReset: 0,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{})
			},
		},
		{
			heightToReset: 1,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{addresses: 5})
			},
		},
		{
			heightToReset: 2,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 3, contracts: 2, ovContracts: 2, addresses: 7,
					sorted: 2, summaries: 2})
			},
		},
		{
			heightToReset: 3,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 4, contracts: 3, ovContracts: 3, addresses: 8,
					sorted: 3, summaries: 3})
			},
		},
		{
			heightToReset: 4,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 5, contracts: 3, ovContracts: 3, callStarts: 1,
					committees: 1, addresses: 8, sorted: 3, sortedCommittees: 1, sortedChanges: 1,
					sortedCommitteesChanges: 1, summaries: 3})
			},
		},
		{
			heightToReset: 5,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 6, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 4,
					summaries: 3})
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				sorted, _ := testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 124, *sorted[0].CountingBlock)
				require.Equal(t, 1, sorted[0].State)
				require.Equal(t, 17, *sorted[1].CountingBlock)
				require.Equal(t, 1, sorted[1].State)
			},
		},
		{
			heightToReset: 6,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 7, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 5,
					callVoteProofs: 1, summaries: 3, summariesChanges: 1})

				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)

				sovccChanges, _ := testCommon.GetSortedOracleVotingContractCommitteeChanges(db)
				require.False(t, *sovccChanges[4].Deleted)
			},
		},
		{
			heightToReset: 7,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 10, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 7,
					callVoteProofs: 4, summaries: 3, summariesChanges: 3})
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
			},
		},
		{
			heightToReset: 8,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 11, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 2, sortedCommitteesChanges: 7,
					callVoteProofs: 4, callVotes: 1, summaries: 3, summariesChanges: 4})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 1, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sovc[0].SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.Equal(t, 5, sovccOracle.State)
				require.True(t, sovccOracle.Voted)
			},
		},
		{
			heightToReset: 9,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 12, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 4, addresses: 8, sorted: 3, sortedCommittees: 4, sortedChanges: 3, sortedCommitteesChanges: 8,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 1, respondentAddress1)
				require.Equal(t, 2, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
			},
		},
		{
			heightToReset: 10,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 4, sortedCommitteesChanges: 11,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
			},
		},
		{
			heightToReset: 16,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 4, sortedCommitteesChanges: 11,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && (oracle.State == 1 || oracle.State == 5) || oracle.TxId == 1 && oracle.State == 2)
				}
			},
		},
		{
			heightToReset: 17,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 13, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 8, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 5, sortedCommitteesChanges: 15,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 1, summaries: 3, summariesChanges: 5})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 3, sovc[1].State)
				require.Equal(t, 17, *sovc[1].CountingBlock)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && oracle.State == 3 || oracle.TxId == 1 && oracle.State == 2)
				}
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress3)
				require.False(t, sovccOracle.Voted)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress4)
				require.False(t, sovccOracle.Voted)

				summaries, _ := testCommon.GetOracleVotingSummaries(db)
				summary := findOracleVotingSummary(summaries, 3)
				require.Nil(t, summary.EpochWithoutGrowth)
			},
		},
		{
			heightToReset: 18,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 14, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 7, sortedCommitteesChanges: 27,
					callVoteProofs: 4, callVotes: 1, callFinishes: 1, callProlongations: 2, summaries: 3, summariesChanges: 6})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Equal(t, 1, sovc[1].State)
				require.Equal(t, 30, *sovc[1].CountingBlock)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && (oracle.State == 1 && !oracle.Voted || oracle.State == 5 && oracle.Voted) || oracle.TxId == 1 && oracle.State == 2)
				}
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress1)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress2)
				require.True(t, sovccOracle.Voted)
				require.Equal(t, 5, sovccOracle.State)
				require.Nil(t, sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress3)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				require.Equal(t, "000000000000000000000000000000.0000000000000006980000000000000000003", *sovccOracle.SortKey)
				sovccOracle = findSortedOracleVotingContractCommittee(sovcc, 3, respondentAddress4)
				require.False(t, sovccOracle.Voted)
				require.Equal(t, 1, sovccOracle.State)
				require.Equal(t, "000000000000000000000000000000.0000000000000006980000000000000000003", *sovccOracle.SortKey)

				summaries, _ := testCommon.GetOracleVotingSummaries(db)
				summary := findOracleVotingSummary(summaries, 3)
				require.Equal(t, 1, *summary.EpochWithoutGrowth)
			},
		},
		{
			heightToReset: 19,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 15, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 8, sortedCommitteesChanges: 31,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, summaries: 3, summariesChanges: 7})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Equal(t, 2, sovc[0].State)
				require.Nil(t, sovc[0].SortKey)
				require.Equal(t, 2, sovc[1].State)
				require.Nil(t, sovc[1].SortKey)
				sovcc, _ = testCommon.GetSortedOracleVotingContractCommittees(db)
				for _, oracle := range sovcc {
					require.True(t, oracle.TxId == 3 && oracle.State == 2 || oracle.TxId == 1 && oracle.State == 2)
				}
			},
		},
		{
			heightToReset: 20,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 15, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 5, sortedChanges: 8, sortedCommitteesChanges: 31,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, summaries: 3, summariesChanges: 7})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Nil(t, sovc[0].SortKey)
				require.Nil(t, sovc[1].SortKey)
			},
		},
		{
			heightToReset: 21,
			assert: func() {
				assertOracleVotingLens(t, db, &OracleVotingLens{receipts: 16, contracts: 3, ovContracts: 3, callStarts: 2,
					committees: 12, addresses: 8, sorted: 3, sortedCommittees: 3, sortedChanges: 9, sortedCommitteesChanges: 35,
					callVoteProofs: 4, callVotes: 1, callFinishes: 2, callProlongations: 2, terminations: 1, summaries: 3, summariesChanges: 8})
				sovc, _ = testCommon.GetSortedOracleVotingContracts(db)
				require.Nil(t, sovc[0].SortKey)
				require.Nil(t, sovc[1].SortKey)
			},
		}}

	indexData()
	for i := len(assertions) - 1; i >= 0; i-- {
		require.Nil(t, dbAccessor.ResetTo(assertions[i].heightToReset))
		assertions[i].assert()
	}
}

type assertion struct {
	heightToReset uint64
	assert        func()
}

type OracleVotingLens struct {
	receipts                int
	contracts               int
	ovContracts             int
	callStarts              int
	callVoteProofs          int
	callVotes               int
	callFinishes            int
	callProlongations       int
	terminations            int
	committees              int
	addresses               int
	sorted                  int
	sortedChanges           int
	sortedCommittees        int
	sortedCommitteesChanges int
	summaries               int
	summariesChanges        int
}

func assertOracleVotingLens(t *testing.T, db *sql.DB, lens *OracleVotingLens) {
	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, lens.receipts, len(txReceipts))

	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, lens.contracts, len(contracts))

	ovContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, lens.ovContracts, len(ovContracts))

	callStarts, err := testCommon.GetOracleVotingContractCallStarts(db)
	require.Nil(t, err)
	require.Equal(t, lens.callStarts, len(callStarts))

	callVoteProofs, err := testCommon.GetOracleVotingContractCallVoteProofs(db)
	require.Nil(t, err)
	require.Equal(t, lens.callVoteProofs, len(callVoteProofs))

	callVotes, err := testCommon.GetOracleVotingContractCallVotes(db)
	require.Nil(t, err)
	require.Equal(t, lens.callVotes, len(callVotes))

	callFinishes, err := testCommon.GetOracleVotingContractCallFinishes(db)
	require.Nil(t, err)
	require.Equal(t, lens.callFinishes, len(callFinishes))

	callProlongations, err := testCommon.GetOracleVotingContractCallProlongations(db)
	require.Nil(t, err)
	require.Equal(t, lens.callProlongations, len(callProlongations))

	//committees, err := testCommon.GetOracleVotingContractCommittees(db)
	//require.Nil(t, err)
	//require.Equal(t, lens.committees, len(committees))

	terminations, err := testCommon.GetOracleVotingContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, lens.terminations, len(terminations))

	addresses, err := testCommon.GetAddresses(db)
	require.Nil(t, err)
	require.Equal(t, lens.addresses, len(addresses))

	sorted, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, lens.sorted, len(sorted))

	sortedChanges, err := testCommon.GetSortedOracleVotingContractChanges(db)
	require.Nil(t, err)
	require.Equal(t, lens.sortedChanges, len(sortedChanges))

	sortedCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, lens.sortedCommittees, len(sortedCommittees))

	sortedCommitteesChanges, err := testCommon.GetSortedOracleVotingContractCommitteeChanges(db)
	require.Nil(t, err)
	require.Equal(t, lens.sortedCommitteesChanges, len(sortedCommitteesChanges))

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, lens.summaries, len(summaries))

	summariesChanges, err := testCommon.GetOracleVotingContractSummaryChanges(db)
	require.Nil(t, err)
	require.Equal(t, lens.summariesChanges, len(summariesChanges))
}

func findSortedOracleVotingContractCommittee(list []testCommon.SortedOracleVotingContractCommittee, contractTxId int, address common.Address) *testCommon.SortedOracleVotingContractCommittee {
	for _, item := range list {
		if item.TxId == contractTxId && item.Address == address.Hex() {
			return &item
		}
	}
	return nil
}

func findOracleVotingSummary(list []testCommon.OracleVotingSummary, contractTxId int) *testCommon.OracleVotingSummary {
	for _, item := range list {
		if item.ContractTxId == contractTxId {
			return &item
		}
	}
	return nil
}
