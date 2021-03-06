/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// DriveV1110Links - The links to other resources that are related to this resource.
type DriveV1110Links struct {

	Chassis OdataV4IdRef `json:"Chassis,omitempty"`

	// An array of links to the endpoints that connect to this drive.
	Endpoints []OdataV4IdRef `json:"Endpoints,omitempty"`

	// The number of items in a collection.
	EndpointsodataCount int64 `json:"Endpoints@odata.count,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// An array of links to the PCIe functions that the drive produces.
	PCIeFunctions []OdataV4IdRef `json:"PCIeFunctions,omitempty"`

	// The number of items in a collection.
	PCIeFunctionsodataCount int64 `json:"PCIeFunctions@odata.count,omitempty"`

	// An array of links to the storage pools to which this drive belongs.
	StoragePools []OdataV4IdRef `json:"StoragePools,omitempty"`

	// The number of items in a collection.
	StoragePoolsodataCount int64 `json:"StoragePools@odata.count,omitempty"`

	// An array of links to the volumes that this drive either wholly or only partially contains.
	Volumes []OdataV4IdRef `json:"Volumes,omitempty"`

	// The number of items in a collection.
	VolumesodataCount int64 `json:"Volumes@odata.count,omitempty"`
}
