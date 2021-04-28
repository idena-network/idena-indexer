package db

import "math/big"

type Accessor interface {
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

	Destroy()
	ResetTo(height uint64) error
}
