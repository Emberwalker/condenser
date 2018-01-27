package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// Config represents the on-disk config file.
type Config struct {
	APIKeys    []APIKey `json:"api_keys"`
	RedisConn  string   `json:"redis_conn"`
	CodeLength int      `json:"codelen"`
	ServerURL  string   `json:"server_url"`
}

var (
	// LoadedConfig represents the config as loaded from the disk at startup.
	_config *Config
	// DefaultConfig represents the out-of-box configuration.
	DefaultConfig = Config{
		APIKeys:    []APIKey{},
		RedisConn:  "127.0.0.1:6379",
		CodeLength: 6,
		ServerURL:  "http://0.0.0.0:8000",
	}
)

func getConfig() *Config {
	if _config != nil {
		return _config
	}

	_config = loadConfigFile()
	return _config
}

func loadConfigFile() *Config {
	bytes, err := ioutil.ReadFile("conf.json")
	if err != nil {
		panic("No config file found.")
		//return &DefaultConfig
	}
	confCopy := DefaultConfig
	conf := &confCopy
	err = json.Unmarshal(bytes, conf)
	if err != nil {
		return &DefaultConfig
	}
	conf.RedisConn = strings.Replace(conf.RedisConn, "redis://", "", 1)
	return conf
}
