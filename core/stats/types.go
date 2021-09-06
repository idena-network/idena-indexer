package stats

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/core/state"
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
	ValidationStats                          *statsTypes.ValidationStats
	MinScoreForInvite                        *float32
	RewardsStats                             *RewardsStats
	MiningRewards                            []*MiningReward
	OriginalFinalCommittee                   map[common.Address]struct{}
	PoolFinalCommittee                       map[common.Address]struct{}
	ChargedPenaltiesByAddr                   map[common.Address]*big.Int
	BurntCoins                               *big.Int
	BurntCoinsByAddr                         map[common.Address][]*db.BurntCoins
	MintedCoins                              *big.Int
	BalanceUpdateAddrs                       mapset.Set
	ActivationTxTransfers                    []db.ActivationTxTransfer
	KillTxTransfers                          []db.KillTxTransfer
	KillInviteeTxTransfers                   []db.KillInviteeTxTransfer
	BalanceUpdates                           []*db.BalanceUpdate
	CommitteeRewardShare                     *big.Int
	IdentityStateChangesByTxHashAndAddress   map[common.Hash]map[common.Address]*IdentityStateChange
	FeesByTxHash                             map[common.Hash]*big.Int
	OracleVotingContracts                    []*db.OracleVotingContract
	OracleVotingContractCallStarts           []*db.OracleVotingContractCallStart
	OracleVotingContractCallVoteProofs       []*db.OracleVotingContractCallVoteProof
	OracleVotingContractCallVotes            []*db.OracleVotingContractCallVote
	OracleVotingContractCallFinishes         []*db.OracleVotingContractCallFinish
	OracleVotingContractCallProlongations    []*db.OracleVotingContractCallProlongation
	OracleVotingContractCallAddStakes        []*db.OracleVotingContractCallAddStake
	OracleVotingContractTerminations         []*db.OracleVotingContractTermination
	OracleLockContracts                      []*db.OracleLockContract
	OracleLockContractCallCheckOracleVotings []*db.OracleLockContractCallCheckOracleVoting
	OracleLockContractCallPushes             []*db.OracleLockContractCallPush
	OracleLockContractTerminations           []*db.OracleLockContractTermination
	RefundableOracleLockContracts            []*db.RefundableOracleLockContract
	RefundableOracleLockContractCallDeposits []*db.RefundableOracleLockContractCallDeposit
	RefundableOracleLockContractCallPushes   []*db.RefundableOracleLockContractCallPush
	RefundableOracleLockContractCallRefunds  []*db.RefundableOracleLockContractCallRefund
	RefundableOracleLockContractTerminations []*db.RefundableOracleLockContractTermination
	MultisigContracts                        []*db.MultisigContract
	MultisigContractCallAdds                 []*db.MultisigContractCallAdd
	MultisigContractCallSends                []*db.MultisigContractCallSend
	MultisigContractCallPushes               []*db.MultisigContractCallPush
	MultisigContractTerminations             []*db.MultisigContractTermination
	TimeLockContracts                        []*db.TimeLockContract
	TimeLockContractCallTransfers            []*db.TimeLockContractCallTransfer
	TimeLockContractTerminations             []*db.TimeLockContractTermination
	TxReceipts                               []*db.TxReceipt
	ContractTxsBalanceUpdates                []*db.ContractTxBalanceUpdates
	ActivationTxs                            []db.ActivationTx
	RemovedTransitiveDelegations             []db.RemovedTransitiveDelegation
}

type RewardsStats struct {
	ValidationResults                    map[common.ShardId]*types.ValidationResults
	Total                                *big.Int
	Validation                           *big.Int
	Flips                                *big.Int
	Reports                              *big.Int
	Invitations                          *big.Int
	FoundationPayouts                    *big.Int
	ZeroWalletFund                       *big.Int
	ValidationShare                      *big.Int
	FlipsShare                           *big.Int
	ReportsShare                         *big.Int
	InvitationsShare                     *big.Int
	Rewards                              []*RewardStats
	AgesByAddress                        map[string]uint16
	RewardedFlipCids                     []string
	RewardedInvites                      []*db.RewardedInvite
	SavedInviteRewardsCountByAddrAndType map[common.Address]map[RewardType]uint8
	ReportedFlipRewards                  []*db.ReportedFlipReward
	TotalRewardsByAddr                   map[common.Address]*big.Int
}

type RewardStats struct {
	Address common.Address
	Balance *big.Int
	Stake   *big.Int
	Type    RewardType
}

type IdentityStateChange struct {
	PrevState state.IdentityState
	NewState  state.IdentityState
}

type MiningReward struct {
	Address  common.Address
	Balance  *big.Int
	Stake    *big.Int
	Proposer bool
}
