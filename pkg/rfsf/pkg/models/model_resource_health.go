/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ResourceHealth string

// List of Resource_Health
const (
	OK_RH ResourceHealth = "OK"
	WARNING_RH ResourceHealth = "Warning"
	CRITICAL_RH ResourceHealth = "Critical"
)
