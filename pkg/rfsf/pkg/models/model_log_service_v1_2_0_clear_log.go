/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// LogServiceV120ClearLog - The action to clear the log for this Log Service.
type LogServiceV120ClearLog struct {

	// Link to invoke action
	Target string `json:"target,omitempty"`

	// Friendly action name
	Title string `json:"title,omitempty"`
}
