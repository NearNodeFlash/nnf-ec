/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// SessionServiceV117SessionService - The SessionService schema describes the session service and its properties, with links to the actual list of sessions.
type SessionServiceV117SessionService struct {

	// The OData description of a payload.
	OdataContext string `json:"@odata.context,omitempty"`

	// The current ETag of the resource.
	OdataEtag string `json:"@odata.etag,omitempty"`

	// The unique identifier for a resource.
	OdataId string `json:"@odata.id"`

	// The type of a resource.
	OdataType string `json:"@odata.type"`

	Actions SessionServiceV117Actions `json:"Actions,omitempty"`

	// The description of this resource.  Used for commonality in the schema definitions.
	Description string `json:"Description,omitempty"`

	// The identifier that uniquely identifies the resource within the collection of similar resources.
	Id string `json:"Id"`

	// The name of the resource or array member.
	Name string `json:"Name"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// An indication of whether this service is enabled.  If `true`, this service is enabled.  If `false`, it is disabled, and new sessions cannot be created, old sessions cannot be deleted, and established sessions can continue operating.
	ServiceEnabled bool `json:"ServiceEnabled,omitempty"`

	// The number of seconds of inactivity that a session can have before the session service closes the session due to inactivity.
	SessionTimeout int64 `json:"SessionTimeout,omitempty"`

	Sessions OdataV4IdRef `json:"Sessions,omitempty"`

	Status ResourceStatus `json:"Status,omitempty"`
}
