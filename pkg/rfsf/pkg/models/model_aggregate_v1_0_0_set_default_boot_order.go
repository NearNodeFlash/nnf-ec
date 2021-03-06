/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// AggregateV100SetDefaultBootOrder - This action is used to restore the boot order to the default state for the computer systems that are members of this aggregate.
type AggregateV100SetDefaultBootOrder struct {

	// Link to invoke action
	Target string `json:"target,omitempty"`

	// Friendly action name
	Title string `json:"title,omitempty"`
}
