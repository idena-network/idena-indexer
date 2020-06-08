package tests

import (
	"encoding/hex"
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/crypto"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_transactionRaw(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	var height uint64
	key, _ := crypto.GenerateKey()
	appState := listener.NodeCtx().AppState
	appState.Precommit()
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	// When
	tx, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SendTx,
		Amount:  big.NewInt(1),
		Payload: []byte{0, 1, 2},
	}, key)
	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{tx}
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	b, _ := tx.ToBytes()
	txRaws, err := testCommon.GetTransactionRaws(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(txRaws))
	require.Equal(t, hex.EncodeToString(b), hex.EncodeToString(txRaws[0].Raw))
}
