/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// StoragePoolV150EndGrpLifetime - This contains properties the Endurance Group Lifetime attributes.
type StoragePoolV150EndGrpLifetime struct {

	// The property contains the total number of data units read from this endurance group.
	DataUnitsRead int64 `json:"DataUnitsRead,omitempty"`

	// The property contains the total number of data units written from this endurance group.
	DataUnitsWritten int64 `json:"DataUnitsWritten,omitempty"`

	// This property contains an estimate of the total number of data bytes that may be written to the Endurance Group over the lifetime of the Endurance Group assuming a write amplication of 1.
	EnduranceEstimate int64 `json:"EnduranceEstimate,omitempty"`

	// This property contains the number of error information log entries over the life of the controller for the endurance group.
	ErrorInformationLogEntryCount int64 `json:"ErrorInformationLogEntryCount,omitempty"`

	// This property contains the number of read commands completed by all controllers in the NVM subsystem for the Endurance Group.
	HostReadCommandCount int64 `json:"HostReadCommandCount,omitempty"`

	// This property contains the number of write commands completed by all controllers in the NVM subsystem for the Endurance Group.
	HostWriteCommandCount int64 `json:"HostWriteCommandCount,omitempty"`

	// This property contains the number of occurences where the controller detected an unrecovered data integrity error for the Endurance Group.
	MediaAndDataIntegrityErrorCount int64 `json:"MediaAndDataIntegrityErrorCount,omitempty"`

	// The property contains the total number of data units written from this endurance group.
	MediaUnitsWritten int64 `json:"MediaUnitsWritten,omitempty"`

	// A vendor-specific estimate of the percent life used for the endurance group based on the actual usage and the manufacturer prediction of NVM life.
	PercentUsed int64 `json:"PercentUsed,omitempty"`
}
