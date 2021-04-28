package indexer

import (
	"fmt"
	"github.com/idena-network/idena-go/config"
	"github.com/idena-network/idena-indexer/core/holder/upgrade"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"math"
	"time"
)

func detectUpgradeVotes(upgradesVotes []*upgrade.Votes) []*db.UpgradeVotes {
	now := time.Now().UTC().Unix()
	activeVotings := detectActiveUpgradeVotings(config.ConsensusVersions, now)
	if len(activeVotings) == 0 {
		return nil
	}
	var res []*db.UpgradeVotes
	for _, upgradeVotes := range upgradesVotes {
		if _, ok := activeVotings[config.ConsensusVerson(upgradeVotes.Upgrade)]; !ok {
			continue
		}
		res = append(res, &db.UpgradeVotes{
			Upgrade:   upgradeVotes.Upgrade,
			Votes:     upgradeVotes.Votes,
			Timestamp: now,
		})
	}
	return res
}

func detectActiveUpgradeVotings(consensusVersionConfigs map[config.ConsensusVerson]*config.ConsensusConf, timestamp int64) map[config.ConsensusVerson]struct{} {
	if len(consensusVersionConfigs) == 0 {
		return nil
	}
	var res map[config.ConsensusVerson]struct{}
	for consensusVersion, consensusVersionConfig := range consensusVersionConfigs {
		if timestamp < consensusVersionConfig.StartActivationDate || timestamp > consensusVersionConfig.EndActivationDate {
			continue
		}
		if res == nil {
			res = make(map[config.ConsensusVerson]struct{})
		}
		res[consensusVersion] = struct{}{}
	}
	return res
}

func (indexer *Indexer) refreshUpgradeVotingHistorySummaries(upgradesVotes []*db.UpgradeVotes, height uint64) {
	select {
	case indexer.upgradeVotingHistoryCtx.queue <- &upgradesVotesWrapper{
		upgradesVotes: upgradesVotes,
		height:        height,
	}:
	default:
		log.Warn("upgradeVotingHistoryCtx.queue limit reached")
	}
}

func (indexer *Indexer) loopRefreshUpgradeVotingHistorySummaries() {
	for {
		upgradeVotesWrapper := <-indexer.upgradeVotingHistoryCtx.queue
		for _, upgradeVotes := range upgradeVotesWrapper.upgradesVotes {
			indexer.refreshUpgradeVotingHistorySummary(upgradeVotes.Upgrade, upgradeVotesWrapper.height)
		}
	}
}

func (indexer *Indexer) refreshUpgradeVotingHistorySummary(upgrade uint32, height uint64) {

	start := time.Now()

	info, err := indexer.db.GetUpgradeVotingShortHistoryInfo(upgrade)
	if err != nil {
		log.Error("Unable to load upgrade votes history info", "err", err)
		return
	}

	itemCount := indexer.upgradeVotingHistoryCtx.shortHistoryItems
	shift := height - info.LastHeight
	if uint64(itemCount) >= info.HistoryItems || shift < uint64(indexer.upgradeVotingHistoryCtx.shortHistoryMinShift) || shift < uint64(info.LastStep) {
		return
	}

	upgradeVotes, err := indexer.db.GetUpgradeVotingHistory(upgrade)
	if err != nil {
		log.Error("Unable to load upgrade votes history", "err", err)
		return
	}
	if len(upgradeVotes) == 0 || len(upgradeVotes) <= itemCount {
		return
	}
	var step float32
	if itemCount > 1 {
		step = float32(len(upgradeVotes)-1) / float32(itemCount-1)
	}
	summary := make([]*db.UpgradeHistoryItem, 0, itemCount)
	var lastStep uint32

	for i := 0; i < itemCount; i++ {
		idx := int(math.Round(float64(i) * float64(step)))
		summary = append(summary, upgradeVotes[idx])
		if i == itemCount-2 {
			lastStep = uint32(idx)
		} else if i == itemCount-1 {
			lastStep = uint32(idx) - lastStep
		}
	}

	if err := indexer.db.UpdateUpgradeVotingShortHistory(upgrade, summary, lastStep, height); err != nil {
		log.Error("Unable to update upgrade votes history summary", "err", err)
	}
	log.Debug(fmt.Sprintf("Refreshed upgrade voting short history, height: %v, upgrade: %v, duration: %v", height, upgrade, time.Since(start)))
}

func (indexer *Indexer) updateUpgradesInfo() {
	if len(config.ConsensusVersions) == 0 {
		return
	}
	upgrades := make([]*db.Upgrade, 0, len(config.ConsensusVersions))
	for consensusVersion, consensusConf := range config.ConsensusVersions {
		upgrades = append(upgrades, &db.Upgrade{
			Upgrade:             uint32(consensusVersion),
			StartActivationDate: consensusConf.StartActivationDate,
			EndActivationDate:   consensusConf.EndActivationDate,
		})
	}
	if err := indexer.db.UpdateUpgrades(upgrades); err != nil {
		log.Warn("Unable to update upgrades info", "err", err)
	}
}
