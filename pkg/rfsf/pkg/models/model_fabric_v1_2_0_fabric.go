/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// FabricV120Fabric - The Fabric schema represents a simple fabric consisting of one or more switches, zero or more endpoints, and zero or more zones.
type FabricV120Fabric struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	Actions FabricV120Actions `json:"Actions,omitempty"`

	AddressPools OdataV4IdRef `json:"AddressPools,omitempty"`

	Connections OdataV4IdRef `json:"Connections,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	EndpointGroups OdataV4IdRef `json:"EndpointGroups,omitempty"`

	Endpoints OdataV4IdRef `json:"Endpoints,omitempty"`

	FabricType ProtocolProtocol `json:"FabricType,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	Links FabricV120Links `json:"Links,omitempty"`

	// The maximum number of zones the switch can currently configure.
	MaxZones int64 `json:"MaxZones,omitempty"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`

	Switches OdataV4IdRef `json:"Switches,omitempty"`

	Zones OdataV4IdRef `json:"Zones,omitempty"`
}
