/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fabric

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
)

type MockSwitchtecController struct {
	devices     []MockSwitchtecDevice
	globalPdfid int
}

type MockSwitchtecDevice struct {
	ctrl      *MockSwitchtecController
	createdAt time.Time

	id     int
	path   string
	open   bool
	exists bool

	config *SwitchConfig
	ports  []MockSwitchtecPort
}

type MockSwitchtecPort struct {
	id        int
	functions []switchtec.DumpEpPortAttachedDeviceFunction

	pdfid uint16

	bindings []*switchtec.DumpEpPortAttachedDeviceFunction

	config *PortConfig
}

func NewMockSwitchtecController() SwitchtecControllerInterface {

	config, err := loadConfig()
	if err != nil {
		return nil
	}

	ctrl := &MockSwitchtecController{
		globalPdfid: rand.Intn(0xFF) << 8,
	}

	ctrl.devices = make([]MockSwitchtecDevice, len(config.Switches))
	for switchIdx, switchConfig := range config.Switches {

		devId, err := strconv.Atoi(switchConfig.Id)
		if err != nil {
			return nil
		}

		ctrl.devices[switchIdx] = MockSwitchtecDevice{
			ctrl:      ctrl,
			id:        devId,
			path:      "", // Path becomes valid when opened
			config:    &config.Switches[switchIdx],
			open:      false,
			exists:    true,
			createdAt: time.Now(),
		}

		dev := &ctrl.devices[switchIdx]

		dev.ports = make([]MockSwitchtecPort, len(switchConfig.Ports))
		for portIdx, portConfig := range switchConfig.Ports {

			dev.ports[portIdx] = MockSwitchtecPort{
				id:     portIdx,
				pdfid:  uint16(ctrl.allocateNewPDFID()),
				config: &switchConfig.Ports[portIdx],
			}
			port := &dev.ports[portIdx]

			switch portConfig.getPortType() {
			case sf.INTERSWITCH_PORT_PV130PT:
				continue
			case sf.MANAGEMENT_PORT_PV130PT:
				port.bindings = make([]*switchtec.DumpEpPortAttachedDeviceFunction, switchConfig.DownstreamPortCount)
			case sf.UPSTREAM_PORT_PV130PT:
				port.bindings = make([]*switchtec.DumpEpPortAttachedDeviceFunction, config.DownstreamPortCount)
			case sf.DOWNSTREAM_PORT_PV130PT:
				{
					port.functions = make([]switchtec.DumpEpPortAttachedDeviceFunction, 1 /* PF */ +1 /*MGMT*/ +config.UpstreamPortCount)

					isPF := func(idx int) uint8 {
						if idx == 0 {
							return 1
						}
						return 0
					}

					for idx := range port.functions {

						f := switchtec.DumpEpPortAttachedDeviceFunction{
							FunctionID:     uint16(idx),
							PDFID:          port.pdfid + uint16(idx),
							SRIOVCapPF:     isPF(idx),
							VFNum:          uint8(idx - 1),
							Bound:          0,
							BoundPAXID:     0,
							BoundHVDPhyPID: 0,
							BoundHVDLogPID: 0,
						}

						port.functions[idx] = f
					}
				}
			default:
				panic("Unhandled port type")
			}

		}
	}

	return ctrl
}

func (c *MockSwitchtecController) Open(path string) (SwitchtecDeviceInterface, error) {
	for deviceIdx := range c.devices {
		device := &c.devices[deviceIdx]
		if device.path == "" || device.path == path {
			device.path = path

			if !device.exists {
				return nil, os.ErrNotExist
			}

			return device, nil
		}
	}
	return nil, fmt.Errorf("Device %s not found", path)
}

func (c *MockSwitchtecController) SetSwitchNotExists(switchIdx int) {
	c.devices[switchIdx].exists = false

}

func (c *MockSwitchtecController) allocateNewPDFID() int {
	pdfid := c.globalPdfid
	c.globalPdfid += 0x100
	return pdfid
}

func (*MockSwitchtecDevice) Device() *switchtec.Device {
	panic("Mock Switchtec Device does not support Device getter")
}

func (*MockSwitchtecDevice) Path() *string {
	panic("Mock Switchtec Device does not support Path getter")
}

func (d *MockSwitchtecDevice) Close() {
	d.path = ""
	d.id = -1
}

func (d *MockSwitchtecDevice) Identify() (int32, error) {
	return int32(d.id), nil
}

func (d *MockSwitchtecDevice) GetFirmwareVersion() (string, error) {
	return "MockFirmware", nil
}

func (d *MockSwitchtecDevice) GetModel() (string, error) {
	return "MockModel", nil
}

func (d *MockSwitchtecDevice) GetManufacturer() (string, error) {
	return "MockMfg", nil
}

func (d *MockSwitchtecDevice) GetSerialNumber() (string, error) {
	return "MockSerialNumber", nil
}

var firstUpstreamPortDisabled bool   // For test, record the first upstream port as being down.
var firstDownstreamPortDisabled bool // For test, record the first downstream port as being down.

func (d *MockSwitchtecDevice) GetPortStatus() ([]switchtec.PortLinkStat, error) {
	stats := make([]switchtec.PortLinkStat, len(d.ports))

	for idx := range stats {
		p := d.ports[idx]
		stats[idx] = switchtec.PortLinkStat{
			PhysPortId: uint8(p.config.Port),

			CfgLinkWidth: uint8(p.config.Width),
			NegLinkWidth: uint8(p.config.Width),

			LinkUp:    true,
			LinkGen:   4,
			LinkState: switchtec.PortLinkState_L0,

			CurLinkRateGBps: switchtec.GetDataRateGBps(4) * float64(p.config.Width),
		}
	}

	return stats, nil
}

func (d *MockSwitchtecDevice) GetPortMetrics() (PortMetrics, error) {

	counters := make(PortMetrics, len(d.ports))

	for idx, p := range d.ports {

		fakeTxRxBytes := rand.Uint64()

		counters[idx] = PortMetric{
			PhysPortId: uint8(p.config.Port),
			BandwidthCounter: switchtec.BandwidthCounter{
				TimeInMicroseconds: uint64(time.Now().Sub(d.createdAt).Microseconds()),
				Egress: switchtec.BandwidthCounterDirectional{
					Posted: fakeTxRxBytes, Completion: 0, NonPosted: 0,
				},
				Ingress: switchtec.BandwidthCounterDirectional{
					Posted: fakeTxRxBytes, Completion: 0, NonPosted: 0,
				},
			},
		}
	}

	return counters, nil
}

func (d *MockSwitchtecDevice) GetEvents() ([]switchtec.GfmsEvent, error) {
	return make([]switchtec.GfmsEvent, 0), nil
}

func (d *MockSwitchtecDevice) EnumerateEndpoint(physPortId uint8, handlerFunc func(epPort *switchtec.DumpEpPortDevice) error) error {

	for _, port := range d.ports {
		if uint8(port.config.Port) == physPortId {
			epPort := switchtec.DumpEpPortDevice{
				Ep: switchtec.DumpEpPortEp{
					Functions: port.functions,
				},
			}

			return handlerFunc(&epPort)
		}
	}

	return nil
}

func (d *MockSwitchtecDevice) ResetEndpoint(pdfid uint16) error {
	return nil
}

func (d *MockSwitchtecDevice) Bind(hostPhysPortId, hostLogPortId uint8, pdfid uint16) error {

	bindPort := func(hostPort *MockSwitchtecPort, hostLogPortId uint8, pdfid uint16) error {
		for deviceIdx := range d.ctrl.devices {
			for portIdx := range d.ctrl.devices[deviceIdx].ports {
				port := &d.ctrl.devices[deviceIdx].ports[portIdx]

				if port.pdfid == pdfid&0xFF00 {
					for functionIdx := range port.functions {
						function := &port.functions[functionIdx]

						if function.PDFID == pdfid {
							if function.Bound != 0 {
								return fmt.Errorf("Device %#04x already bound", pdfid)
							}

							function.Bound = 1
							function.BoundHVDPhyPID = hostPhysPortId
							function.BoundHVDLogPID = hostLogPortId
							function.BoundPAXID = uint8(d.id)

							hostPort.bindings[hostLogPortId] = function

							return nil
						}
					}
				}
			}
		}

		return fmt.Errorf("PDFID %#04x Not Found", pdfid)
	}

	for portIdx, port := range d.ports {
		if port.config.Port == int(hostPhysPortId) {
			if port.bindings[hostLogPortId] != nil {
				return fmt.Errorf("host Port %d Logical Port ID %d already bound", hostPhysPortId, hostLogPortId)
			}

			return bindPort(&d.ports[portIdx], hostLogPortId, pdfid)
		}
	}

	return fmt.Errorf("host Physical Port ID %d not found", hostPhysPortId)
}
