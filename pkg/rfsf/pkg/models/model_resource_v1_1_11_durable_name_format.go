/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ResourceV1111DurableNameFormat string

// List of Resource_v1_1_11_DurableNameFormat
const (
	NAA_RV1111DNF ResourceV1111DurableNameFormat = "NAA"
	I_QN_RV1111DNF ResourceV1111DurableNameFormat = "iQN"
	FC_WWN_RV1111DNF ResourceV1111DurableNameFormat = "FC_WWN"
	UUID_RV1111DNF ResourceV1111DurableNameFormat = "UUID"
	EUI_RV1111DNF ResourceV1111DurableNameFormat = "EUI"
)
