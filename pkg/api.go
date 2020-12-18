package nnf

import (
	"net/http"
)

// DefaultApi defines the interface into NNF; this functions will
// be routed to the servicer over the element controller.
type DefaultApi interface {
	RedfishV1FabricsGet(w http.ResponseWriter, r *http.Request)
	RedfishV1FabricsFabricIdGet(w http.ResponseWriter, r *http.Request)

	/*
		RedfishV1FabricsFabricIdSwitchesGet(w http.ResponseWriter, r *http.Request)
		RedfishV1FabricsFabricIdSwitchesSwitchIdGet(w http.ResponseWriter, r *http.Request)


		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet(w http.ResponseWriter, r *http.Request)
		RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet(w http.ResponseWriter, r *http.Request)

		RedfishV1FabricsFabricIdEndpointsGet(w http.ResponseWriter, r *http.Request)
		RedfishV1FabricsFabricIdEndpointsEndpointIdGet(w http.ResponseWriter, r *http.Request)

		RedfishV1FabricsFabricIdConnectionsGet(w http.ResponseWriter, r *http.Request)
		RedfishV1FabricsFabricIdConnectionsPost(w http.ResponseWriter, r *http.Request)
	*/
}
