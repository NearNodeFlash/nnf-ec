/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ComputerSystemV1130PowerState string

// List of ComputerSystem_v1_13_0_PowerState
const (
	ON_CSV1130PST ComputerSystemV1130PowerState = "On"
	OFF_CSV1130PST ComputerSystemV1130PowerState = "Off"
	POWERING_ON_CSV1130PST ComputerSystemV1130PowerState = "PoweringOn"
	POWERING_OFF_CSV1130PST ComputerSystemV1130PowerState = "PoweringOff"
)
