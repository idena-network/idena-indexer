package tests

import (
	"github.com/idena-network/idena-go/blockchain/attachments"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_tx_receipt(t *testing.T) {
	ctx := testCommon.InitIndexer2(testCommon.Options{
		ClearDb:           true,
		Schema:            testCommon.PostgresSchema,
		ScriptsPathPrefix: "..",
	})
	db, listener, bus := ctx.DbConnector, ctx.Listener, ctx.EventBus
	defer listener.Destroy()

	appState := listener.NodeCtx().AppState

	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()

	statsCollector.EnableCollecting()
	block := buildBlock(2)

	contractAddress := tests.GetRandAddr()
	from := tests.GetRandAddr()

	attachment := attachments.CreateDeployContractAttachment(common.Hash{}, []byte{0x1, 0x2}, []byte{})
	payload, err := attachment.ToBytes()
	require.NoError(t, err)
	tx := &types.Transaction{AccountNonce: 1, Type: types.DeployContractTx, Payload: payload}
	statsCollector.BeginApplyingTx(tx, appState)
	statsCollector.AddWasmContract(contractAddress, []byte{0x1})

	txReceipt := &types.TxReceipt{
		Success:         true,
		TxHash:          tx.Hash(),
		GasUsed:         11,
		GasCost:         big.NewInt(1100),
		ContractAddress: contractAddress,
		Method:          "deploy1",
		ActionResult:    []byte{0x1, 0x3, 0x5, 0x7},
		Error:           errors.New("error message"),
		From:            from,
		Events: []*types.TxEvent{
			{
				EventName: "event1",
			},
			{
				EventName: "event2",
				Data:      [][]byte{{0x1, 0x2}, nil, {}, {0x3, 0x4}},
			},
		},
	}

	statsCollector.AddTxReceipt(txReceipt, appState)
	statsCollector.CompleteApplyingTx(appState)
	block.Body.Transactions = append(block.Body.Transactions, tx)

	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	txReceipts, err := testCommon.GetTxReceipts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(txReceipts))
	require.True(t, txReceipts[0].Success)
	require.Equal(t, 1, txReceipts[0].TxId)
	require.Equal(t, "0.000000000000001100", txReceipts[0].GasCost)
	require.Equal(t, 11, txReceipts[0].GasUsed)
	require.Equal(t, "deploy1", txReceipts[0].Method)
	require.Equal(t, "error message", txReceipts[0].Error)
	require.Equal(t, []byte{0x1, 0x3, 0x5, 0x7}, txReceipts[0].ActionResult)
	require.Equal(t, contractAddress.Hex(), txReceipts[0].ContractAddress)
	require.Equal(t, from.Hex(), txReceipts[0].From)

	txEvents, err := testCommon.GetTxEvents(db)
	require.Nil(t, err)
	require.Equal(t, 2, len(txEvents))
	require.Equal(t, "event1", txEvents[0].EventName)
	require.Equal(t, 0, txEvents[0].Index)
	require.Equal(t, 1, txEvents[0].TxId)
	require.Nil(t, txEvents[0].Data)
	require.Equal(t, "event2", txEvents[1].EventName)
	require.Equal(t, 1, txEvents[1].Index)
	require.Equal(t, 1, txEvents[1].TxId)
	require.Equal(t, [][]byte{{0x1, 0x2}, nil, nil, {0x3, 0x4}}, txEvents[1].Data)

	contracts, err := testCommon.GetContracts(db)
	require.Nil(t, err)
	require.Equal(t, 1, len(contracts))
	require.Equal(t, 6, contracts[0].Type)
	require.Zero(t, contracts[0].Stake.Sign())
	require.Equal(t, contractAddress.Hex(), contracts[0].Address)
	require.Equal(t, 1, contracts[0].TxId)
}
