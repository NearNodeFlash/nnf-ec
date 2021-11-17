/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// IoConnectivityLineOfServiceV121IoConnectivityLineOfService - A service option within the IO Connectivity line of service.
type IoConnectivityLineOfServiceV121IoConnectivityLineOfService struct {

	// SupportedAccessProtocols.
	AccessProtocols []ProtocolProtocol `json:"AccessProtocols,omitempty"`

	Actions IoConnectivityLineOfServiceV121Actions `json:"Actions,omitempty"`

	// The maximum Bandwidth in bytes per second that a connection can support.
	MaxBytesPerSecond int64 `json:"MaxBytesPerSecond,omitempty"`

	// The maximum supported IOs per second that the connection will support for the selected access protocol.
	MaxIOPS int64 `json:"MaxIOPS,omitempty"`
}