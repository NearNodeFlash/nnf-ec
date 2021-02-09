package nvmenamespace

import (
	//"embed" look into embedded go files

	"gopkg.in/yaml.v2"
)

const configFile = `
version: v1
metadata:
  name: Rabbit
storage:
  controller:
    virtualFunctions: 17
  devices:
    # Switch 0
    '0000:19:00.0', '0000:1a:00.0', '0000:1b:00.0', '0000:1c:00.0', '0000:1d:00.0', '0000:1e:00.0', '0000:1f:00.0', '0000:20:00.0', '0000:21:00.0',
    # Switch 1
    '0000:24:00.0', '0000:25:00.0', '0000:26:00.0', '0000:27:00.0', '0000:28:00.0', '0000:29:00.0', '0000:2a:00.0', '0000:2b:00.0', '0000:2c:00.0',
  ]
`

type ControllerConfig struct {
	VirtualFunctions int
}

type ConfigFile struct {
	Version  string
	Metadata struct {
		Name string
	}
	ControllerConfig ControllerConfig
	Devices          []string `yaml:",flow"`
}

func loadConfig() (*ConfigFile, error) {
	data := []byte(configFile)

	var config = new(ConfigFile)
	if err := yaml.Unmarshal(data, config); err != nil {
		return config, err
	}

	return config, nil
}
