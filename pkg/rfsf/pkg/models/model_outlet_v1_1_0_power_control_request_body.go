/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// OutletV110PowerControlRequestBody - This action turns the outlet on or off.
type OutletV110PowerControlRequestBody struct {

	PowerState OutletPowerState `json:"PowerState,omitempty"`
}
