package tests

import (
	"github.com/idena-network/idena-go/blockchain/attachments"
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_reportedFlipRewards(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	// Given
	var height uint64
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)

	reporter1 := tests.GetRandAddr()
	appState.State.SetState(reporter1, state.Verified)
	reporter2 := tests.GetRandAddr()
	appState.State.SetState(reporter2, state.Verified)

	appState.Precommit()
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	cid1Str := "bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa"
	cid2Str := "bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova"
	cid1, _ := cid.Parse(cid1Str)
	cid2, _ := cid.Parse(cid2Str)
	flipTx1, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cid1.Bytes(), 0),
	}, key)
	flipTx2, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cid2.Bytes(), 1),
	}, key)
	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{flipTx1, flipTx2}
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// When
	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{
		FlipCids: [][]byte{cid1.Bytes(), cid2.Bytes()},
	})
	statsCollector.SetValidationResults(&types2.ValidationResults{})
	statsCollector.AddReportedFlipsReward(reporter1, 0, new(big.Int).SetInt64(int64(1000_000)), new(big.Int).SetInt64(int64(100_000)))
	statsCollector.AddReportedFlipsReward(reporter1, 1, new(big.Int).SetInt64(int64(300_000)), new(big.Int).SetInt64(int64(200_000)))
	statsCollector.AddReportedFlipsReward(reporter2, 1, new(big.Int).SetInt64(int64(200_000)), new(big.Int).SetInt64(int64(300_000)))
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()

	// Then
	validationRewards, _ := testCommon.GetValidationRewards(dbConnector)
	require.Equal(t, 2, len(validationRewards))
	for _, r := range validationRewards {
		switch {
		case reporter1.Hex() == r.Address:
			require.Equal(t, 9, r.Type)
			require.Equal(t, 0, r.Epoch)
			require.Zero(t, decimal.RequireFromString("0.0000000000013").Cmp(r.Balance))
			require.Zero(t, decimal.RequireFromString("0.0000000000003").Cmp(r.Stake))
			break
		case reporter2.Hex() == r.Address:
			require.Equal(t, 9, r.Type)
			require.Equal(t, 0, r.Epoch)
			require.Zero(t, decimal.RequireFromString("0.0000000000002").Cmp(r.Balance))
			require.Zero(t, decimal.RequireFromString("0.0000000000003").Cmp(r.Stake))
			break
		default:
			panic("Unexpected reward")
		}
	}

	reportedFlipRewards, _ := testCommon.GetReportedFlipRewards(dbConnector)
	require.Equal(t, 3, len(reportedFlipRewards))
	for _, r := range reportedFlipRewards {
		switch {
		case reporter1.Hex() == r.Address && r.FlipTxId == 1:
			require.Equal(t, 0, r.Epoch)
			require.Zero(t, decimal.RequireFromString("0.000000000001").Cmp(r.Balance))
			require.Zero(t, decimal.RequireFromString("0.0000000000001").Cmp(r.Stake))
			break
		case reporter1.Hex() == r.Address && r.FlipTxId == 2:
			require.Equal(t, 0, r.Epoch)
			require.Zero(t, decimal.RequireFromString("0.0000000000003").Cmp(r.Balance))
			require.Zero(t, decimal.RequireFromString("0.0000000000002").Cmp(r.Stake))
			break
		case reporter2.Hex() == r.Address && r.FlipTxId == 2:
			require.Equal(t, 0, r.Epoch)
			require.Zero(t, decimal.RequireFromString("0.0000000000002").Cmp(r.Balance))
			require.Zero(t, decimal.RequireFromString("0.0000000000003").Cmp(r.Stake))
			break
		default:
			panic("Unexpected reported flip reward")
		}
	}
}
