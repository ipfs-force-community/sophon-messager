package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
)

func ReadConfig(path string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	if err := toml.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}
	return config, nil
}
