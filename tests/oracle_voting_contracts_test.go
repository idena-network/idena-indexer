package tests

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/eventbus"
	"github.com/idena-network/idena-go/core/state"
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

func Test_OracleVotingContractDeploy(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(contracts))
	require.Equal(t, contractAddress1.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
	require.Equal(t, 2, contracts[0].Type)
	require.Equal(t, "0.0000000000000123", contracts[0].Stake.String())

	require.Equal(t, contractAddress2.Hex(), contracts[1].Address)
	require.Equal(t, 3, contracts[1].TxId)

	feContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(feContracts))
	require.Equal(t, 1, feContracts[0].TxId)
	require.Equal(t, startTime.UTC().Unix(), feContracts[0].StartTime)
	require.Equal(t, 1, feContracts[0].VotingDuration)
	require.Equal(t, "0.0000000000000234", feContracts[0].VotingMinPayment.String())
	require.Equal(t, []byte{0x1, 0x2}, feContracts[0].Fact)
	require.Equal(t, 2, feContracts[0].PublicVotingDuration)
	require.Equal(t, 3, feContracts[0].WinnerThreshold)
	require.Equal(t, 4, feContracts[0].Quorum)
	require.Equal(t, 5, feContracts[0].CommitteeSize)
	require.Equal(t, 7, feContracts[0].OwnerFee)

	require.Equal(t, 3, feContracts[1].TxId)
	require.Nil(t, feContracts[1].VotingMinPayment)
	require.Nil(t, feContracts[1].Fact)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, 1, sortedFeContracts[0].TxId)
	require.Equal(t, 0, sortedFeContracts[0].State)
	require.Equal(t, common.Address{}.Hex(), sortedFeContracts[0].Author)
	require.Nil(t, sortedFeContracts[0].CountingBlock)
	require.Nil(t, sortedFeContracts[0].Epoch)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)
	require.Equal(t, common.Address{}.Hex(), sortedFeContracts[1].Author)

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sortedFeContractCommittees))

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].ContractTxId)
	require.Equal(t, 0, summaries[0].VoteProofs)
	require.Equal(t, 0, summaries[0].Votes)
	require.Equal(t, "0.0000000000000123", summaries[0].Stake.String())
	require.Nil(t, summaries[0].TotalReward)
	require.Equal(t, 3, summaries[1].ContractTxId)
	require.Equal(t, 0, summaries[1].VoteProofs)
	require.Equal(t, 0, summaries[1].Votes)
	require.Equal(t, "0.0000000000000123", summaries[1].Stake.String())
}

func deployOracleVotingContracts(t *testing.T, listener incoming.Listener, bus eventbus.Bus, startTime time.Time, addr1, addr2 common.Address) {
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
	tx := &types.Transaction{AccountNonce: 1, Type: types.DeployContract}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingDeploy(addr1, uint64(startTime.Unix()), new(big.Int).SetUint64(23400), []byte{0x1, 0x2},
		0, 1, 2, 3, 4, 5, 7)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 11, GasCost: big.NewInt(1100), ContractAddress: addr1, Method: "deploy1"}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Failed tx receipt
	failedContractAddress := tests.GetRandAddr()
	tx = &types.Transaction{AccountNonce: 2, Type: types.DeployContract}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingDeploy(failedContractAddress, uint64(startTime.Unix()), new(big.Int).SetUint64(23400), []byte{0x1, 0x2},
		0, 1, 2, 3, 4, 5, 7)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash(), Error: errors.New("error message"), Method: "deploy2"}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Success, fact and voting min payment are nil
	tx = &types.Transaction{AccountNonce: 3, Type: types.DeployContract}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingDeploy(addr2, uint64(startTime.Unix()), nil, []byte{},
		0, 11, 12, 13, 14, 15, 17)
	statsCollector.AddContractStake(new(big.Int).SetUint64(12300))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), GasUsed: 33, GasCost: big.NewInt(3300), ContractAddress: addr2, Method: "deploy3"}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
}

func Test_TxReceipts(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(txReceipts))

	require.True(t, txReceipts[0].Success)
	require.Equal(t, 1, txReceipts[0].TxId)
	require.Equal(t, "0.000000000000001100", txReceipts[0].GasCost)
	require.Equal(t, 11, txReceipts[0].GasUsed)
	require.Equal(t, "deploy1", txReceipts[0].Method)
	require.Empty(t, txReceipts[0].Error)

	require.False(t, txReceipts[1].Success)
	require.Equal(t, 2, txReceipts[1].TxId)
	require.Equal(t, "0.000000000000000000", txReceipts[1].GasCost)
	require.Equal(t, 0, txReceipts[1].GasUsed)
	require.Equal(t, "deploy2", txReceipts[1].Method)
	require.Equal(t, "error message", txReceipts[1].Error)

	require.True(t, txReceipts[2].Success)
	require.Equal(t, 3, txReceipts[2].TxId)
	require.Equal(t, "0.000000000000003300", txReceipts[2].GasCost)
	require.Equal(t, 33, txReceipts[2].GasUsed)
	require.Equal(t, "deploy3", txReceipts[2].Method)
	require.Empty(t, txReceipts[2].Error)
}

func Test_OracleVotingContractCallStart(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetPubKey(addr1, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress1}
	statsCollector.BeginApplyingTx(tx, appState)
	// empty committee
	statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 45, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)

	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.AddBalance(contractAddress1, big.NewInt(54321000))
	statsCollector.CompleteBalanceUpdate(appState)

	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	tx = &types.Transaction{AccountNonce: 5, To: &contractAddress2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, 234, 2, new(big.Int).SetUint64(100), []byte{0x3, 0x4}, 50, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(contracts))

	feContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(feContracts))

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 1, sortedFeContracts[0].State)
	require.Equal(t, 124, *sortedFeContracts[0].CountingBlock)
	require.Equal(t, 2, *sortedFeContracts[0].Epoch)
	require.Equal(t, "000000000000000000000000000000.0000000000101053440000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, 1, sortedFeContracts[1].State)
	require.Equal(t, 245, *sortedFeContracts[1].CountingBlock)

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(sortedFeContractCommittees))
	require.Equal(t, addr1.Hex(), sortedFeContractCommittees[0].Address)
	require.Equal(t, 1, sortedFeContractCommittees[0].State)
	require.Equal(t, 1, sortedFeContractCommittees[0].TxId)
	require.Equal(t, "000000000000000000000000000000.0000000000101053440000000000000000001", *sortedFeContractCommittees[0].SortKey)
	require.False(t, sortedFeContractCommittees[0].Voted)
	require.Equal(t, common.Address{}.Hex(), sortedFeContractCommittees[0].Author)

	require.Equal(t, 1, sortedFeContractCommittees[1].State)
	require.Equal(t, 3, sortedFeContractCommittees[1].TxId)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[1].SortKey)
	require.False(t, sortedFeContractCommittees[1].Voted)
	require.Equal(t, 1, sortedFeContractCommittees[2].State)
	require.Equal(t, 3, sortedFeContractCommittees[2].TxId)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[2].SortKey)
	require.False(t, sortedFeContractCommittees[2].Voted)
	require.Contains(t, []string{addr1.Hex(), addr2.Hex()}, sortedFeContractCommittees[1].Address)
	require.Contains(t, []string{addr1.Hex(), addr2.Hex()}, sortedFeContractCommittees[2].Address)
	require.Equal(t, common.Address{}.Hex(), sortedFeContractCommittees[1].Author)
	require.Equal(t, common.Address{}.Hex(), sortedFeContractCommittees[2].Author)

	starts, err := testCommon.GetOracleVotingContractCallStarts(db)
	require.Equal(t, 2, len(starts))
	require.Nil(t, starts[0].VotingMinPayment)
	require.Equal(t, []byte{0x2, 0x3}, starts[0].VrfSeed)
	require.Equal(t, 123, starts[0].StartBlockHeight)
	require.Equal(t, 2, starts[0].Epoch)
	require.Equal(t, "0.0000000000000001", starts[1].VotingMinPayment.String())
	require.Equal(t, []byte{0x3, 0x4}, starts[1].VrfSeed)
	require.Equal(t, 234, starts[1].StartBlockHeight)
	require.Equal(t, 3, starts[1].ContractTxId)

	balanceUpdates, err := testCommon.GetBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(balanceUpdates))
}

func Test_OracleVotingContractCallStartEmptyCommittee(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetPubKey(addr1, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress1}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 70, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)

	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	// Then
	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sortedFeContractCommittees))
}

func Test_OracleVotingContractCallStartFail(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetPubKey(addr1, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	tx := &types.Transaction{AccountNonce: 4}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 70, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: false, TxHash: tx.Hash()}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(contracts))

	feContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(feContracts))

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 0, sortedFeContracts[0].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, 0, sortedFeContracts[1].State)

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sortedFeContractCommittees))

	starts, err := testCommon.GetOracleVotingContractCallStarts(db)
	require.Equal(t, 0, len(starts))

	balanceUpdates, err := testCommon.GetBalanceUpdates(db)
	require.Nil(t, err)
	require.Empty(t, balanceUpdates)
}

func Test_OracleVotingContractCallVoteProof(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	respondentKey, _ := crypto.GenerateKey()
	respondentAddress := crypto.PubkeyToAddress(respondentKey.PublicKey)
	addr2 := tests.GetRandAddr()
	appState.State.SetState(respondentAddress, state.Verified)
	appState.State.SetPubKey(respondentAddress, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	statsCollector := listener.StatsCollector()
	// Start voting
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 50, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.AddBalance(contractAddress, big.NewInt(54321000))
	statsCollector.CompleteBalanceUpdate(appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Send vote proof
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteProof([]byte{0x3, 0x4})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(respondentAddress, big.NewInt(2000))
	appState.State.AddBalance(contractAddress, big.NewInt(3000))
	statsCollector.CompleteBalanceUpdate(appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	feContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(feContracts))

	voteProofs, err := testCommon.GetOracleVotingContractCallVoteProofs(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(voteProofs))
	require.Equal(t, 1, voteProofs[0].ContractTxId)
	require.Equal(t, respondentAddress.Hex(), voteProofs[0].Address)
	require.Equal(t, 5, voteProofs[0].TxId)
	require.Equal(t, []byte{0x3, 0x4}, voteProofs[0].VoteHash)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 1, sortedFeContracts[0].State)
	require.Equal(t, "000000000000000000000000000000.0000000000101059020000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, 0, sortedFeContracts[1].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContractCommittees))

	for _, committeeMember := range sortedFeContractCommittees {
		require.Contains(t, []string{respondentAddress.Hex(), addr2.Hex()}, committeeMember.Address)
		if respondentAddress.Hex() == committeeMember.Address {
			require.Equal(t, 5, committeeMember.State)
			require.Nil(t, committeeMember.SortKey)
		} else {
			require.Equal(t, 1, committeeMember.State)
			require.Equal(t, "000000000000000000000000000000.0000000000101059020000000000000000001", *committeeMember.SortKey)
		}
		require.Equal(t, 1, committeeMember.TxId)
		require.Equal(t, respondentAddress.Hex() == committeeMember.Address, committeeMember.Voted)
	}

	balanceUpdates, err := testCommon.GetBalanceUpdates(db)
	require.Nil(t, err)
	require.Equal(t, 3, len(balanceUpdates))

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].ContractTxId)
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 0, summaries[0].Votes)
	require.Equal(t, 3, summaries[1].ContractTxId)
	require.Equal(t, 0, summaries[1].VoteProofs)
	require.Equal(t, 0, summaries[1].Votes)
}

func Test_OracleVotingContractCallVoteProofTwice(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	respondentKey, _ := crypto.GenerateKey()
	respondentAddress := crypto.PubkeyToAddress(respondentKey.PublicKey)
	addr2 := tests.GetRandAddr()
	appState.State.SetState(respondentAddress, state.Verified)
	appState.State.SetPubKey(respondentAddress, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	statsCollector := listener.StatsCollector()
	// Start voting
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, 123, 2, nil, []byte{0x2, 0x3}, 50, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.AddBalance(contractAddress, big.NewInt(54321000))
	statsCollector.CompleteBalanceUpdate(appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Send vote proof
	statsCollector.EnableCollecting()
	height = uint64(4)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteProof([]byte{0x3, 0x4})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Send vote proof again
	statsCollector.EnableCollecting()
	height = uint64(5)
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 6, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVoteProof([]byte{0x4, 0x5})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	voteProofs, err := testCommon.GetOracleVotingContractCallVoteProofs(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(voteProofs))

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].ContractTxId)
	require.Equal(t, 1, summaries[0].VoteProofs)
	require.Equal(t, 0, summaries[0].Votes)
	require.Equal(t, 3, summaries[1].ContractTxId)
	require.Equal(t, 0, summaries[1].VoteProofs)
	require.Equal(t, 0, summaries[1].Votes)
}

func Test_OracleVotingContractCallVote(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Send vote
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallVote(7, []byte{0x4, 0x5})
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	votes, err := testCommon.GetOracleVotingContractCallVotes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(votes))
	require.Equal(t, 1, votes[0].ContractTxId)
	require.Equal(t, 4, votes[0].TxId)
	require.Equal(t, byte(7), votes[0].Vote)
	require.Equal(t, []byte{0x4, 0x5}, votes[0].Salt)

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, 1, summaries[0].ContractTxId)
	require.Equal(t, 0, summaries[0].VoteProofs)
	require.Equal(t, 1, summaries[0].Votes)
	require.Equal(t, 3, summaries[1].ContractTxId)
	require.Equal(t, 0, summaries[1].VoteProofs)
	require.Equal(t, 0, summaries[1].Votes)

	results, err := testCommon.GetOracleVotingResults(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, 1, results[0].ContractTxId)
	require.Equal(t, 7, results[0].Option)
	require.Equal(t, 1, results[0].Count)
}

func Test_OracleVotingContractCallFinish(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	startTime := time.Now().UTC()
	contractAddress1, contractAddress2 := tests.GetRandAddr(), tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress1, contractAddress2)

	statsCollector := listener.StatsCollector()
	appState := listener.NodeCtx().AppState

	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)

	tx := &types.Transaction{AccountNonce: 4, To: &contractAddress1}
	statsCollector.BeginApplyingTx(tx, appState)
	result := byte(5)
	statsCollector.AddOracleVotingCallFinish(33, &result, big.NewInt(500), big.NewInt(700), big.NewInt(800))
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress1}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	tx = &types.Transaction{AccountNonce: 5, To: &contractAddress2}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallFinish(34, nil, nil, nil, nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress2}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(contracts))

	feContracts, err := testCommon.GetOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(feContracts))

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 2, sortedFeContracts[0].State)
	require.Equal(t, 2, sortedFeContracts[1].State)

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, "-0.0000000000000003", summaries[0].TotalReward.String())
	require.Equal(t, "0.0000000000000131", summaries[0].Stake.String())
	require.Equal(t, "0", summaries[1].TotalReward.String())
	require.Equal(t, "0.0000000000000123", summaries[1].Stake.String())

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 0, len(sortedFeContractCommittees))

	finishes, err := testCommon.GetOracleVotingContractCallFinishes(db)
	require.Equal(t, 2, len(finishes))
	require.Equal(t, 4, finishes[0].TxId)
	require.Equal(t, 1, finishes[0].ContractTxId)
	require.Equal(t, 33, finishes[0].State)
	require.Equal(t, byte(5), *finishes[0].Result)
	require.Equal(t, "0.0000000000000005", finishes[0].Fund.String())
	require.Equal(t, "0.0000000000000007", finishes[0].OracleReward.String())
	require.Equal(t, "0.0000000000000008", finishes[0].OwnerReward.String())
	require.Equal(t, 5, finishes[1].TxId)
	require.Equal(t, 3, finishes[1].ContractTxId)
	require.Equal(t, 34, finishes[1].State)
	require.Nil(t, finishes[1].Result)
	require.Equal(t, "0", finishes[1].Fund.String())
	require.Equal(t, "0", finishes[1].OracleReward.String())
	require.Equal(t, "0", finishes[1].OwnerReward.String())
}

func Test_OracleVotingContractCallProlongation(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Send prolongation
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	startBlock := uint64(7)
	statsCollector.AddOracleVotingCallProlongation(&startBlock, 7890, []byte{0x4, 0x5}, 999, 999)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	prolongations, err := testCommon.GetOracleVotingContractCallProlongations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(prolongations))
	require.Equal(t, 1, prolongations[0].ContractTxId)
	require.Equal(t, 4, prolongations[0].TxId)
	require.Equal(t, 7, *prolongations[0].StartBlock)
	require.Equal(t, 7890, prolongations[0].Epoch)
	require.Equal(t, []byte{0x4, 0x5}, prolongations[0].VrfSeed)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 7890, *sortedFeContracts[0].Epoch)
}

func Test_OracleVotingContractCallProlongationWithoutStartBlock(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Send prolongation
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallProlongation(nil, 7890, []byte{0x4, 0x5}, 999, 999)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	prolongations, err := testCommon.GetOracleVotingContractCallProlongations(db)
	require.Nil(t, err)
	require.Nil(t, prolongations[0].StartBlock)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 7890, *sortedFeContracts[0].Epoch)
}

func Test_OracleVotingContractCallAddStake(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	summaries, err := testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, "0.0000000000000123", summaries[0].Stake.String())
	require.Equal(t, "0.0000000000000123", summaries[1].Stake.String())

	// Add stake
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress, Amount: new(big.Int).SetUint64(23400)}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallAddStake()
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	votes, err := testCommon.GetOracleVotingContractCallAddStakes(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(votes))
	require.Equal(t, 1, votes[0].ContractTxId)
	require.Equal(t, 4, votes[0].TxId)

	summaries, err = testCommon.GetOracleVotingSummaries(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(summaries))
	require.Equal(t, "0.0000000000000357", summaries[0].Stake.String())
	require.Equal(t, "0.0000000000000123", summaries[1].Stake.String())
}

func Test_OracleVotingContractTermination(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	// Send prolongation
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingTermination(big.NewInt(7800), big.NewInt(7900), nil)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	terminations, err := testCommon.GetOracleVotingContractTerminations(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(terminations))
	require.Equal(t, 1, terminations[0].ContractTxId)
	require.Equal(t, 4, terminations[0].TxId)
	require.Equal(t, "0.0000000000000078", terminations[0].Fund.String())
	require.Equal(t, "0.0000000000000079", terminations[0].OracleReward.String())
	require.Nil(t, terminations[0].OwnerReward)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 4, sortedFeContracts[0].State)
}

func Test_OracleVotingContractUpdateBalance(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	appState := listener.NodeCtx().AppState
	statsCollector := listener.StatsCollector()

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, contractAddress, tests.GetRandAddr())

	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.BeginTxBalanceUpdate(tx, appState)
	appState.State.SetBalance(contractAddress, big.NewInt(12300))
	statsCollector.CompleteBalanceUpdate(appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, "000000000000000000000000000000.0000000000000039260000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, 1, sortedFeContracts[0].TxId)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)
}

func Test_OracleVotingContractSetNewCommitteeAndSwitchToCountingState(t *testing.T) {
	db, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	appState := listener.NodeCtx().AppState
	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	appState.State.SetState(addr1, state.Verified)
	appState.State.SetPubKey(addr1, []byte{0x1, 0x2})
	appState.State.SetState(addr2, state.Verified)
	appState.State.SetPubKey(addr2, []byte{0x2, 0x3})

	// Deploy contract
	startTime := time.Now().UTC()
	contractAddress := tests.GetRandAddr()
	deployOracleVotingContracts(t, listener, bus, startTime, tests.GetRandAddr(), contractAddress)

	sortedFeContracts, err := testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)

	sortedFeContractCommittees, err := testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Empty(t, sortedFeContractCommittees)

	statsCollector := listener.StatsCollector()

	// Start voting
	respondentKey, _ := crypto.GenerateKey()
	statsCollector.EnableCollecting()
	height := uint64(3)
	block := buildBlock(height)
	tx, _ := types.SignTx(&types.Transaction{AccountNonce: 4, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallStart(1, height, 2, nil, []byte{0x3, 0x4}, 50, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	sortedFeContracts, err = testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, "000000000000000000000000000000.0000000000000016380000000000000000001", *sortedFeContracts[0].SortKey)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)

	sortedFeContractCommittees, err = testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContractCommittees))
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[0].SortKey)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[1].SortKey)

	for height = 4; height < 14; height++ {
		block := buildBlock(height)
		statsCollector.EnableCollecting()
		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()
	}

	// Then
	sortedFeContracts, err = testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 1, sortedFeContracts[1].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)

	sortedFeContractCommittees, err = testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContractCommittees))
	require.Equal(t, 1, sortedFeContractCommittees[0].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[0].SortKey)
	require.Equal(t, 1, sortedFeContractCommittees[1].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[1].SortKey)

	// Apply block with height = start_block + voting_duration
	block = buildBlock(height)
	statsCollector.EnableCollecting()
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	sortedFeContracts, err = testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 3, sortedFeContracts[1].State)
	require.Nil(t, sortedFeContracts[1].SortKey)

	sortedFeContractCommittees, err = testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContractCommittees))
	require.Equal(t, 3, sortedFeContractCommittees[0].State)
	require.Nil(t, sortedFeContractCommittees[0].SortKey)
	require.Equal(t, 3, sortedFeContractCommittees[1].State)
	require.Nil(t, sortedFeContractCommittees[1].SortKey)

	// Prolong voting
	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	tx, _ = types.SignTx(&types.Transaction{AccountNonce: 5, To: &contractAddress}, respondentKey)
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddOracleVotingCallProlongation(&height, 1000, []byte{0x5, 0x6}, 50, 100)
	statsCollector.AddTxReceipt(&types.TxReceipt{Success: true, TxHash: tx.Hash(), ContractAddress: contractAddress}, appState)

	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	sortedFeContracts, err = testCommon.GetSortedOracleVotingContracts(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(sortedFeContracts))
	require.Equal(t, 1, sortedFeContracts[1].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContracts[1].SortKey)
	require.Equal(t, int(height)+11, *sortedFeContracts[1].CountingBlock)

	sortedFeContractCommittees, err = testCommon.GetSortedOracleVotingContractCommittees(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(sortedFeContractCommittees))
	require.Equal(t, 1, sortedFeContractCommittees[0].State)
	require.Equal(t, "000000000000000000000000000000.0000000000000000000000000000000000003", *sortedFeContractCommittees[0].SortKey)
}
