/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// FileSystemV122FileSystem - An instance of a hierarchical namespace of files.
type FileSystemV122FileSystem struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	// An array of supported IO access capabilities.
	AccessCapabilities []DataStorageLoSCapabilitiesStorageAccessCapability `json:"AccessCapabilities,omitempty"`

	Actions FileSystemV122Actions `json:"Actions,omitempty"`

	// Block size of the file system in bytes.
	BlockSizeBytes int64 `json:"BlockSizeBytes,omitempty"`

	// The size in bytes of this file system.
	CapacityBytes int64 `json:"CapacityBytes,omitempty"`

	Capacity CapacityV100Capacity `json:"Capacity,omitempty"`

	// An array of capacity sources for the file system.
	CapacitySources []CapacityCapacitySource `json:"CapacitySources,omitempty"`

	// The number of items in a collection.
	CapacitySourcesodataCount int64 `json:"CapacitySources@odata.count,omitempty"`

	// The case of file names is preserved by the file system.
	CasePreserved bool `json:"CasePreserved,omitempty"`

	// Case sensitive file names are supported by the file system.
	CaseSensitive bool `json:"CaseSensitive,omitempty"`

	// An array of the character sets or encodings supported by the file system.
	CharacterCodeSet []FileSystemV122CharacterCodeSet `json:"CharacterCodeSet,omitempty"`

	// A value indicating the minimum file allocation size imposed by the file system.
	ClusterSizeBytes int64 `json:"ClusterSizeBytes,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	ExportedShares OdataV4IdRef `json:"ExportedShares,omitempty"`

	IOStatistics IoStatisticsIoStatistics `json:"IOStatistics,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	// The durable names for this file system.
	Identifiers []ResourceIdentifier `json:"Identifiers,omitempty"`

	// An array of imported file shares.
	ImportedShares []FileSystemImportedShare `json:"ImportedShares,omitempty"`

	Links FileSystemV122Links `json:"Links,omitempty"`

	// An array of low space warning threshold percentages for the file system.
	LowSpaceWarningThresholdPercents []int64 `json:"LowSpaceWarningThresholdPercents,omitempty"`

	// A value indicating the maximum length of a file name within the file system.
	MaxFileNameLengthBytes int64 `json:"MaxFileNameLengthBytes,omitempty"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// Current number of capacity source resources that are available as replacements.
	RecoverableCapacitySourceCount int64 `json:"RecoverableCapacitySourceCount,omitempty"`

	RemainingCapacity CapacityV100Capacity `json:"RemainingCapacity,omitempty"`

	// The percentage of the capacity remaining in the FileSystem.
	RemainingCapacityPercent int64 `json:"RemainingCapacityPercent,omitempty"`

	ReplicaInfo StorageReplicaInfoReplicaInfo `json:"ReplicaInfo,omitempty"`

	// The resources that are target replicas of this source.
	ReplicaTargets []OdataV4IdRef `json:"ReplicaTargets,omitempty"`

	// The number of items in a collection.
	ReplicaTargetsodataCount int64 `json:"ReplicaTargets@odata.count,omitempty"`

	// The storage pool backing this file system
	StoragePool OdataV4IdRef `json:"StoragePool,omitempty"`
}
