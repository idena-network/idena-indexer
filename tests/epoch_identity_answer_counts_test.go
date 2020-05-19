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

func Test_epochIdentityAnswerCounts(t *testing.T) {
	dbConnector, _, listener, _, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")

	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	appState := listener.NodeCtx().AppState
	appState.State.SetState(addr, state.Verified)
	appState.Precommit()
	require.Nil(t, appState.CommitAt(1))
	require.Nil(t, appState.Initialize(1))

	statsCollector := listener.StatsCollector()
	cid1Str := "bafkreigzhxfx4kgbarbezmqg5kah5cqt5nv3kbyzvi3oavc4owiaqdwynu"
	cid2Str := "bafkreicf32vx755vf2gj5iygf46wsvxll45qtu6hhkbque7kjadd2cbvuu"
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
	block := buildBlock(2)
	block.Body.Transactions = []*types2.Transaction{flipTx1, flipTx2}
	applyBlock(bus, block)
	statsCollector.CompleteCollecting()

	// When
	statsCollector.EnableCollecting()
	statsCollector.SetValidation(&types.ValidationStats{
		FlipCids: [][]byte{cid1.Bytes(), cid2.Bytes()},
		FlipsPerIdx: map[int]*types.FlipStats{
			0: {
				Status:     byte(ceremony.Qualified),
				Answer:     types2.Left,
				WrongWords: false,
				ShortAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						WrongWords: false,
						Point:      1,
					},
				},
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Left,
						WrongWords: false,
						Point:      1,
					},
				},
			},
			1: {
				Status:     byte(ceremony.Qualified),
				Answer:     types2.Right,
				WrongWords: false,
				LongAnswers: []types.FlipAnswerStats{
					{
						Respondent: addr,
						Answer:     types2.Right,
						WrongWords: false,
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
	block = buildBlock(3)
	block.Header.ProposedHeader.Flags = types2.ValidationFinished
	applyBlock(bus, block)
	statsCollector.CompleteCollecting()

	// Then
	epochIdentities, err := testCommon.GetEpochIdentities(dbConnector)
	require.Nil(t, err)
	require.Equal(t, 1, len(epochIdentities))
	require.Equal(t, 3, epochIdentities[0].Id)
	require.Equal(t, 1, epochIdentities[0].ShortAnswers)
	require.Equal(t, 2, epochIdentities[0].LongAnswers)
}
