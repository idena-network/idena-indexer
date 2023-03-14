package tests

import (
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/core/stats"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

type tokenContractHolder struct {
	balances map[common.Address]*big.Int
}

func (t *tokenContractHolder) Balance(appState *appstate.AppState, contractAddress common.Address, address []byte) (*big.Int, error) {
	return t.balances[common.BytesToAddress(address)], nil
}

func (t *tokenContractHolder) Info(appState *appstate.AppState, contractAddress common.Address) (stats.Token, error) {
	return stats.Token{}, errors.New("not token")
}

func Test_token_balance(t *testing.T) {
	balances := make(map[common.Address]*big.Int)
	holder := &tokenContractHolder{
		balances: balances,
	}
	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:                   true,
		Schema:                    testCommon.PostgresSchema,
		ScriptsPathPrefix:         "..",
		ChangesHistoryBlocksCount: 10,
		TokenContractHolder:       holder,
	})
	db, listener, bus := ctx.DbConnector, ctx.Listener, ctx.EventBus
	defer listener.Destroy()

	appState := listener.NodeCtx().AppState

	contractAddress := tests.GetRandAddr()
	addr1 := tests.GetRandAddr()
	addr2 := tests.GetRandAddr()
	addr3 := tests.GetRandAddr()
	addr4 := tests.GetRandAddr()

	balances[addr1] = new(big.Int).SetInt64(999999999)
	balances[addr2] = new(big.Int).SetInt64(999)
	balances[addr3] = new(big.Int).SetInt64(9)
	balances[addr4] = new(big.Int).SetInt64(1)

	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	findBalance := func(address common.Address, balances []testCommon.TokenBalance) (testCommon.TokenBalance, bool) {
		for _, v := range balances {
			if common.HexToAddress(v.Address) == address {
				return v, true
			}
		}
		return testCommon.TokenBalance{}, false
	}

	{
		statsCollector.EnableCollecting()
		block := buildBlock(2)

		attachment := attachments.CreateDeployContractAttachment(common.Hash{}, []byte{0x1, 0x2}, []byte{})
		payload, err := attachment.ToBytes()
		require.NoError(t, err)
		tx := &types.Transaction{AccountNonce: 1, Type: types.DeployContractTx, Payload: payload}
		statsCollector.BeginApplyingTx(tx, appState)

		txReceipt := &types.TxReceipt{
			Success:         true,
			TxHash:          tx.Hash(),
			ContractAddress: contractAddress,
			Events: []*types.TxEvent{
				{
					EventName: "Transfer",
					Contract:  contractAddress,
				},
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr1.Bytes(), addr2.Bytes(), new(big.Int).Bytes()},
				},
			},
		}

		statsCollector.AddTxReceipt(txReceipt, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 2, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999999999).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999).Cmp(balance.Balance))

	}

	{
		statsCollector.EnableCollecting()
		block := buildBlock(3)

		attachment := attachments.CreateCallContractAttachment("someMethod")
		payload, err := attachment.ToBytes()
		require.NoError(t, err)
		tx := &types.Transaction{AccountNonce: 2, Type: types.CallContractTx, To: &contractAddress, Payload: payload}
		statsCollector.BeginApplyingTx(tx, appState)

		balances[addr1] = new(big.Int).SetInt64(100)
		balances[addr2] = new(big.Int).SetInt64(10)

		txReceipt := &types.TxReceipt{
			Success: true,
			TxHash:  tx.Hash(),
			Events: []*types.TxEvent{
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr2.Bytes(), addr3.Bytes(), new(big.Int).Bytes()},
				},
				{
					EventName: "_Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr1.Bytes(), addr1.Bytes(), new(big.Int).Bytes()},
				},
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr4.Bytes(), addr4.Bytes(), new(big.Int).Bytes()},
				},
			},
		}

		statsCollector.AddTxReceipt(txReceipt, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 4, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999999999).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(10).Cmp(balance.Balance))

		balance, ok = findBalance(addr3, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(9).Cmp(balance.Balance))

		balance, ok = findBalance(addr4, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(1).Cmp(balance.Balance))
	}

	{
		statsCollector.EnableCollecting()
		block := buildBlock(4)

		attachment := attachments.CreateCallContractAttachment("someMethod")
		payload, err := attachment.ToBytes()
		require.NoError(t, err)
		tx := &types.Transaction{AccountNonce: 3, Type: types.CallContractTx, To: &contractAddress, Payload: payload}
		statsCollector.BeginApplyingTx(tx, appState)

		balances[addr4] = new(big.Int).SetInt64(0)

		txReceipt := &types.TxReceipt{
			Success: true,
			TxHash:  tx.Hash(),
			Events: []*types.TxEvent{
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr1.Bytes(), addr1.Bytes(), new(big.Int).Bytes()},
				},
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr1.Bytes(), addr1.Bytes(), new(big.Int).Bytes()},
				},
				{
					EventName: "Transfer",
					Contract:  contractAddress,
					Data:      [][]byte{addr4.Bytes(), addr4.Bytes(), new(big.Int).Bytes()},
				},
			},
		}

		statsCollector.AddTxReceipt(txReceipt, appState)
		statsCollector.CompleteApplyingTx(appState)
		block.Body.Transactions = append(block.Body.Transactions, tx)

		require.Nil(t, applyBlock(bus, block, appState))
		statsCollector.CompleteCollecting()

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 3, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(100).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(10).Cmp(balance.Balance))

		balance, ok = findBalance(addr3, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(9).Cmp(balance.Balance))
	}

	{
		require.Nil(t, ctx.DbAccessor.ResetTo(4))

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 3, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(100).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(10).Cmp(balance.Balance))

		balance, ok = findBalance(addr3, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(9).Cmp(balance.Balance))
	}

	{
		require.Nil(t, ctx.DbAccessor.ResetTo(3))

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 4, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999999999).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(10).Cmp(balance.Balance))

		balance, ok = findBalance(addr3, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(9).Cmp(balance.Balance))

		balance, ok = findBalance(addr4, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(1).Cmp(balance.Balance))
	}

	{
		require.Nil(t, ctx.DbAccessor.ResetTo(2))

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 1, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 2, len(tokenBalances))

		balance, ok := findBalance(addr1, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999999999).Cmp(balance.Balance))

		balance, ok = findBalance(addr2, tokenBalances)
		require.True(t, ok)
		require.Equal(t, contractAddress, common.HexToAddress(balance.ContractAddress))
		require.Zero(t, new(big.Int).SetInt64(999).Cmp(balance.Balance))
	}

	{
		require.Nil(t, ctx.DbAccessor.ResetTo(1))

		tokens, err := testCommon.Tokens(db)
		require.Nil(t, err)
		require.Equal(t, 0, len(tokens))

		tokenBalances, err := testCommon.TokenBalances(db)
		require.Nil(t, err)
		require.Equal(t, 0, len(tokenBalances))
	}
}
