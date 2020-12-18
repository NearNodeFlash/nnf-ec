package nnf

import (
	"net/http"

	"github.com/gorilla/mux"
)

// ApiController singleton object used to communicated to the NNF API
var (
	ApiController = newDefaultApiController()
)

// DefaultApiController -
type defaultApiController struct{}

// NewDefaultApiController -
func newDefaultApiController() DefaultApi {
	return &defaultApiController{}
}

// RedfishV1FabricsGet -
func (*defaultApiController) RedfishV1FabricsGet(w http.ResponseWriter, r *http.Request) {
	Controller.Send(w, nil)
}

// RedfishV1FabricsFabricIdGet -
func (*defaultApiController) RedfishV1FabricsFabricIdGet(w http.ResponseWriter, r *http.Request) {
	Controller.Send(w, mux.Vars(r))
}

/*
// RedfishV1FabricsFabricIdSwitchesGet -
func (*DefaultApiController) RedfishV1FabricsFabricIdSwitchesGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fabricId := params["fabricId"]

	model := sf.SwitchCollectionSwitchCollection{
		OdataId: fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches", fabricId),
		OdataType: "#SwitchCollection.v1_0_0.SwitchCollection",
	}

	Controller.Send(w, model)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdGet -
func (*DefaultApiController) RedfishV1FabricsFabricIdSwitchesSwitchIdGet(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	fabricId := params["fabricId"]
	switchId := params["switchId"]

	model := sf.SwitchV140Switch{
		OdataId: fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s", fabricId, switchId),
		OdataType: "#Switch.v1_4_0.Switch",
		Id: switchId,
	}


	Controller.Send(w, mux.Vars(r))
}
*/
