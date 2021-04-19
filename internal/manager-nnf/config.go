package nnf

import (
	_ "embed"

	"gopkg.in/yaml.v2"
)

//go:embed config.yaml
var configFile []byte

type AllocationConfig struct {
	// This is the default allocation policy for the NNF controller. An allocation policy
	// defines the way in which underlying storage is allocated when a user requests storage from
	// the NNF Controller. Valid values are "spares", "global", "switch-local", or "compute-local"
	// with the default being "spares". For more information see allocation_policy.go
	Policy string `yaml:"policy,omitempty"`

	// The Standard defines the level at which the allocation policy should function.
	// Valid values are "strict" or "relaxed", with the default being "strict". See allocation_policy.go
	Standard string `yaml:"standard,omitempty"`
}

type ServerConfig struct {
	Count         int `yaml:"count"`
	StartingIndex int `yaml:"startingIndex"`
}

type ConfigFile struct {
	Version  string
	Metadata struct {
		Name string
	}

	Id string

	ServerConfig     ServerConfig     `yaml:"serverConfig"`
	AllocationConfig AllocationConfig `yaml:"allocationConfig"`
}

func loadConfig() (*ConfigFile, error) {
	var config = new(ConfigFile)
	if err := yaml.Unmarshal(configFile, config); err != nil {
		return config, err
	}

	return config, nil
}