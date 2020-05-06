package config

import (
	"encoding/json"
	"github.com/idena-network/idena-indexer/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"time"
)

type DynamicConfig struct {
	DumpCid string
}

type DynamicConfigHolder struct {
	filePath string
	modTime  *time.Time
	config   *DynamicConfig
	logger   log.Logger
}

func (holder *DynamicConfigHolder) GetConfig() *DynamicConfig {
	return holder.config
}

func NewDynamicConfigHolder(filePath string, logger log.Logger) *DynamicConfigHolder {
	holder := &DynamicConfigHolder{
		filePath: filePath,
		logger:   logger,
	}
	if ok, err := holder.updateIfNeeded(); err != nil {
		if ok {
			logger.Warn(err.Error())
		} else {
			panic(err)
		}
	}
	go holder.updateLoop()
	return holder
}

func readDynamicConfig(configPath string) (*DynamicConfig, error) {
	if jsonFile, err := os.Open(configPath); err != nil {
		return nil, errors.Errorf("Config file cannot be opened, path: %v", configPath)
	} else {
		conf := &DynamicConfig{}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err := json.Unmarshal(byteValue, conf)
		if err != nil {
			return nil, errors.Errorf("Cannot parse JSON config, path: %v", configPath)
		}
		return conf, nil
	}
}

func (holder *DynamicConfigHolder) updateLoop() {
	for {
		time.Sleep(time.Minute)
		ok, err := holder.updateIfNeeded()
		if err != nil {
			holder.logger.Warn(err.Error())
			if !ok {
				continue
			}
		}
		if ok {
			holder.logger.Info("Dynamic config updated")
		}
	}
}

func (holder *DynamicConfigHolder) updateIfNeeded() (bool, error) {
	fileInfo, err := os.Stat(holder.filePath)
	if err != nil {
		holder.config = &DynamicConfig{}
		return true, errors.New("unable to find dynamic config file")
	}
	if holder.modTime != nil && !fileInfo.ModTime().After(*holder.modTime) {
		return false, nil
	}
	conf, err := readDynamicConfig(holder.filePath)
	if err != nil {
		return false, err
	}
	modeTime := fileInfo.ModTime()
	holder.modTime = &modeTime
	holder.config = conf
	return true, nil
}
