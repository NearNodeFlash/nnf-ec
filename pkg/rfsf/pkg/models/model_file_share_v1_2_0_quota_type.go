/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi
// FileShareV120QuotaType : Indicates whether quotas are enabled and enforced by this file share. A value of Soft means that quotas are enabled but not enforced, and a value of Hard means that quotas are enabled and enforced.
type FileShareV120QuotaType string

// List of FileShare_v1_2_0_QuotaType
const (
	SOFT_FSV120QT FileShareV120QuotaType = "Soft"
	HARD_FSV120QT FileShareV120QuotaType = "Hard"
)
