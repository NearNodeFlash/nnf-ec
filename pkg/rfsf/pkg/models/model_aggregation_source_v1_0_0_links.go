/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// AggregationSourceV100Links - The links to other resources that are related to this resource.
type AggregationSourceV100Links struct {

	ConnectionMethod OdataV4IdRef `json:"ConnectionMethod,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// An array links to the resources added to the service through this aggregation source.  It is recommended that this be the minimal number of properties needed to find the resources that would be lost when the aggregation source is deleted.
	ResourcesAccessed []ResourceResource `json:"ResourcesAccessed,omitempty"`

	// The number of items in a collection.
	ResourcesAccessedodataCount int64 `json:"ResourcesAccessed@odata.count,omitempty"`
}
