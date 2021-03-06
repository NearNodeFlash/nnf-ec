/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type PowerV161PowerLimitException string

// List of Power_v1_6_1_PowerLimitException
const (
	NO_ACTION_PV161PLE PowerV161PowerLimitException = "NoAction"
	HARD_POWER_OFF_PV161PLE PowerV161PowerLimitException = "HardPowerOff"
	LOG_EVENT_ONLY_PV161PLE PowerV161PowerLimitException = "LogEventOnly"
	OEM_PV161PLE PowerV161PowerLimitException = "Oem"
)
