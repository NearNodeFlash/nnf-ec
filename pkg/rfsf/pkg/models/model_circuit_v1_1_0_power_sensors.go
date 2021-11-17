/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CircuitV110PowerSensors - This property contains the power sensors.
type CircuitV110PowerSensors struct {

	Line1ToLine2 SensorSensorPowerExcerpt `json:"Line1ToLine2,omitempty"`

	Line1ToNeutral SensorSensorPowerExcerpt `json:"Line1ToNeutral,omitempty"`

	Line2ToLine3 SensorSensorPowerExcerpt `json:"Line2ToLine3,omitempty"`

	Line2ToNeutral SensorSensorPowerExcerpt `json:"Line2ToNeutral,omitempty"`

	Line3ToLine1 SensorSensorPowerExcerpt `json:"Line3ToLine1,omitempty"`

	Line3ToNeutral SensorSensorPowerExcerpt `json:"Line3ToNeutral,omitempty"`
}