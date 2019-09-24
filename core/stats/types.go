package stats

import (
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"math/big"
)

type RewardType byte

const (
	Validation        RewardType = 0
	Flips             RewardType = 1
	Invitations       RewardType = 2
	FoundationPayouts RewardType = 3
	ZeroWalletFund    RewardType = 4
)

type Stats struct {
	ValidationStats *statsTypes.ValidationStats
	RewardsStats    *RewardsStats
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
}

type RewardStats struct {
	Address common.Address
	Balance *big.Int
	Stake   *big.Int
	Type    RewardType
}
