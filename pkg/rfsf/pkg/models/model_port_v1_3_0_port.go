/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// PortV130Port - The Port schema contains properties that describe a port of a switch, controller, chassis, or any other device that could be connected to another entity.
type PortV130Port struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	Actions PortV130Actions `json:"Actions,omitempty"`

	// The number of active lanes for this interface.
	ActiveWidth int64 `json:"ActiveWidth,omitempty"`

	// The current speed of this port.
	CurrentSpeedGbps float32 `json:"CurrentSpeedGbps,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	Ethernet PortV130EthernetProperties `json:"Ethernet,omitempty"`

	FibreChannel PortV130FibreChannelProperties `json:"FibreChannel,omitempty"`

	GenZ PortV130GenZ `json:"GenZ,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	// An indication of whether the interface is enabled.
	InterfaceEnabled bool `json:"InterfaceEnabled,omitempty"`

	// The link configuration of this port.
	LinkConfiguration []PortV130LinkConfiguration `json:"LinkConfiguration,omitempty"`

	LinkNetworkTechnology PortV130LinkNetworkTechnology `json:"LinkNetworkTechnology,omitempty"`

	LinkState PortV130LinkState `json:"LinkState,omitempty"`

	LinkStatus PortV130LinkStatus `json:"LinkStatus,omitempty"`

	// The number of link state transitions for this interface.
	LinkTransitionIndicator int64 `json:"LinkTransitionIndicator,omitempty"`

	Links PortV130Links `json:"Links,omitempty"`

	Location ResourceLocation `json:"Location,omitempty"`

	// An indicator allowing an operator to physically locate this resource.
	LocationIndicatorActive bool `json:"LocationIndicatorActive,omitempty"`

	// The maximum frame size supported by the port.
	MaxFrameSize int64 `json:"MaxFrameSize,omitempty"`

	// The maximum speed of this port as currently configured.
	MaxSpeedGbps float32 `json:"MaxSpeedGbps,omitempty"`

	Metrics OdataV4IdRef `json:"Metrics,omitempty"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// The label of this port on the physical package for this port.
	PortId string `json:"PortId,omitempty"`

	PortMedium PortV130PortMedium `json:"PortMedium,omitempty"`

	PortProtocol ProtocolProtocol `json:"PortProtocol,omitempty"`

	PortType PortV130PortType `json:"PortType,omitempty"`

	// An indication of whether a signal is detected on this interface.
	SignalDetected bool `json:"SignalDetected,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`

	// The number of lanes, phys, or other physical transport links that this port contains.
	Width int64 `json:"Width,omitempty"`
}
