/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// PcIeSlotsV140PcIeSlot - This type defines information for a PCIe slot.
type PcIeSlotsV140PcIeSlot struct {

	// An indication of whether this PCIe slot supports hotplug.
	HotPluggable bool `json:"HotPluggable,omitempty"`

	// The number of PCIe lanes supported by this slot.
	Lanes int64 `json:"Lanes,omitempty"`

	Links PcIeSlotsV140PcIeLinks `json:"Links,omitempty"`

	Location ResourceLocation `json:"Location,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	PCIeType PCIeDevicePCIeTypes `json:"PCIeType,omitempty"`

	SlotType PCIeSlotsV140SlotTypes `json:"SlotType,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`
}
