package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	PostgresConnStr string
	NodeConfigFile  string
	Verbosity       int
	NodeVerbosity   int
	ScriptsDir      string
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
		ScriptsDir:     filepath.Join("resources", "scripts", "indexer"),
		Verbosity:      3,
		NodeVerbosity:  0,
	}
}
