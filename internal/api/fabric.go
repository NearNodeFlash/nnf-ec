package api

import (
	"stash.us.cray.com/rabsw/nnf-ec/internal/events"

	"stash.us.cray.com/rabsw/switchtec-fabric/pkg/switchtec"
)

// FabricApi - Presents an API into the fabric outside of the fabric manager
type FabricControllerApi interface {
	// Returns the Switchtec Device for a given fabric, switch
	GetSwitchtecDevice(fabricId, switchId string) (*switchtec.Device, error)

	// Returns the PDFID descriptor to manage a Physical or Virtual Device Function
	GetPortPDFID(fabricId, switchId, portId string, functionId uint16) (uint16, error)

	// Takes a PortEvent and converts it to the realtive port index within the Fabric Controller.
	// For example, if the fabric consists of a single switch USP and 4 DSPs labeled 0,1,2,3  then
	// a port event of type DSP with event attributes: <FabricId = 0, SwitchId = 0, PortId = 2>
	// would return 2 as the DSP index is 2 is the second DSP type within the fabric.
	ConvertPortEventToRelativePortIndex(events.PortEvent) (int, error)

	FindDownstreamEndpoint(portId, functionId string) (string, error)
}

type FabricNvmeDeviceApi interface {
}

var FabricController FabricControllerApi

func RegisterFabricController(f FabricControllerApi) {
	FabricController = f
}
