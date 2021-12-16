package indexer

import (
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"os"
	"path"
)

func (indexer *Indexer) makeEpochTreeSnapshot() error {
	if indexer.listener.NodeCtx().AppState.State.Epoch() == 0 {
		return nil
	}
	snapshotFilePath := indexer.getEpochTreeSnapshotFilePath()
	if _, err := os.Stat(snapshotFilePath); !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to check snapshot file")
	}
	log.Info("Start making epoch tree snapshot")
	if err := os.MkdirAll(indexer.treeSnapshotDir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create tree dump dir")
	}
	epochBlock := indexer.listener.NodeCtx().AppState.State.EpochBlock()
	snapshotFile, err := os.Create(snapshotFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to create snapshot file")
	}
	_, err = indexer.listener.NodeCtx().AppState.State.WriteSnapshot2(epochBlock, snapshotFile)
	snapshotFile.Close()
	if err != nil {
		indexer.removeEpochTreeSnapshot()
	}
	log.Info("Epoch tree snapshot made")
	return errors.Wrap(err, "failed to write snapshot")
}

func (indexer *Indexer) getEpochTreeSnapshotFilePath() string {
	epoch := indexer.listener.NodeCtx().AppState.State.Epoch()
	return path.Join(indexer.treeSnapshotDir, fmt.Sprintf("%v.tar", epoch))
}

func (indexer *Indexer) removeEpochTreeSnapshot() error {
	snapshotFilePath := indexer.getEpochTreeSnapshotFilePath()
	if _, err := os.Stat(snapshotFilePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(snapshotFilePath)
}
