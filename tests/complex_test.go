package tests

import (
	"github.com/idena-network/idena-go/blockchain/attachments"
	types2 "github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/ceremony"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/crypto"
	"github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-go/tests"
	"github.com/idena-network/idena-indexer/db"
	testCommon "github.com/idena-network/idena-indexer/tests/common"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func Test_complex(t *testing.T) {
	dbConnector, _, listener, dbAccessor, bus := testCommon.InitIndexer(true, 0, testCommon.PostgresSchema, "..")
	defer listener.Destroy()

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
		Shards: map[common.ShardId]*types.ValidationShardStats{
			1: {
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
								Index:      2,
								Considered: true,
							},
						},
						LongAnswers: []types.FlipAnswerStats{
							{
								Respondent: addr,
								Answer:     types2.Left,
								Grade:      types2.GradeA,
								Point:      1,
								Index:      2,
								Considered: true,
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
								Index:      2,
								Considered: true,
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
			},
		},
	})
	appState.State.IncEpoch()
	appState.State.ClearFlips(addr)
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
		Shards: map[common.ShardId]*types.ValidationShardStats{
			1: {
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
								Index:      2,
								Considered: true,
							},
						},
						LongAnswers: []types.FlipAnswerStats{
							{
								Respondent: addr,
								Answer:     types2.Left,
								Grade:      types2.GradeA,
								Point:      1,
								Index:      2,
								Considered: true,
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
								Index:      2,
								Considered: true,
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
			},
		},
	})

	delegatee := tests.GetRandAddr()
	statsCollector.AddValidationReward(delegatee, addr, 100, new(big.Int).SetUint64(200), new(big.Int).SetUint64(100))
	statsCollector.AddFlipsReward(delegatee, addr, new(big.Int).SetUint64(400), new(big.Int).SetUint64(300), nil)
	rewardedInvitationTxHash := deleteFlipTx.Hash()
	statsCollector.AddInvitationsReward(delegatee, addr, new(big.Int).SetUint64(600), new(big.Int).SetUint64(500), 2, &rewardedInvitationTxHash, 99, false)

	statsCollector.BeginEpochRewardBalanceUpdate(delegatee, addr, appState)
	appState.State.AddBalance(delegatee, new(big.Int).SetUint64(1))
	appState.State.AddStake(addr, new(big.Int).SetUint64(5))
	statsCollector.CompleteBalanceUpdate(appState)

	statsCollector.BeginEpochRewardBalanceUpdate(delegatee, addr, appState)
	appState.State.AddBalance(delegatee, new(big.Int).SetUint64(10))
	appState.State.AddStake(addr, new(big.Int).SetUint64(50))
	statsCollector.CompleteBalanceUpdate(appState)

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

	rewardBounds, err := testCommon.GetRewardBounds(dbConnector)
	require.Nil(t, err)
	require.Len(t, rewardBounds, 1)
	require.Equal(t, uint64(1), rewardBounds[0].Epoch)
	require.Equal(t, addr.Hex(), rewardBounds[0].MinAddress)
	require.Equal(t, "0.0000000000000021", rewardBounds[0].MinAmount.String())
	require.Equal(t, addr.Hex(), rewardBounds[0].MaxAddress)
	require.Equal(t, "0.0000000000000021", rewardBounds[0].MaxAmount.String())

	rewardedInvitations, err := testCommon.GetRewardedInvitations(dbConnector)
	require.Nil(t, err)
	require.Len(t, rewardedInvitations, 1)
	require.Equal(t, 99, *rewardedInvitations[0].EpochHeight)
	require.Equal(t, 7, rewardedInvitations[0].InviteTxId)

	balanceUpdates, err := testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Len(t, balanceUpdates, 3)

	require.Equal(t, addr.Hex(), balanceUpdates[0].Address)
	require.Equal(t, "0", balanceUpdates[0].BalanceOld.String())
	require.Equal(t, "0", balanceUpdates[0].StakeOld.String())
	require.Nil(t, balanceUpdates[0].PenaltyOld)
	require.Equal(t, "0.000000000000000011", balanceUpdates[0].BalanceNew.String())
	require.Equal(t, "0.000000000000000055", balanceUpdates[0].StakeNew.String())
	require.Nil(t, balanceUpdates[0].PenaltyNew)
	require.Equal(t, db.EpochRewardReason, balanceUpdates[0].Reason)

	require.Equal(t, addr.Hex(), balanceUpdates[1].Address)
	require.Equal(t, "0.000000000000000011", balanceUpdates[1].BalanceOld.String())
	require.Equal(t, "0.000000000000000055", balanceUpdates[1].StakeOld.String())
	require.Nil(t, balanceUpdates[1].PenaltyOld)
	require.Equal(t, "0", balanceUpdates[1].BalanceNew.String())
	require.Equal(t, "0.000000000000000055", balanceUpdates[1].StakeNew.String())
	require.Nil(t, balanceUpdates[1].PenaltyNew)
	require.Equal(t, db.DelegatorEpochRewardReason, balanceUpdates[1].Reason)

	require.Equal(t, delegatee.Hex(), balanceUpdates[2].Address)
	require.Equal(t, "0", balanceUpdates[2].BalanceOld.String())
	require.Equal(t, "0", balanceUpdates[2].StakeOld.String())
	require.Nil(t, balanceUpdates[2].PenaltyOld)
	require.Equal(t, "0.000000000000000011", balanceUpdates[2].BalanceNew.String())
	require.Equal(t, "0", balanceUpdates[2].StakeNew.String())
	require.Nil(t, balanceUpdates[2].PenaltyNew)
	require.Equal(t, db.DelegateeEpochRewardReason, balanceUpdates[2].Reason)

	delegateeTotalValidationRewards, err := testCommon.GetDelegateeTotalValidationRewards(dbConnector)
	require.Nil(t, err)
	require.Len(t, delegateeTotalValidationRewards, 1)
	require.Equal(t, 1, delegateeTotalValidationRewards[0].Epoch)
	require.Equal(t, delegatee.Hex(), delegateeTotalValidationRewards[0].DelegateeAddress)
	require.Equal(t, "0.0000000000000012", delegateeTotalValidationRewards[0].TotalBalance.String())
	require.Equal(t, "0.0000000000000002", delegateeTotalValidationRewards[0].ValidationBalance.String())
	require.Equal(t, "0.0000000000000004", delegateeTotalValidationRewards[0].FlipsBalance.String())
	require.Equal(t, "0.0000000000000006", delegateeTotalValidationRewards[0].Invitations2Balance.String())
	require.Nil(t, delegateeTotalValidationRewards[0].InvitationsBalance)
	require.Nil(t, delegateeTotalValidationRewards[0].Invitations3Balance)
	require.Nil(t, delegateeTotalValidationRewards[0].SavedInvitesBalance)
	require.Nil(t, delegateeTotalValidationRewards[0].SavedInvitesWinBalance)
	require.Nil(t, delegateeTotalValidationRewards[0].ReportsBalance)
	require.Equal(t, 1, delegateeTotalValidationRewards[0].Delegators)

	delegateeValidationRewards, err := testCommon.GetDelegateeValidationRewards(dbConnector)
	require.Nil(t, err)
	require.Len(t, delegateeValidationRewards, 1)
	require.Equal(t, 1, delegateeValidationRewards[0].Epoch)
	require.Equal(t, delegatee.Hex(), delegateeValidationRewards[0].DelegateeAddress)
	require.Equal(t, addr.Hex(), delegateeValidationRewards[0].DelegatorAddress)
	require.Equal(t, "0.0000000000000012", delegateeValidationRewards[0].TotalBalance.String())
	require.Equal(t, "0.0000000000000002", delegateeValidationRewards[0].ValidationBalance.String())
	require.Equal(t, "0.0000000000000004", delegateeValidationRewards[0].FlipsBalance.String())
	require.Equal(t, "0.0000000000000006", delegateeValidationRewards[0].Invitations2Balance.String())
	require.Nil(t, delegateeValidationRewards[0].InvitationsBalance)
	require.Nil(t, delegateeValidationRewards[0].Invitations3Balance)
	require.Nil(t, delegateeValidationRewards[0].SavedInvitesBalance)
	require.Nil(t, delegateeValidationRewards[0].SavedInvitesWinBalance)
	require.Nil(t, delegateeValidationRewards[0].ReportsBalance)

	// When
	err = dbAccessor.ResetTo(heightToReset)
	// Then
	require.Nil(t, err)
	addressSummaries, _ = testCommon.GetAddressSummaries(dbConnector)
	require.Equal(t, 1, len(addressSummaries))
	require.Equal(t, 1, addressSummaries[0].AddressId)
	require.Equal(t, 3, addressSummaries[0].Flips)
	require.Equal(t, 1, addressSummaries[0].WrongWordsFlips)

	rewardBounds, err = testCommon.GetRewardBounds(dbConnector)
	require.Nil(t, err)
	require.Empty(t, rewardBounds)

	balanceUpdates, err = testCommon.GetBalanceUpdates(dbConnector)
	require.Nil(t, err)
	require.Empty(t, balanceUpdates)
}
