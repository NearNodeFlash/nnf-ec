/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// MemoryV1100RegionSet - Memory region information within a memory device.
type MemoryV1100RegionSet struct {

	MemoryClassification MemoryV1100MemoryClassification `json:"MemoryClassification,omitempty"`

	// Offset within the memory that corresponds to the start of this memory region in mebibytes (MiB).
	OffsetMiB int64 `json:"OffsetMiB,omitempty"`

	// An indication of whether the passphrase is enabled for this region.
	PassphraseEnabled bool `json:"PassphraseEnabled,omitempty"`

	// An indication of whether the state of the passphrase for this region is enabled.
	PassphraseState bool `json:"PassphraseState,omitempty"`

	// Unique region ID representing a specific region within the memory device.
	RegionId string `json:"RegionId,omitempty"`

	// Size of this memory region in mebibytes (MiB).
	SizeMiB int64 `json:"SizeMiB,omitempty"`
}
