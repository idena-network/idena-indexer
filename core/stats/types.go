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
	SavedInvite       RewardType = 7
	SavedInviteWin    RewardType = 8
	ReportedFlips     RewardType = 9
)

type Stats struct {
	ValidationStats        *statsTypes.ValidationStats
	MinScoreForInvite      *float32
	RewardsStats           *RewardsStats
	MiningRewards          []*db.MiningReward
	FinalCommittee         []common.Address
	BurntPenaltiesByAddr   map[common.Address]*big.Int
	BurntCoins             *big.Int
	BurntCoinsByAddr       map[common.Address][]*db.BurntCoins
	MintedCoins            *big.Int
	BalanceUpdateAddrs     mapset.Set
	ActivationTxTransfers  []db.ActivationTxTransfer
	KillTxTransfers        []db.KillTxTransfer
	KillInviteeTxTransfers []db.KillInviteeTxTransfer
	BalanceUpdates         []*db.BalanceUpdate
	CommitteeRewardShare   *big.Int
}

type RewardsStats struct {
	ValidationResults                    *types.ValidationResults
	Total                                *big.Int
	Validation                           *big.Int
	Flips                                *big.Int
	Invitations                          *big.Int
	FoundationPayouts                    *big.Int
	ZeroWalletFund                       *big.Int
	ValidationShare                      *big.Int
	FlipsShare                           *big.Int
	InvitationsShare                     *big.Int
	Rewards                              []*RewardStats
	AgesByAddress                        map[string]uint16
	RewardedFlipCids                     []string
	RewardedInvites                      []*db.RewardedInvite
	SavedInviteRewardsCountByAddrAndType map[common.Address]map[RewardType]uint8
	ReportedFlipRewards                  []*db.ReportedFlipReward
}

type RewardStats struct {
	Address common.Address
	Balance *big.Int
	Stake   *big.Int
	Type    RewardType
}
