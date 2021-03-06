/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RedundancyV125Redundancy - The common redundancy definition and structure used in other Redfish schemas.
type RedundancyV125Redundancy struct {

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	Actions RedundancyV125Actions `json:"Actions,omitempty"`

	// The maximum number of members allowable for this particular redundancy group.
	MaxNumSupported int64 `json:"MaxNumSupported,omitempty"`

	// The identifier for the member within the collection.
	MemberId string `json:"MemberId"`

	// The minimum number of members needed for this group to be redundant.
	MinNumNeeded int64 `json:"MinNumNeeded"`

	Mode RedundancyV125RedundancyMode `json:"Mode"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// An indication of whether redundancy is enabled.
	RedundancyEnabled bool `json:"RedundancyEnabled,omitempty"`

	// The links to components of this redundancy set.
	RedundancySet []OdataV4IdRef `json:"RedundancySet"`

	// The number of items in a collection.
	RedundancySetodataCount int64 `json:"RedundancySet@odata.count,omitempty"`

	Status ResourceStatus `json:"Status"`
}
