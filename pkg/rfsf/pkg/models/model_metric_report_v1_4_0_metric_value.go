/*
 * Swordfish API
 *
 * This contains the definition of the Swordfish extensions to a Redfish service.
 *
 * API version: v1.2.c
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"time"
)

// MetricReportV140MetricValue - Properties that capture a metric value and other associated information.
type MetricReportV140MetricValue struct {

	MetricDefinition OdataV4IdRef `json:"MetricDefinition,omitempty"`

	// The metric definitions identifier for this metric.
	MetricId string `json:"MetricId,omitempty"`

	// The URI for the property from which this metric is derived.
	MetricProperty string `json:"MetricProperty,omitempty"`

	// The metric value, as a string.
	MetricValue string `json:"MetricValue,omitempty"`

	// The OEM extension.
	Oem map[string]interface{} `json:"Oem,omitempty"`

	// The date and time when the metric is obtained.  A management application can establish a time series of metric data by retrieving the instances of metric value and sorting them according to their timestamp.
	Timestamp *time.Time `json:"Timestamp,omitempty"`
}
