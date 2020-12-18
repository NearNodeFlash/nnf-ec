package nnf

import (
	"net/http"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// DefaultApiController -
type DefaultApiController struct{}

// NewDefaultApiController -
func NewDefaultApiController() DefaultApi {
	return &DefaultApiController{}
}

// RedfishV1FabricsGet -
func (*DefaultApiController) RedfishV1FabricsGet(w http.ResponseWriter, r *http.Request) {
	log.Info("RedfishV1FabricsGet")
	Controller.Send(w, nil)
}

// RedfishV1FabricsFabricIdGet -
func (*DefaultApiController) RedfishV1FabricsFabricIdGet(w http.ResponseWriter, r *http.Request) {
	log.Info("RedfishV1FabricsFabricIdGet")
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
