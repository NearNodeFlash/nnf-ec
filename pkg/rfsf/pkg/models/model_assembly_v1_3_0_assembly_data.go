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

type AssemblyV130AssemblyData struct {

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	Actions AssemblyV130AssemblyDataActions `json:"Actions,omitempty"`

	// The URI at which to access an image of the assembly information.
	BinaryDataURI string `json:"BinaryDataURI,omitempty"`

	// The description of the assembly.
	Description string `json:"Description,omitempty"`

	// The engineering change level of the assembly.
	EngineeringChangeLevel string `json:"EngineeringChangeLevel,omitempty"`

	Location ResourceLocation `json:"Location,omitempty"`

	// An indicator allowing an operator to physically locate this resource.
	LocationIndicatorActive bool `json:"LocationIndicatorActive,omitempty"`

	// The identifier for the member within the collection.
	MemberId string `json:"MemberId"`

	// The model number of the assembly.
	Model string `json:"Model,omitempty"`

	// The name of the assembly.
	Name string `json:"Name,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// The part number of the assembly.
	PartNumber string `json:"PartNumber,omitempty"`

	PhysicalContext PhysicalContextPhysicalContext `json:"PhysicalContext,omitempty"`

	// The producer or manufacturer of the assembly.
	Producer string `json:"Producer,omitempty"`

	// The production date of the assembly.
	ProductionDate *time.Time `json:"ProductionDate,omitempty"`

	// The SKU of the assembly.
	SKU string `json:"SKU,omitempty"`

	// The serial number of the assembly.
	SerialNumber string `json:"SerialNumber,omitempty"`

	// The spare part number of the assembly.
	SparePartNumber string `json:"SparePartNumber,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`

	// The vendor of the assembly.
	Vendor string `json:"Vendor,omitempty"`

	// The hardware version of the assembly.
	Version string `json:"Version,omitempty"`
}
