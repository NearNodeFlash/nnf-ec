/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ResourceV1111Identifier - Any additional identifiers for a resource.
type ResourceV1111Identifier struct {

	// The world-wide, persistent name of the resource.
	DurableName string `json:"DurableName,omitempty"`

	DurableNameFormat ResourceV1111DurableNameFormat `json:"DurableNameFormat,omitempty"`
}
