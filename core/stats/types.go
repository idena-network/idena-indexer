package stats

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/db"
	"math/big"
)

type RewardType byte

const (
	Validation        RewardType = 0
	Flips             RewardType = 1
	Invitations       RewardType = 2
	FoundationPayouts RewardType = 3
	ZeroWalletFund    RewardType = 4
	Invitations2      RewardType = 5
	Invitations3      RewardType = 6
)

type Stats struct {
	ValidationStats      *statsTypes.ValidationStats
	RewardsStats         *RewardsStats
	MiningRewards        []*db.MiningReward
	FinalCommittee       []common.Address
	BurntPenaltiesByAddr map[common.Address]*big.Int
	BurntCoins           *big.Int
	BurntCoinsByAddr     map[common.Address][]*db.BurntCoins
	MintedCoins          *big.Int
	BalanceUpdateAddrs   mapset.Set
	KilledAddrs          mapset.Set
	ActivationTxs        []db.ActivationTxSpecificPart
	KillTxs              []db.KillTxSpecificPart
	KillInviteeTxs       []db.KillInviteeTxSpecificPart
}

type RewardsStats struct {
	Authors           *types.ValidationAuthors
	Total             *big.Int
	Validation        *big.Int
	Flips             *big.Int
	Invitations       *big.Int
	FoundationPayouts *big.Int
	ZeroWalletFund    *big.Int
	Rewards           []*RewardStats
	AgesByAddress     map[string]uint16
}

type RewardStats struct {
	Address common.Address
	Balance *big.Int
	Stake   *big.Int
	Type    RewardType
}

type BalanceUpdate struct {
	Balance *big.Int
	Stake   *big.Int
}
