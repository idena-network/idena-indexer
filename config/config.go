package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Postgres              PostgresConfig
	FlipMigrationPostgres *PostgresConfig
	Migration             *Migration
	NodeConfigFile        string
	Verbosity             int
	NodeVerbosity         int
	GenesisBlockHeight    int
}

type PostgresConfig struct {
	ConnStr    string
	ScriptsDir string
}

type Migration struct {
	OldSchema  string
	Height     uint64
	ScriptsDir string
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
		Verbosity:          3,
		NodeVerbosity:      0,
		GenesisBlockHeight: 1,
	}
}
