package fabric

import (
	//"embed" look into embedded go files
	"fmt"

	"gopkg.in/yaml.v2"

	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

//go:embed config.yaml
//var configFile embed.FS

const configFile = `
version: v1
metadata:
  name: Rabbit Fabric Configuration
switches:
  - id: 0
    metadata:
      name: PAX 0
    ports:
      - name: Interswitch Fabric
        type: InterswitchPort
        port: 0
        width: 16
      - name: Rabbit
        type: ManagementPort
        port: 24
        width: 16
      - name: Compute 0
        type: UpstreamPort
        port: 32
        width: 4
      - name: Compute 1
        type: UpstreamPort
        port: 34
        width: 4
      - name: Compute 2
        type: UpstreamPort
        port: 36
        width: 4
      - name: Compute 3
        type: UpstreamPort
        port: 38
        width: 4
      - name: Compute 4
        type: UpstreamPort
        port: 40
        width: 4
      - name: Compute 5
        type: UpstreamPort
        port: 42
        width: 4
      - name: Compute 6
        type: UpstreamPort
        port: 44
        width: 4
      - name: Compute 7
        type: UpstreamPort
        port: 46
        width: 4
      - name: SSD 0
        type: DownstreamPort
        port: 8
        width: 4
      - name: SSD 1
        type: DownstreamPort
        port: 10
        width: 4
      - name: SSD 2
        type: DownstreamPort
        port: 12
        width: 4
      - name: SSD 3
        type: DownstreamPort
        port: 14
        width: 4
      - name: SSD 4
        type: DownstreamPort
        port: 16
        width: 4
      - name: SSD 5
        type: DownstreamPort
        port: 18
        width: 4
      - name: SSD 6
        type: DownstreamPort
        port: 20
        width: 4
      - name: SSD 7
        type: DownstreamPort
        port: 22
        width: 4
      - name: SSD 8
        type: DownstreamPort
        port: 48
        width: 4

  - id: 1
    metadata:
      name: PAX 1
    ports:
      - name: Interswitch Fabric
        type: InterswitchPort
        port: 0
        width: 16
      - name: Rabbit
        type: ManagementPort
        port: 24
        width: 16
      - name: Compute 8
        type: UpstreamPort
        port: 32
        width: 4
      - name: Compute 9
        type: UpstreamPort
        port: 34
        width: 4
      - name: Compute 10
        type: UpstreamPort
        port: 36
        width: 4
      - name: Compute 11
        type: UpstreamPort
        port: 38
        width: 4
      - name: Compute 12
        type: UpstreamPort
        port: 40
        width: 4
      - name: Compute 13
        type: UpstreamPort
        port: 42
        width: 4
      - name: Compute 14
        type: UpstreamPort
        port: 44
        width: 4
      - name: Compute 15
        type: UpstreamPort
        port: 46
        width: 4
      - name: SSD 9
        type: DownstreamPort
        port: 8
        width: 4
      - name: SSD 10
        type: DownstreamPort
        port: 10
        width: 4
      - name: SSD 11
        type: DownstreamPort
        port: 12
        width: 4
      - name: SSD 12
        type: DownstreamPort
        port: 14
        width: 4
      - name: SSD 13
        type: DownstreamPort
        port: 16
        width: 4
      - name: SSD 14
        type: DownstreamPort
        port: 18
        width: 4
      - name: SSD 15
        type: DownstreamPort
        port: 20
        width: 4
      - name: SSD 16
        type: DownstreamPort
        port: 22
        width: 4
      - name: SSD 17
        type: DownstreamPort
        port: 48
        width: 4
`

type ConfigFile struct {
	Version  string
	Metadata struct {
		Name string
	}
	Switches []SwitchConfig

	ManagementPortCount int
	UpstreamPortCount   int
	DownstreamPortCount int
}

type SwitchConfig struct {
	Id       string
	Metadata struct {
		Name string
	}
	Ports []PortConfig

	ManagementPortCount int
	UpstreamPortCount   int
	DownstreamPortCount int
}

type PortConfig struct {
	Id    string
	Name  string
	Type  string
	Port  int
	Width int
}

var Config *ConfigFile

func loadConfig() error {

	/*
		data, err := configFile.ReadFile("config.yaml")
		if err != nil {
			return err
		}
	*/
	data := []byte(configFile)

	Config = new(ConfigFile)
	if err := yaml.Unmarshal(data, Config); err != nil {
		return err
	}

	// For usability we convert the port index to a string - this
	// allows for easier comparisons for functions receiving portId
	// as a string. We also tally the number of each port type
	for switchIdx := range Config.Switches {
		s := &Config.Switches[switchIdx]
		for _, p := range s.Ports {

			switch p.getPortType() {
			case sf.MANAGEMENT_PORT_PV130PT:
				s.ManagementPortCount++
			case sf.UPSTREAM_PORT_PV130PT:
				s.UpstreamPortCount++
			case sf.DOWNSTREAM_PORT_PV130PT:
				s.DownstreamPortCount++
			case sf.INTERSWITCH_PORT_PV130PT:
				continue
			default:
				return fmt.Errorf("Unhandled port type %s", p.Type)
			}
		}

		Config.ManagementPortCount += s.ManagementPortCount
		Config.UpstreamPortCount += s.UpstreamPortCount
		Config.DownstreamPortCount += s.DownstreamPortCount
	}

	// Only a single management endpoint for ALL switches (but unique ports)
	if Config.ManagementPortCount != len(Config.Switches) {
		return fmt.Errorf("Misconfigured Switch Ports: Expected %d Management Ports, Received: %d", len(Config.Switches), Config.ManagementPortCount)
	}

	return nil
}

func (c *ConfigFile) findSwitch(switchId string) (*SwitchConfig, error) {
	for _, s := range c.Switches {
		if s.Id == switchId {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("Switch %s not found", switchId)
}

func (c *ConfigFile) findSwitchPort(switchId string, portId string) (*PortConfig, error) {
	s, err := c.findSwitch(switchId)
	if err != nil {
		return nil, err
	}

	for _, port := range s.Ports {
		if portId == port.Id {
			return &port, nil
		}
	}

	return nil, fmt.Errorf("Switch %s Port %s not found", switchId, portId)
}

func (p *PortConfig) getPortType() sf.PortV130PortType {
	switch p.Type {
	case "InterswitchPort":
		return sf.INTERSWITCH_PORT_PV130PT
	case "UpstreamPort":
		return sf.UPSTREAM_PORT_PV130PT
	case "DownstreamPort":
		return sf.DOWNSTREAM_PORT_PV130PT
	case "ManagementPort":
		return sf.MANAGEMENT_PORT_PV130PT
	default:
		return sf.UNCONFIGURED_PORT_PV130PT
	}
}
