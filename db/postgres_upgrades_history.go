package db

import (
	"github.com/pkg/errors"
)

const (
	getUpgradeVotingShortHistoryInfoQuery = "getUpgradeVotingShortHistoryInfo.sql"
	getUpgradeVotingHistoryQuery          = "getUpgradeVotingHistory.sql"
	updateUpgradeVotingShortHistoryQuery  = "updateUpgradeVotingShortHistory.sql"
	updateUpgradesQuery                   = "updateUpgrades.sql"
)

func (a *postgresAccessor) GetUpgradeVotingShortHistoryInfo(upgrade uint32) (*UpgradeVotingShortHistoryInfo, error) {
	res := &UpgradeVotingShortHistoryInfo{}
	err := a.db.QueryRow(a.getQuery(getUpgradeVotingShortHistoryInfoQuery), upgrade).Scan(
		&res.HistoryItems,
		&res.LastHeight,
		&res.LastStep,
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *postgresAccessor) GetUpgradeVotingHistory(upgrade uint32) ([]*UpgradeHistoryItem, error) {
	rows, err := a.db.Query(a.getQuery(getUpgradeVotingHistoryQuery), upgrade)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*UpgradeHistoryItem
	for rows.Next() {
		item := &UpgradeHistoryItem{}
		err = rows.Scan(
			&item.BlockHeight,
			&item.Votes,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (a *postgresAccessor) UpdateUpgradeVotingShortHistory(upgrade uint32, history []*UpgradeHistoryItem, lastStep uint32, lastHeight uint64) error {
	data := &upgradeVotingShortHistoryData{
		Upgrade:    upgrade,
		LastStep:   lastStep,
		LastHeight: lastHeight,
		History:    make([]*upgradeVotingShortHistoryItem, 0, len(history)),
	}
	for _, item := range history {
		data.History = append(data.History, &upgradeVotingShortHistoryItem{
			BlockHeight: item.BlockHeight,
			Votes:       item.Votes,
		})
	}
	_, err := a.db.Exec(a.getQuery(updateUpgradeVotingShortHistoryQuery), data)
	return errors.Wrap(err, "unable to update upgrade voting short history")
}

func (a *postgresAccessor) UpdateUpgrades(upgrades []*Upgrade) error {
	data := &upgradesData{
		Upgrades: make([]*upgrade, 0, len(upgrades)),
	}
	for _, item := range upgrades {
		data.Upgrades = append(data.Upgrades, &upgrade{
			Upgrade:             item.Upgrade,
			StartActivationDate: item.StartActivationDate,
			EndActivationDate:   item.EndActivationDate,
		})
	}
	_, err := a.db.Exec(a.getQuery(updateUpgradesQuery), data)
	return errors.Wrap(err, "unable to update upgrades")
}
