/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CertificateV121RekeyResponse - The response body for the Rekey action.
type CertificateV121RekeyResponse struct {

	// The string for the certificate signing request.
	CSRString string `json:"CSRString"`

	Certificate OdataV4IdRef `json:"Certificate"`
}