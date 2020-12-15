package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Postgres                   PostgresConfig
	RuntimeMigration           RuntimeMigrationConfig
	Migration                  Migration
	NodeConfigFile             string
	WordsFile                  string
	Verbosity                  int
	NodeVerbosity              int
	RestoreInitially           bool
	PerformanceMonitor         PerformanceMonitorConfig
	FlipContentLoader          FlipContentLoaderConfig
	CommitteeRewardBlocksCount int
	MiningRewards              bool
	Enabled                    *bool
}

type PerformanceMonitorConfig struct {
	Enabled     bool
	BlocksToLog int
}

type PostgresConfig struct {
	ConnStr    string
	ScriptsDir string
}

type RuntimeMigrationConfig struct {
	Enabled  bool
	Postgres PostgresConfig
}

type Migration struct {
	Enabled    bool
	OldSchema  string
	Height     uint64
	ScriptsDir string
}

type FlipContentLoaderConfig struct {
	BatchSize        int
	AttemptsLimit    int
	RetryIntervalMin int
}

func LoadConfig(configPath string) *Config {
	if _, err := os.Stat(configPath); err != nil {
		panic(errors.Errorf("Config file cannot be found, path: %v", configPath))
	}
	if jsonFile, err := os.Open(configPath); err != nil {
		panic(errors.Errorf("Config file cannot be opened, path: %v", configPath))
	} else {
		conf := newDefaultConfig()
		byteValue, _ := ioutil.ReadAll(jsonFile)
		err := json.Unmarshal(byteValue, conf)
		if err != nil {
			panic(errors.Errorf("Cannot parse JSON config, path: %v", configPath))
		}
		return conf
	}
}

func newDefaultConfig() *Config {
	return &Config{
		NodeConfigFile: filepath.Join("conf", "node.json"),
		Postgres: PostgresConfig{
			ScriptsDir: filepath.Join("resources", "scripts", "indexer"),
		},
		Verbosity:     3,
		NodeVerbosity: 0,
		FlipContentLoader: FlipContentLoaderConfig{
			BatchSize:        50,
			AttemptsLimit:    5,
			RetryIntervalMin: 10,
		},
		CommitteeRewardBlocksCount: 1000,
	}
}
