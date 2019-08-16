package report

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/idena-network/idena-indexer/log"
	"github.com/idena-network/idena-indexer/report/db"
	"github.com/idena-network/idena-indexer/report/types"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type Provider struct {
	db        db.Accessor
	outputDir string
	logger    log.Logger
}

func NewProvider(db db.Accessor, outputDir string, logger log.Logger) *Provider {
	return &Provider{
		db:        db,
		outputDir: outputDir,
		logger:    logger,
	}
}

func (p *Provider) ExportAllFlips() error {
	epochsCount, err := p.db.EpochsCount()
	if err != nil {
		return errors.Wrap(err, "Unable to get epochs count")
	}
	p.logger.Debug("Got epochs count", "c", epochsCount)
	for epoch := uint64(0); epoch < epochsCount; epoch++ {
		cids, err := p.db.FlipCids(epoch)
		if err != nil {
			p.logger.Error("Unable to get epoch flip cids", "epoch", epoch, "err", err)
			continue
		}
		p.logger.Debug("Got epoch flip cids", "e", epoch, "length", len(cids))
		for _, cid := range cids {
			flip, err := p.db.FlipContent(cid)
			if err != nil {
				p.logger.Error("Unable to get flip data", "cid", cid, "err", err)
				continue
			}
			if err = p.exportFlip(epoch, cid, flip); err != nil {
				p.logger.Error("Unable to export flip", "cid", cid, "err", err)
			}
			p.logger.Debug("Saved flip", "cid", cid)
		}
	}
	return nil
}

func (p *Provider) exportFlip(epoch uint64, cid string, content types.FlipContent) error {
	dir := filepath.Join(p.outputDir, fmt.Sprintf("epoch-%d", epoch), fmt.Sprintf("flip-%s", cid))
	if err := checkDir(dir); err != nil {
		return err
	}

	type MetaData struct {
		LeftOrder  []uint16
		RightOrder []uint16
	}
	metadataBytes, err := json.Marshal(MetaData{
		LeftOrder:  content.LeftOrder,
		RightOrder: content.RightOrder,
	})
	if err != nil {
		return err
	}
	if err := writeToFile(metadataBytes, filepath.Join(dir, "metadata.json")); err != nil {
		return err
	}
	for i, pic := range content.Pics {
		if err := writeToFile(pic, filepath.Join(dir, fmt.Sprintf("%d.png", i))); err != nil {
			p.logger.Error("Unable to save flip pic", "cid", cid, "idx", i, "err", err)
			continue
		}
	}

	return nil
}

func writeToFile(data []byte, file string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err = w.Write(data); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func checkDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
