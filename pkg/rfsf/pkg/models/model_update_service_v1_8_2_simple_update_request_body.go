/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// UpdateServiceV182SimpleUpdateRequestBody - This action updates software components.
type UpdateServiceV182SimpleUpdateRequestBody struct {

	// The URI of the software image to install.
	ImageURI string `json:"ImageURI"`

	// The password to access the URI specified by the ImageURI parameter.
	Password string `json:"Password,omitempty"`

	// An array of URIs that indicate where to apply the update image.
	Targets []string `json:"Targets,omitempty"`

	TransferProtocol UpdateServiceV182TransferProtocolType `json:"TransferProtocol,omitempty"`

	// The user name to access the URI specified by the ImageURI parameter.
	Username string `json:"Username,omitempty"`
}
