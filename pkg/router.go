package nnf

import (
	"strings"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
)

// Router contains all the Redfish / Swordfish API calls that are hosted by
// the NNF module. All handler calls are of the form RedfishV1{endpoint}, where
// and endpoint is unique to the RF/SF caller.
//
// Router calls must reflect the same function name as the RF/SF API, as the
// element controller will perform a 1:1 function call based on the RF/SF
// caller's name.

// DefaultApiRouter -
type DefaultApiRouter struct {
	servicer Api
}

// NewDefaultApiRouter -
func NewDefaultApiRouter(s Api) ec.Router {
	return &DefaultApiRouter{servicer: s}
}

// Routes -
func (r *DefaultApiRouter) Routes() ec.Routes {
	s := r.servicer
	return ec.Routes{
		{
			Name:        "RedfishV1FabricsGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics",
			HandlerFunc: s.RedfishV1FabricsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesSwitchIdGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesSwitchIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}/Ports",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointsGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}/Endpoints",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointsEndpointIdGet",
			Method:      strings.ToUpper("Get"),
			Path:        "/redfish/v1/Fabrics/{FabricId}/Endpoints/{EndpointId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointsEndpointIdGet,
		},
	}
}
