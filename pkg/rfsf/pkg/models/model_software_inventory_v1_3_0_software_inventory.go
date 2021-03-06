/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"time"
)

// SoftwareInventoryV130SoftwareInventory - The SoftwareInventory schema contains an inventory of software components.  This can include software components such as BIOS, BMC firmware, firmware for other devices, system drivers, or provider software.
type SoftwareInventoryV130SoftwareInventory struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	Actions SoftwareInventoryV130Actions `json:"Actions,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	// The lowest supported version of this software.
	LowestSupportedVersion string `json:"LowestSupportedVersion,omitempty"`

	// The manufacturer or producer of this software.
	Manufacturer string `json:"Manufacturer,omitempty"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// The IDs of the Resources associated with this software inventory item.
	RelatedItem []OdataV4IdRef `json:"RelatedItem,omitempty"`

	// The number of items in a collection.
	RelatedItemodataCount int64 `json:"RelatedItem@odata.count,omitempty"`

	// The release date of this software.
	ReleaseDate *time.Time `json:"ReleaseDate,omitempty"`

	// The implementation-specific label that identifies this software.
	SoftwareId string `json:"SoftwareId,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`

	// The list of UEFI device paths of the components associated with this software inventory item.
	UefiDevicePaths []string `json:"UefiDevicePaths,omitempty"`

	// An indication of whether the Update Service can update this software.
	Updateable bool `json:"Updateable,omitempty"`

	// The version of this software.
	Version string `json:"Version,omitempty"`

	// Indicates if the software is write-protected.
	WriteProtected bool `json:"WriteProtected,omitempty"`
}
