package db

import (
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/shopspring/decimal"
	"time"
)

type Accessor interface {
	Search(value string) ([]types.Entity, error)

	Coins() (types.AllCoins, error)
	CirculatingSupply() (decimal.Decimal, error)

	EpochsCount() (uint64, error)
	Epochs(startIndex uint64, count uint64) ([]types.EpochSummary, error)

	LastEpoch() (types.EpochDetail, error)
	Epoch(epoch uint64) (types.EpochDetail, error)
	EpochBlocksCount(epoch uint64) (uint64, error)
	EpochBlocks(epoch uint64, startIndex uint64, count uint64) ([]types.BlockSummary, error)
	EpochFlipsCount(epoch uint64) (uint64, error)
	EpochFlips(epoch uint64, startIndex uint64, count uint64) ([]types.FlipSummary, error)
	EpochFlipAnswersSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochFlipStatesSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochFlipWrongWordsSummary(epoch uint64) ([]types.NullableBoolValueCount, error)
	EpochIdentitiesCount(epoch uint64, prevStates []string, states []string) (uint64, error)
	EpochIdentities(epoch uint64, prevStates []string, states []string, startIndex uint64,
		count uint64) ([]types.EpochIdentitySummary, error)
	EpochIdentityStatesSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochInvitesSummary(epoch uint64) (types.InvitesSummary, error)
	EpochInviteStatesSummary(epoch uint64) ([]types.StrValueCount, error)
	EpochInvitesCount(epoch uint64) (uint64, error)
	EpochInvites(epoch uint64, startIndex uint64, count uint64) ([]types.Invite, error)
	EpochTxsCount(epoch uint64) (uint64, error)
	EpochTxs(epoch uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error)
	EpochCoins(epoch uint64) (types.AllCoins, error)
	EpochRewardsSummary(epoch uint64) (types.RewardsSummary, error)
	EpochBadAuthorsCount(epoch uint64) (uint64, error)
	EpochBadAuthors(epoch uint64, startIndex uint64, count uint64) ([]types.BadAuthor, error)
	EpochGoodAuthorsCount(epoch uint64) (uint64, error)
	EpochGoodAuthors(epoch uint64, startIndex uint64, count uint64) ([]types.AuthorValidationSummary, error)
	EpochRewardsCount(epoch uint64) (uint64, error)
	EpochRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Reward, error)
	EpochIdentitiesRewardsCount(epoch uint64) (uint64, error)
	EpochIdentitiesRewards(epoch uint64, startIndex uint64, count uint64) ([]types.Rewards, error)
	EpochFundPayments(epoch uint64) ([]types.FundPayment, error)

	EpochIdentity(epoch uint64, address string) (types.EpochIdentity, error)
	EpochIdentityShortFlipsToSolve(epoch uint64, address string) ([]string, error)
	EpochIdentityLongFlipsToSolve(epoch uint64, address string) ([]string, error)
	EpochIdentityShortAnswers(epoch uint64, address string) ([]types.Answer, error)
	EpochIdentityLongAnswers(epoch uint64, address string) ([]types.Answer, error)
	EpochIdentityFlips(epoch uint64, address string) ([]types.FlipSummary, error)
	EpochIdentityValidationTxs(epoch uint64, address string) ([]types.TransactionSummary, error)
	EpochIdentityRewards(epoch uint64, address string) ([]types.Reward, error)

	BlockByHeight(height uint64) (types.BlockDetail, error)
	BlockTxsCountByHeight(height uint64) (uint64, error)
	BlockTxsByHeight(height uint64, startIndex uint64, count uint64) ([]types.TransactionSummary, error)
	BlockByHash(hash string) (types.BlockDetail, error)
	BlockTxsCountByHash(hash string) (uint64, error)
	BlockTxsByHash(hash string, startIndex uint64, count uint64) ([]types.TransactionSummary, error)
	BlockCoinsByHeight(height uint64) (types.AllCoins, error)
	BlockCoinsByHash(hash string) (types.AllCoins, error)

	Flip(hash string) (types.Flip, error)
	FlipContent(hash string) (types.FlipContent, error)
	FlipAnswersCount(hash string, isShort bool) (uint64, error)
	FlipAnswers(hash string, isShort bool, startIndex uint64, count uint64) ([]types.Answer, error)
	FlipEpochAdjacentFlips(hash string) (types.AdjacentStrValues, error)
	FlipAddressAdjacentFlips(hash string) (types.AdjacentStrValues, error)
	FlipEpochIdentityAdjacentFlips(hash string) (types.AdjacentStrValues, error)

	Identity(address string) (types.Identity, error)
	IdentityAge(address string) (uint64, error)
	IdentityCurrentFlipCids(address string) ([]string, error)
	IdentityEpochsCount(address string) (uint64, error)
	IdentityEpochs(address string, startIndex uint64, count uint64) ([]types.EpochIdentitySummary, error)
	IdentityFlipsCount(address string) (uint64, error)
	IdentityFlips(address string, startIndex uint64, count uint64) ([]types.FlipSummary, error)
	IdentityFlipQualifiedAnswers(address string) ([]types.StrValueCount, error)
	IdentityFlipStates(address string) ([]types.StrValueCount, error)
	IdentityInvitesCount(address string) (uint64, error)
	IdentityInvites(address string, startIndex uint64, count uint64) ([]types.Invite, error)
	IdentityTxsCount(address string) (uint64, error)
	IdentityTxs(address string, startIndex uint64, count uint64) ([]types.TransactionSummary, error)
	IdentityRewardsCount(address string) (uint64, error)
	IdentityRewards(address string, startIndex uint64, count uint64) ([]types.Reward, error)
	IdentityEpochRewardsCount(address string) (uint64, error)
	IdentityEpochRewards(address string, startIndex uint64, count uint64) ([]types.Rewards, error)

	Address(address string) (types.Address, error)
	AddressPenaltiesCount(address string) (uint64, error)
	AddressPenalties(address string, startIndex uint64, count uint64) ([]types.Penalty, error)
	AddressMiningRewardsCount(address string) (uint64, error)
	AddressMiningRewards(address string, startIndex uint64, count uint64) ([]types.Reward, error)
	AddressBlockMiningRewardsCount(address string) (uint64, error)
	AddressBlockMiningRewards(address string, startIndex uint64, count uint64) ([]types.BlockRewards, error)
	AddressStatesCount(address string) (uint64, error)
	AddressStates(address string, startIndex uint64, count uint64) ([]types.AddressState, error)
	AddressTotalLatestMiningReward(afterTime time.Time, address string) (types.TotalMiningReward, error)
	AddressTotalLatestBurntCoins(afterTime time.Time, address string) (types.AddressBurntCoins, error)
	AddressBadAuthorsCount(address string) (uint64, error)
	AddressBadAuthors(address string, startIndex uint64, count uint64) ([]types.BadAuthor, error)

	Transaction(hash string) (types.TransactionDetail, error)

	BalancesCount() (uint64, error)
	Balances(startIndex uint64, count uint64) ([]types.Balance, error)

	TotalLatestMiningRewardsCount(afterTime time.Time) (uint64, error)
	TotalLatestMiningRewards(afterTime time.Time, startIndex uint64, count uint64) ([]types.TotalMiningReward, error)
	TotalLatestBurntCoinsCount(afterTime time.Time) (uint64, error)
	TotalLatestBurntCoins(afterTime time.Time, startIndex uint64, count uint64) ([]types.AddressBurntCoins, error)

	Destroy()
}
