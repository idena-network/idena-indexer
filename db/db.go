package db

import (
	"github.com/idena-network/idena-go/common"
	data2 "github.com/idena-network/idena-indexer/data"
	"math/big"
	"time"
)

type Accessor interface {
	data2.DbAccessor

	GetLastHeight() (uint64, error)
	Save(data *Data) error
	SaveRestoredData(data *RestoredData) error
	SaveMemPoolData(data *MemPoolData) error

	SaveFlipSize(flipCid string, size int) error
	GetEpochFlipsWithoutSize(epoch uint64, limit int) (cids []string, err error)
	GetFlipsToLoadContent(timestamp *big.Int, limit int) ([]*FlipToLoadContent, error)
	SaveFlipsContent(failedFlips []*FailedFlipContent, flipsContent []*FlipContent) error

	GetUpgradeVotingShortHistoryInfo(upgrade uint32) (*UpgradeVotingShortHistoryInfo, error)
	GetUpgradeVotingHistory(upgrade uint32) ([]*UpgradeHistoryItem, error)
	UpdateUpgradeVotingShortHistory(upgrade uint32, history []*UpgradeHistoryItem, lastStep uint32, lastHeight uint64) error
	UpdateUpgrades(upgrades []*Upgrade) error

	SavePeersCount(count int, timestamp time.Time) error
	SaveVoteCountingStepResult(value *VoteCountingStepResult) error
	SaveVoteCountingResult(value *VoteCountingResult) error
	SaveProofProposal(value *ProofProposal) error
	SaveBlockProposal(value *BlockProposal) error
	DeleteVoteCountingOldData() error

	ActualOracleVotings() ([]common.Address, error)

	Destroy()
	ResetTo(height uint64) error
}
