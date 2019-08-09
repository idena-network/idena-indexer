package db

import (
	"github.com/idena-network/idena-indexer/explorer/types"
)

type Accessor interface {
	EpochsCount() (uint64, error)
	Epochs(startIndex uint64, count uint64) ([]types.EpochSummary, error)

	Epoch(epoch uint64) (types.EpochDetail, error)
	EpochBlocksCount(epoch uint64) (uint64, error)
	EpochBlocks(epoch uint64, startIndex uint64, count uint64) ([]types.BlockSummary, error)
	EpochFlipsCount(epoch uint64) (uint64, error)
	EpochFlips(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error)
	EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochIdentitiesCount(epoch uint64) (uint64, error)
	EpochIdentities(epoch uint64, startIndex uint64, count uint64) ([]types.EpochIdentitySummary, error)
	EpochIdentityStatesSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochInvitesSummary(epoch uint64) (types.InvitesSummary, error)
	EpochInvitesCount(epoch uint64) (uint64, error)
	EpochInvites(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error)
	EpochTxsCount(epoch uint64) (uint64, error)
	EpochTxs(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error)

	EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error)
	EpochIdentityShortFlipsToSolve(epoch uint64, address string) ([]string, error)
	EpochIdentityLongFlipsToSolve(epoch uint64, address string) ([]string, error)
	EpochIdentityShortAnswers(epoch uint64, address string) ([]types.Answer, error)
	EpochIdentityLongAnswers(epoch uint64, address string) ([]types.Answer, error)
	EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error)
	EpochIdentityValidationTxs(epoch uint64, address string) ([]types.TransactionSummary, error)

	BlockByHeight(height uint64) (types.BlockDetail, error)
	BlockTxsCountByHeight(height uint64) (uint64, error)
	BlockTxsByHeight(height uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error)
	BlockByHash(hash string) (types.BlockDetail, error)
	BlockTxsCountByHash(hash string) (uint64, error)
	BlockTxsByHash(hash string, startIndex uint64, count uint64) ([]types.TransactionSummary, error)

	Flip(hash string) (types.Flip, error)
	FlipContent(hash string) (types.FlipContent, error)
	FlipAnswersCount(hash string, isShort bool) (uint64, error)
	FlipAnswers(hash string, isShort bool, startIndex uint64, count uint64) ([]types.Answer, error)

	Identity(address string) (types.Identity, error)
	IdentityAge(address string) (uint64, error)
	IdentityCurrentFlipCids(address string) ([]string, error)
	IdentityEpochsCount(address string) (uint64, error)
	IdentityEpochs(address string, startIndex uint64, count uint64) ([]types.EpochIdentitySummary, error)
	IdentityFlipQualifiedAnswers(address string) ([]types.StrValueCount, error)
	IdentityFlipStates(address string) ([]types.StrValueCount, error)
	IdentityInvitesCount(address string) (uint64, error)
	IdentityInvites(address string, startIndex uint64, count uint64) ([]types.Invite, error)
	IdentityStatesCount(address string) (uint64, error)
	IdentityStates(address string, startIndex uint64, count uint64) ([]types.IdentityState, error)
	IdentityTxsCount(address string) (uint64, error)
	IdentityTxs(address string, startIndex uint64, count uint64) ([]types.TransactionSummary, error)

	Address(address string) (types.Address, error)

	Transaction(hash string) (types.TransactionDetail, error)

	Destroy()
}
