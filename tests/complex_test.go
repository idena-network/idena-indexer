package tests

import (
	"github.com/idena-network/idena-go/blockchain/attachments"
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/types"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_complex(t *testing.T) {
	dbConnector, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	var height uint64
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	appState.Precommit()
	height++
	require.Nil(t, appState.CommitAt(height))
	require.Nil(t, appState.Initialize(height))

	statsCollector := listener.StatsCollector()

	// ----- Epoch 0

	// When
	cid1Str := "bafkreiar6xq6j4ok5pfxaagtec7jwq6fdrdntkrkzpitqenmz4cyj6qswa"
	cidWrongWordsFlipStr := "bafkreifyajvupl2o22zwnkec22xrtwgieovymdl7nz5uf25aqv7lsguova"
	cidFlipToDeleteStr := "bafkreihcvhijrwwts3xl3zufbi2mjng5gltc7ojw2syue7zyritkq3gbii"
	cid1, _ := cid.Parse(cid1Str)
	cidWrongWordsFlip, _ := cid.Parse(cidWrongWordsFlipStr)
	cidFlipToDelete, _ := cid.Parse(cidFlipToDeleteStr)
	flipTx1, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cid1.Bytes(), 0),
	}, key)
	flipTxWrongWords, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cidWrongWordsFlip.Bytes(), 1),
	}, key)
	flipTxToDelete, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cidFlipToDelete.Bytes(), 2),
	}, key)
	statsCollector.EnableCollecting()
	height++
	block := buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{flipTx1, flipTxWrongWords, flipTxToDelete}
	appState.State.AddFlip(addr, cid1.Bytes(), 0)
	appState.State.AddFlip(addr, cidWrongWordsFlip.Bytes(), 1)
	appState.State.AddFlip(addr, cidFlipToDelete.Bytes(), 2)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	// Then
	addressSummaries, _ := testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 3, addressSummaries[0].Flips)
	require.Equal(t, 0, addressSummaries[0].WrongWordsFlips)

	// When
	deleteFlipTx, _ := types2.SignTx(&types2.Transaction{
		Type:    types2.DeleteFlipTx,
		Payload: attachments.CreateDeleteFlipAttachment(cidFlipToDelete.Bytes()),
	}, key)
	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{deleteFlipTx}
	appState.State.DeleteFlip(addr, cidFlipToDelete.Bytes())
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	// Then
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 2, addressSummaries[0].Flips)
	require.Equal(t, 0, addressSummaries[0].WrongWordsFlips)

	// When
	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{
		FlipCids: [][]byte{cid1.Bytes(), cidWrongWordsFlip.Bytes()},
		FlipsPerIdx: map[int]*types.FlipStats{
			0: {
				Status: byte(ceremony.Qualified),
				Answer: types2.Left,
				Grade:  types2.GradeA,
				ShortAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						Grade:      types2.GradeA,
						Point:      1,
					},
				},
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						Grade:      types2.GradeA,
						Point:      1,
					},
				},
			},
			1: {
				Status: byte(ceremony.Qualified),
				Answer: types2.Right,
				Grade:  types2.GradeReported,
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Right,
						Grade:      types2.GradeReported,
						Point:      1,
					},
				},
			},
		},
		IdentitiesPerAddr: map[common.Address]*types.IdentityStats{
			addr: {
				ShortPoint:        1,
				ShortFlips:        1,
				LongPoint:         2,
				LongFlips:         2,
				Approved:          true,
				Missed:            false,
				ShortFlipsToSolve: []int{0},
				LongFlipsToSolve:  []int{0, 1},
			},
		},
	})
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	// Then
	epochIdentities, err := testCommon.GetEpochIdentities(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(epochIdentities))
	require.Equal(t, 3, epochIdentities[0].Id)
	require.Equal(t, 1, epochIdentities[0].ShortAnswers)
	require.Equal(t, 2, epochIdentities[0].LongAnswers)
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 2, addressSummaries[0].Flips)
	require.Equal(t, 1, addressSummaries[0].WrongWordsFlips)

	appState.State.IncEpoch()
	appState.State.ClearFlips(addr)
	// ----- Epoch 1
	// When
	cid1Str = "bafkreiaaogfa4o64mhhvsfczpvjhfrflk7a3kslcmlamyogi2h5yjorto4"
	cidWrongWordsFlipStr = "bafkreibfvtdgosyqev27d2ocklbuwo4bksgisonnq7dnmbceo6kphvgf44"
	cidFlipToDeleteStr = "bafkreig7uxuq7rhv2eoyfdtf3zkb6frdvkid2qljkpzowu5efdtium7hg4"
	cid1, _ = cid.Parse(cid1Str)
	cidWrongWordsFlip, _ = cid.Parse(cidWrongWordsFlipStr)
	cidFlipToDelete, _ = cid.Parse(cidFlipToDeleteStr)
	flipTx1, _ = types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cid1.Bytes(), 0),
	}, key)
	flipTxWrongWords, _ = types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cidWrongWordsFlip.Bytes(), 1),
	}, key)
	flipTxToDelete, _ = types2.SignTx(&types2.Transaction{
		Type:    types2.SubmitFlipTx,
		Payload: attachments.CreateFlipSubmitAttachment(cidFlipToDelete.Bytes(), 2),
	}, key)
	deleteFlipTx, _ = types2.SignTx(&types2.Transaction{
		Type:    types2.DeleteFlipTx,
		Payload: attachments.CreateDeleteFlipAttachment(cidFlipToDelete.Bytes()),
	}, key)
	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{flipTx1}
	appState.State.AddFlip(addr, cid1.Bytes(), 0)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{flipTxToDelete}
	appState.State.AddFlip(addr, cidFlipToDelete.Bytes(), 2)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	statsCollector.EnableCollecting()
	height++
	heightToReset := height
	block = buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{deleteFlipTx}
	appState.State.DeleteFlip(addr, cidFlipToDelete.Bytes())
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	statsCollector.EnableCollecting()
	height++
	block = buildBlock(height)
	block.Body.Transactions = []*types2.Transaction{flipTxWrongWords}
	appState.State.AddFlip(addr, cidWrongWordsFlip.Bytes(), 1)
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	// Then
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 4, addressSummaries[0].Flips)
	require.Equal(t, 1, addressSummaries[0].WrongWordsFlips)

	// When
	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{
		FlipCids: [][]byte{cid1.Bytes(), cidWrongWordsFlip.Bytes()},
		FlipsPerIdx: map[int]*types.FlipStats{
			0: {
				Status: byte(ceremony.Qualified),
				Answer: types2.Left,
				Grade:  types2.GradeA,
				ShortAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						Grade:      types2.GradeA,
						Point:      1,
					},
				},
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						Grade:      types2.GradeA,
						Point:      1,
					},
				},
			},
			1: {
				Status: byte(ceremony.Qualified),
				Answer: types2.Right,
				Grade:  types2.GradeReported,
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Right,
						Grade:      types2.GradeReported,
						Point:      1,
					},
				},
			},
		},
		IdentitiesPerAddr: map[common.Address]*types.IdentityStats{
			addr: {
				ShortPoint:        1,
				ShortFlips:        1,
				LongPoint:         2,
				LongFlips:         2,
				Approved:          true,
				Missed:            false,
				ShortFlipsToSolve: []int{0},
				LongFlipsToSolve:  []int{0, 1},
			},
		},
	})
	height++
	block = buildBlock(height)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	require.Nil(t, applyBlock(bus, block, appState))
	statsCollector.CompleteCollecting()
	// Then
	epochIdentities, err = testCommon.GetEpochIdentities(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 2, len(epochIdentities))
	require.Equal(t, 4, epochIdentities[1].Id)
	require.Equal(t, 1, epochIdentities[1].ShortAnswers)
	require.Equal(t, 2, epochIdentities[1].LongAnswers)
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 4, addressSummaries[0].Flips)
	require.Equal(t, 2, addressSummaries[0].WrongWordsFlips)

	// When
	err = dbAccessor.ResetTo(heightToReset)
	// Then
	require.Nil(t, err)
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 3, addressSummaries[0].Flips)
	require.Equal(t, 1, addressSummaries[0].WrongWordsFlips)
}
