package db

import (
	"github.com/idena-network/idena-go/common"
	"github.com/shopspring/decimal"
	"math/big"
)

type BurntCoinsReason = uint8
type BalanceUpdateReason = uint8

type ContractType = uint8

type ContractCallMethod = uint8
type TimeLockCall = ContractCallMethod
type OracleVotingCall = ContractCallMethod
type OracleLockCall = ContractCallMethod
type MultisigCall = ContractCallMethod
type RefundableOracleLockCall = ContractCallMethod

const (
	PenaltyBurntCoins      BurntCoinsReason = 0x0
	InviteBurntCoins       BurntCoinsReason = 0x1
	FeeBurntCoins          BurntCoinsReason = 0x2
	KilledBurntCoins       BurntCoinsReason = 0x4
	BurnTxBurntCoins       BurntCoinsReason = 0x5
	DustClearingBurntCoins BurntCoinsReason = 0x6

	TxReason                          BalanceUpdateReason = 0x0
	VerifiedStakeTransferReason       BalanceUpdateReason = 0x1
	ProposerRewardReason              BalanceUpdateReason = 0x2
	CommitteeRewardReason             BalanceUpdateReason = 0x3
	EpochRewardReason                 BalanceUpdateReason = 0x4
	FailedValidationReason            BalanceUpdateReason = 0x5
	PenaltyReason                     BalanceUpdateReason = 0x6
	EpochPenaltyResetReason           BalanceUpdateReason = 0x7
	DustClearingReason                BalanceUpdateReason = 0x9
	EmbeddedContractReason            BalanceUpdateReason = 0xA
	EmbeddedContractTerminationReason BalanceUpdateReason = 0xB
	DelegatorEpochRewardReason        BalanceUpdateReason = 0xC
	DelegateeEpochRewardReason        BalanceUpdateReason = 0xD

	TimeLockCallTransfer TimeLockCall = 0

	OracleVotingCallStart     OracleVotingCall = 0
	OracleVotingCallVoteProof OracleVotingCall = 1
	OracleVotingCallVote      OracleVotingCall = 2
	OracleVotingCallFinish    OracleVotingCall = 3
	OracleVotingCallProlong   OracleVotingCall = 4
	OracleVotingCallAddStake  OracleVotingCall = 5

	OracleLockCallPush              OracleLockCall = 0
	OracleLockCallCheckOracleVoting OracleLockCall = 1

	MultisigCallAdd  MultisigCall = 0
	MultisigCallSend MultisigCall = 1
	MultisigCallPush MultisigCall = 2

	RefundableOracleLockCallDeposit RefundableOracleLockCall = 0
	RefundableOracleLockCallPush    RefundableOracleLockCall = 1
	RefundableOracleLockCallRefund  RefundableOracleLockCall = 2
)

type RestoredData struct {
	Balances    []Balance
	Birthdays   []Birthday
	PoolSizes   []*PoolSize
	Delegations []*Delegation
}

type PoolSize struct {
	Address        common.Address
	Size           uint64
	TotalDelegated uint64
}

type Delegation struct {
	Delegator  common.Address
	Delegatee  common.Address
	BirthEpoch *uint16
}

type Data struct {
	Epoch                                    uint64
	PrevStateRoot                            string
	ValidationTime                           big.Int
	Block                                    Block
	ActivationTxTransfers                    []ActivationTxTransfer
	KillTxTransfers                          []KillTxTransfer
	KillInviteeTxTransfers                   []KillInviteeTxTransfer
	ActivationTxs                            []ActivationTx
	KillInviteeTxs                           []KillInviteeTx
	BecomeOnlineTxs                          []string
	BecomeOfflineTxs                         []string
	SubmittedFlips                           []Flip
	DeletedFlips                             []DeletedFlip
	FlipKeys                                 []FlipKey
	FlipsWords                               []FlipWords
	Addresses                                []Address
	ChangedBalances                          []Balance
	Coins                                    Coins
	Penalties                                []Penalty
	MiningRewards                            []*MiningReward
	BurntCoinsPerAddr                        map[common.Address][]*BurntCoins
	BalanceUpdates                           []*BalanceUpdate
	CommitteeRewardShare                     *big.Int
	OracleVotingContracts                    []*OracleVotingContract
	OracleVotingContractCallStarts           []*OracleVotingContractCallStart
	OracleVotingContractCallVoteProofs       []*OracleVotingContractCallVoteProof
	OracleVotingContractCallVotes            []*OracleVotingContractCallVote
	OracleVotingContractCallFinishes         []*OracleVotingContractCallFinish
	OracleVotingContractCallProlongations    []*OracleVotingContractCallProlongation
	OracleVotingContractCallAddStakes        []*OracleVotingContractCallAddStake
	OracleVotingContractTerminations         []*OracleVotingContractTermination
	ClearOldEpochCommittees                  bool
	OracleLockContracts                      []*OracleLockContract
	OracleLockContractCallCheckOracleVotings []*OracleLockContractCallCheckOracleVoting
	OracleLockContractCallPushes             []*OracleLockContractCallPush
	OracleLockContractTerminations           []*OracleLockContractTermination
	RefundableOracleLockContracts            []*RefundableOracleLockContract
	RefundableOracleLockContractCallDeposits []*RefundableOracleLockContractCallDeposit
	RefundableOracleLockContractCallPushes   []*RefundableOracleLockContractCallPush
	RefundableOracleLockContractCallRefunds  []*RefundableOracleLockContractCallRefund
	RefundableOracleLockContractTerminations []*RefundableOracleLockContractTermination
	MultisigContracts                        []*MultisigContract
	MultisigContractCallAdds                 []*MultisigContractCallAdd
	MultisigContractCallSends                []*MultisigContractCallSend
	MultisigContractCallPushes               []*MultisigContractCallPush
	MultisigContractTerminations             []*MultisigContractTermination
	TimeLockContracts                        []*TimeLockContract
	TimeLockContractCallTransfers            []*TimeLockContractCallTransfer
	TimeLockContractTerminations             []*TimeLockContractTermination
	TxReceipts                               []*TxReceipt
	ContractTxsBalanceUpdates                []*ContractTxBalanceUpdates
	EpochResult                              *EpochResult
	DelegationSwitches                       []*DelegationSwitch
	UpgradesVotes                            []*UpgradeVotes
	PoolSizes                                []PoolSize
	MinersHistoryItem                        *MinersHistoryItem
	RemovedTransitiveDelegations             []RemovedTransitiveDelegation
	EpochSummaryUpdate                       EpochSummaryUpdate
	OracleVotingContractsToProlong           []common.Address
}

type EpochRewards struct {
	BadAuthors             []*BadAuthor
	Total                  *TotalRewards
	ValidationRewards      []*Reward
	FundRewards            []*Reward
	AgesByAddress          map[string]uint16
	StakedAmountsByAddress map[string]*big.Int
	RewardedFlipCids       []string
	RewardedInvitations    []*RewardedInvite
	SavedInviteRewards     []*SavedInviteRewards
	ReportedFlipRewards    []*ReportedFlipReward
}

type RewardedInvite struct {
	TxHash      string
	Type        byte
	EpochHeight uint32
}

type SavedInviteRewards struct {
	Address string
	Type    byte
	Count   uint8
}

type ReportedFlipReward struct {
	Address string
	Balance decimal.Decimal
	Stake   decimal.Decimal
	Cid     string
}

type TotalRewards struct {
	Total             decimal.Decimal
	Validation        decimal.Decimal
	Staking           decimal.Decimal
	Candidate         decimal.Decimal
	Flips             decimal.Decimal
	Reports           decimal.Decimal
	Invitations       decimal.Decimal
	FoundationPayouts decimal.Decimal
	ZeroWalletFund    decimal.Decimal
	ValidationShare   decimal.Decimal
	StakingShare      decimal.Decimal
	CandidateShare    decimal.Decimal
	FlipsShare        decimal.Decimal
	ReportsShare      decimal.Decimal
	InvitationsShare  decimal.Decimal
}

type Reward struct {
	Address string
	Balance decimal.Decimal
	Stake   decimal.Decimal
	Type    byte
}

type BadAuthor struct {
	Address string
	Reason  byte
}

type MiningReward struct {
	Address  string
	Balance  decimal.Decimal
	Stake    decimal.Decimal
	Proposer bool
}

type Block struct {
	ValidationFinished      bool
	Height                  uint64
	Hash                    string
	Transactions            []Transaction
	Time                    int64
	Proposer                string
	Flags                   []string
	IsEmpty                 bool
	BodySize                int
	FullSize                int
	VrfProposerThreshold    float64
	OriginalValidatorsCount int
	PoolValidatorsCount     int
	ProposerVrfScore        float64
	FeeRate                 decimal.Decimal
	Upgrade                 *uint32
	OfflineAddress          *string
}

type Transaction struct {
	Hash   string          `json:"hash"`
	Type   uint16          `json:"type"`
	From   string          `json:"from"`
	To     string          `json:"to"`
	Amount decimal.Decimal `json:"amount"`
	Tips   decimal.Decimal `json:"tips"`
	MaxFee decimal.Decimal `json:"maxFee"`
	Fee    decimal.Decimal `json:"fee"`
	Size   int             `json:"size"`
	Raw    string          `json:"raw"`
	Nonce  uint32          `json:"nonce"`

	Data interface{} `json:"data,omitempty"`
}

type ShortAnswerTxData struct {
	ClientType byte `json:"clientType"`
}

type ActivationTxTransfer struct {
	TxHash          string
	BalanceTransfer decimal.Decimal
}

type KillTxTransfer struct {
	TxHash        string
	StakeTransfer decimal.Decimal
}

type KillInviteeTxTransfer struct {
	TxHash        string
	StakeTransfer decimal.Decimal
}

type ActivationTx struct {
	TxHash       string
	InviteTxHash string
	ShardId      common.ShardId
}

type KillInviteeTx struct {
	TxHash       string
	InviteTxHash string
}

type EpochIdentity struct {
	Address              string
	State                uint8
	ShortPoint           float32
	ShortFlips           uint32
	TotalShortPoint      float32
	TotalShortFlips      uint32
	LongPoint            float32
	LongFlips            uint32
	Approved             bool
	Missed               bool
	RequiredFlips        uint8
	AvailableFlips       uint8
	MadeFlips            uint8
	NextEpochInvites     uint8
	BirthEpoch           uint64
	ShortFlipCidsToSolve []string
	LongFlipCidsToSolve  []string
	DelegateeAddress     string
	ShardId              common.ShardId
	NewShardId           common.ShardId
}

type Flip struct {
	TxId   uint64
	TxHash string
	Cid    string
	Pair   uint8
}

type DeletedFlip struct {
	TxHash string
	Cid    string
}

type FlipStats struct {
	Author       string
	Cid          string
	ShortAnswers []Answer
	LongAnswers  []Answer
	Status       byte
	Answer       byte
	Grade        byte
}

type Answer struct {
	Address    string
	Answer     byte
	Point      float32
	Grade      byte
	Index      uint16
	Considered bool
}

type FlipKey struct {
	TxHash string
	Key    string
}

type MemPoolFlipKey struct {
	Address string
	Key     string
}

type FlipWords struct {
	Cid    string
	TxHash string
	Word1  uint16
	Word2  uint16
}

type FlipContent struct {
	Cid    string
	Pics   [][]byte
	Orders [][]byte
	Icon   []byte
}

type Address struct {
	Address      string
	IsTemporary  bool
	StateChanges []AddressStateChange
}

type AddressStateChange struct {
	PrevState uint8
	NewState  uint8
	TxHash    string
}

type Balance struct {
	Address string
	TxHash  string
	Balance decimal.Decimal
	Stake   decimal.Decimal
}

type Coins struct {
	Minted       decimal.Decimal
	Burnt        decimal.Decimal
	TotalBalance decimal.Decimal
	TotalStake   decimal.Decimal
}

type AddressFlipCid struct {
	FlipId  uint64
	Address string
	Cid     string
}

type Penalty struct {
	Address string
	Penalty decimal.Decimal
}

type Birthday struct {
	Address    string
	BirthEpoch uint64
}

type BurntCoins struct {
	Amount decimal.Decimal
	Reason BurntCoinsReason
	TxHash string
}

type MemPoolData struct {
	FlipKeyTimestamps       []*MemPoolActionTimestamp
	AnswersHashTxTimestamps []*MemPoolActionTimestamp
}

type MemPoolActionTimestamp struct {
	Address string
	Epoch   uint64
	Time    *big.Int
}

type FlipToLoadContent struct {
	Cid      string
	Key      string
	Attempts int
}

type FailedFlipContent struct {
	Cid                  string
	AttemptsLimitReached bool
	NextAttemptTimestamp *big.Int
}

type BalanceUpdate struct {
	Address    common.Address
	BalanceOld *big.Int
	StakeOld   *big.Int
	PenaltyOld *big.Int
	BalanceNew *big.Int
	StakeNew   *big.Int
	PenaltyNew *big.Int
	TxHash     *common.Hash
	Reason     BalanceUpdateReason
}

type EpochResult struct {
	Identities                []EpochIdentity
	FlipStats                 []FlipStats
	Birthdays                 []Birthday
	MemPoolFlipKeys           []*MemPoolFlipKey
	FailedValidation          bool
	EpochRewards              *EpochRewards
	MinScoreForInvite         float32
	RewardsBounds             []*RewardBounds
	FlipStatuses              []FlipStatusCount
	ReportedFlips             uint32
	DelegateesEpochRewards    []DelegateeEpochRewards
	ValidationRewardSummaries []ValidationRewardSummaries
}

type OracleVotingContract struct {
	TxHash               common.Hash
	ContractAddress      common.Address
	Stake                *big.Int
	StartTime            uint64
	VotingDuration       uint64
	VotingMinPayment     *big.Int
	Fact                 []byte
	State                byte
	PublicVotingDuration uint64
	WinnerThreshold      byte
	Quorum               byte
	CommitteeSize        uint64
	OwnerFee             byte
}

type OracleVotingContractCallStart struct {
	TxHash           common.Hash
	State            byte
	StartHeight      uint64
	Epoch            uint16
	VotingMinPayment *big.Int
	VrfSeed          []byte
	Committee        []common.Address
}

type OracleVotingContractCallVoteProof struct {
	TxHash              common.Hash
	VoteHash            []byte
	Votes               uint64
	NewSecretVotesCount *uint64
	Discriminated       bool
}

type OracleVotingContractCallVote struct {
	TxHash           common.Hash
	Vote             byte
	Salt             []byte
	OptionVotes      *uint64
	OptionAllVotes   *uint64 // TODO make non-pointer after deprecated AddOracleVotingCallVoteOld is removed
	SecretVotesCount *uint64
	Delegatee        *common.Address
	Discriminated    bool
	PrevPoolVote     *byte
	PrevOptionVotes  *uint64
}

type OracleVotingContractCallFinish struct {
	TxHash       common.Hash
	State        byte
	Result       *byte
	Fund         *big.Int
	OracleReward *big.Int
	OwnerReward  *big.Int
}

type OracleVotingContractCallProlongation struct {
	TxHash             common.Hash
	Epoch              uint16
	StartBlock         *uint64
	VrfSeed            []byte
	EpochWithoutGrowth *byte
	ProlongVoteCount   *uint64
	Committee          []common.Address
}

type OracleVotingContractCallAddStake struct {
	TxHash common.Hash
}

type OracleVotingContractTermination struct {
	TxHash       common.Hash
	Fund         *big.Int
	OracleReward *big.Int
	OwnerReward  *big.Int
}

type OracleLockContract struct {
	TxHash              common.Hash
	ContractAddress     common.Address
	Stake               *big.Int
	OracleVotingAddress common.Address
	ExpectedValue       byte
	SuccessAddress      common.Address
	FailAddress         common.Address
}

type OracleLockContractCallCheckOracleVoting struct {
	TxHash             common.Hash
	OracleVotingResult *byte
}

type OracleLockContractCallPush struct {
	TxHash             common.Hash
	Success            bool
	OracleVotingResult byte
	Transfer           *big.Int
}

type OracleLockContractTermination struct {
	TxHash common.Hash
	Dest   common.Address
}

type RefundableOracleLockContract struct {
	TxHash              common.Hash
	ContractAddress     common.Address
	Stake               *big.Int
	OracleVotingAddress common.Address
	ExpectedValue       byte
	SuccessAddress      *common.Address
	FailAddress         *common.Address
	RefundDelay         uint64
	DepositDeadline     uint64
	OracleVotingFee     byte
}

type RefundableOracleLockContractCallDeposit struct {
	TxHash common.Hash
	OwnSum *big.Int
	Sum    *big.Int
	Fee    *big.Int
}

type RefundableOracleLockContractCallPush struct {
	TxHash             common.Hash
	State              byte
	OracleVotingExists bool
	OracleVotingResult *byte
	Transfer           *big.Int
	RefundBlock        uint64
}

type RefundableOracleLockContractCallRefund struct {
	TxHash  common.Hash
	Balance *big.Int
	Coef    float64
}

type RefundableOracleLockContractTermination struct {
	TxHash common.Hash
	Dest   common.Address
}

type MultisigContract struct {
	TxHash          common.Hash
	ContractAddress common.Address
	Stake           *big.Int
	MinVotes        byte
	MaxVotes        byte
	State           byte
}

type MultisigContractCallAdd struct {
	TxHash   common.Hash
	Address  common.Address
	NewState *byte
}

type MultisigContractCallSend struct {
	TxHash common.Hash
	Dest   common.Address
	Amount *big.Int
}

type MultisigContractCallPush struct {
	TxHash         common.Hash
	Dest           common.Address
	Amount         *big.Int
	VoteAddressCnt byte
	VoteAmountCnt  byte
}

type MultisigContractTermination struct {
	TxHash common.Hash
	Dest   common.Address
}

type TimeLockContract struct {
	TxHash          common.Hash
	ContractAddress common.Address
	Stake           *big.Int
	Timestamp       uint64
}

type TimeLockContractCallTransfer struct {
	TxHash common.Hash
	Dest   common.Address
	Amount *big.Int
}

type TimeLockContractTermination struct {
	TxHash common.Hash
	Dest   common.Address
}

type TxReceipt struct {
	TxHash  common.Hash
	Success bool
	GasUsed uint64
	GasCost *big.Int
	Method  string
	Error   string
}

type ContractTxBalanceUpdates struct {
	TxHash             common.Hash
	ContractAddress    common.Address
	ContractCallMethod *ContractCallMethod
	Updates            []*ContractTxBalanceUpdate
}

type ContractTxBalanceUpdate struct {
	Address    common.Address
	BalanceOld *big.Int
	BalanceNew *big.Int
}

type RewardBounds struct {
	Type     byte
	Min, Max *RewardBound
}

type RewardBound struct {
	Amount  *big.Int
	Address common.Address
}

type DelegationSwitch struct {
	Delegator  common.Address
	Delegatee  *common.Address
	BirthEpoch *uint16
}

type UpgradeVotes struct {
	Upgrade   uint32
	Votes     uint64
	Timestamp int64
}

type UpgradeHistoryItem struct {
	BlockHeight uint64
	Votes       uint64
}

type UpgradeVotingShortHistoryInfo struct {
	HistoryItems uint64
	LastStep     uint32
	LastHeight   uint64
}

type Upgrade struct {
	Upgrade             uint32
	StartActivationDate int64
	EndActivationDate   int64
}

type MinersHistoryItem struct {
	OnlineValidators uint64 `json:"onlineValidators"`
	OnlineMiners     uint64 `json:"onlineMiners"`
}

type FlipStatusCount struct {
	Status byte   `json:"status"`
	Count  uint64 `json:"count"`
}

type RemovedTransitiveDelegation struct {
	Delegator, Delegatee common.Address
}

type DelegateeEpochRewards struct {
	Address          common.Address
	TotalRewards     []DelegationEpochReward
	DelegatorRewards []DelegatorEpochReward
}

type DelegatorEpochReward struct {
	Address      common.Address
	TotalRewards []DelegationEpochReward
}

type DelegationEpochReward struct {
	Balance *big.Int
	Type    byte
}

type ValidationRewardSummaries struct {
	Address     string
	Validation  ValidationRewardSummary
	Flips       ValidationRewardSummary
	Invitations ValidationRewardSummary
	Reports     ValidationRewardSummary
	Candidate   ValidationRewardSummary
	Staking     ValidationRewardSummary
}

type ValidationRewardSummary struct {
	Earned       *big.Int
	Missed       *big.Int
	MissedReason *byte
}

type VoteCountingStepResult struct {
	Timestamp           int64   `json:"timestamp"`
	Round               uint64  `json:"round"`
	Step                uint8   `json:"step"`
	NecessaryVotesCount int     `json:"necessaryVotesCount"`
	CheckedRoundVotes   int     `json:"checkedRoundVotes"`
	Votes               []*Vote `json:"votes,omitempty"`
}

type Vote struct {
	Voter       common.Address `json:"voter"`
	ParentHash  common.Hash    `json:"parentHash"`
	VotedHash   common.Hash    `json:"votedHash"`
	TurnOffline bool           `json:"turnOffline"`
	Upgrade     uint32         `json:"upgrade"`
}

type VoteCountingResult struct {
	Timestamp  int64           `json:"timestamp"`
	Round      uint64          `json:"round"`
	Step       uint8           `json:"step"`
	Validators *StepValidators `json:"validators"`
	Hash       common.Hash     `json:"hash"`
	Cert       *FullBlockCert  `json:"cert"`
	Err        *string         `json:"err"`
}

type StepValidators struct {
	Original           []common.Address `json:"original,omitempty"`
	Validators         []common.Address `json:"validators,omitempty"`
	ApprovedValidators []common.Address `json:"approvedValidators,omitempty"`
	Size               int              `json:"size"`
}

type FullBlockCert struct {
	Votes []*Vote `json:"votes,omitempty"`
}

type ProofProposal struct {
	Timestamp int64          `json:"timestamp"`
	Round     uint64         `json:"round"`
	Proposer  common.Address `json:"proposer"`
	Hash      common.Hash    `json:"hash"`
	Modifier  int64          `json:"modifier"`
	VrfScore  float64        `json:"vrfScore"`
}

type BlockProposal struct {
	ReceivingTime int64          `json:"receivingTime"`
	Height        uint64         `json:"height"`
	Proposer      common.Address `json:"proposer"`
	Hash          common.Hash    `json:"hash"`
}

type EpochSummaryUpdate struct {
	CandidateCountDiff int `json:"candidateCountDiff,omitempty"`
}
