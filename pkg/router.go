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

var (
	GET_METHOD    = strings.ToUpper("Get")
	POST_METHOD   = strings.ToUpper("Post")
	PATCH_METHOD  = strings.ToUpper("Patch")
	DELETE_METHOD = strings.ToUpper("Delete")
)

// Routes -
func (r *DefaultApiRouter) Routes() ec.Routes {
	s := r.servicer
	return ec.Routes{
		{
			Name:        "RedfishV1FabricsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics",
			HandlerFunc: s.RedfishV1FabricsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesSwitchIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesSwitchIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}/Ports",
			HandlerFunc: s.RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Endpoints",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointsEndpointIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Endpoints/{EndpointId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointsEndpointIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointGroupsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/EndpointGroups",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointGroupsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdEndpointGroupsEndpointGroupIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/EndpointGroups/{EndpointGroupId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdEndpointGroupsEndpointGroupIdGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdConnectionsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Connections",
			HandlerFunc: s.RedfishV1FabricsFabricIdConnectionsGet,
		},
		{
			Name:        "RedfishV1FabricsFabricIdConnectionsConnectionIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Fabrics/{FabricId}/Connections/{ConnectionId}",
			HandlerFunc: s.RedfishV1FabricsFabricIdConnectionsConnectionIdGet,
		},
		{
			Name:        "RedfishV1StorageGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage",
			HandlerFunc: s.RedfishV1StorageGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}",
			HandlerFunc: s.RedfishV1StorageStorageIdGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdStoragePoolsGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/StoragePools",
			HandlerFunc: s.RedfishV1StorageStorageIdStoragePoolsGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdStoragePoolsStoragePoolIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/StoragePool/{StoragePoolId}",
			HandlerFunc: s.RedfishV1StorageStorageIdStoragePoolsStoragePoolIdGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdControllersGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Controllers",
			HandlerFunc: s.RedfishV1StorageStorageIdControllersGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdControllersControllerIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Controllers/{ControllerId}",
			HandlerFunc: s.RedfishV1StorageStorageIdControllersControllerIdGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdVolumesGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Volumes",
			HandlerFunc: s.RedfishV1StorageStorageIdVolumesGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdVolumesPost",
			Method:      POST_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Volumes",
			HandlerFunc: s.RedfishV1StorageStorageIdVolumesPost,
		},
		{
			Name:        "RedfishV1StorageStorageIdVolumesVolumeIdGet",
			Method:      GET_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Volumes/{VolumeId}",
			HandlerFunc: s.RedfishV1StorageStorageIdVolumesVolumeIdGet,
		},
		{
			Name:        "RedfishV1StorageStorageIdVolumesVolumeIdDelete",
			Method:      DELETE_METHOD,
			Path:        "/redfish/v1/Storage/{StorageId}/Volumes/{VolumeId}",
			HandlerFunc: s.RedfishV1StorageStorageIdVolumesVolumeIdDelete,
		},
	}
}
