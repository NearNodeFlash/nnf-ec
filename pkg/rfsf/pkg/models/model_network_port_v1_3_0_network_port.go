/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// NetworkPortV130NetworkPort - The NetworkPort schema describes a network port, which is a discrete physical port that can connect to a network.
type NetworkPortV130NetworkPort struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	Actions NetworkPortV130Actions `json:"Actions,omitempty"`

	ActiveLinkTechnology NetworkPortV130LinkNetworkTechnology `json:"ActiveLinkTechnology,omitempty"`

	// An array of configured MAC or WWN network addresses that are associated with this network port, including the programmed address of the lowest numbered network device function, the configured but not active address, if applicable, the address for hardware port teaming, or other network addresses.
	AssociatedNetworkAddresses []string `json:"AssociatedNetworkAddresses,omitempty"`

	// Network port current link speed.
	CurrentLinkSpeedMbps int64 `json:"CurrentLinkSpeedMbps,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	// An indication of whether IEEE 802.3az Energy-Efficient Ethernet (EEE) is enabled for this network port.
	EEEEnabled bool `json:"EEEEnabled,omitempty"`

	// The FC Fabric Name provided by the switch.
	FCFabricName string `json:"FCFabricName,omitempty"`

	FCPortConnectionType NetworkPortV130PortConnectionType `json:"FCPortConnectionType,omitempty"`

	FlowControlConfiguration NetworkPortV130FlowControl `json:"FlowControlConfiguration,omitempty"`

	FlowControlStatus NetworkPortV130FlowControl `json:"FlowControlStatus,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	LinkStatus NetworkPortV130LinkStatus `json:"LinkStatus,omitempty"`

	// The maximum frame size supported by the port.
	MaxFrameSize int64 `json:"MaxFrameSize,omitempty"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// An array of maximum bandwidth allocation percentages for the network device functions associated with this port.
	NetDevFuncMaxBWAlloc []NetworkPortV130NetDevFuncMaxBwAlloc `json:"NetDevFuncMaxBWAlloc,omitempty"`

	// An array of minimum bandwidth allocation percentages for the network device functions associated with this port.
	NetDevFuncMinBWAlloc []NetworkPortV130NetDevFuncMinBwAlloc `json:"NetDevFuncMinBWAlloc,omitempty"`

	// The number of ports not on this adapter that this port has discovered.
	NumberDiscoveredRemotePorts int64 `json:"NumberDiscoveredRemotePorts,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// The physical port number label for this port.
	PhysicalPortNumber string `json:"PhysicalPortNumber,omitempty"`

	// The largest maximum transmission unit (MTU) that can be configured for this network port.
	PortMaximumMTU int64 `json:"PortMaximumMTU,omitempty"`

	// An indication of whether the port has detected enough signal on enough lanes to establish a link.
	SignalDetected bool `json:"SignalDetected,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`

	// The set of Ethernet capabilities that this port supports.
	SupportedEthernetCapabilities []NetworkPortV130SupportedEthernetCapabilities `json:"SupportedEthernetCapabilities,omitempty"`

	// The link capabilities of this port.
	SupportedLinkCapabilities []NetworkPortV130SupportedLinkCapabilities `json:"SupportedLinkCapabilities,omitempty"`

	// The vendor Identification for this port.
	VendorId string `json:"VendorId,omitempty"`

	// An indication of whether Wake on LAN (WoL) is enabled for this network port.
	WakeOnLANEnabled bool `json:"WakeOnLANEnabled,omitempty"`
}
