package nnf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric"
	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

// DefaultApiService -
type DefaultApiService struct {
}

// NewDefaultApiService -
func NewDefaultApiService() Api {
	return &DefaultApiService{}
}

// Params -
func Params(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// RedfishV1FabricsGet -
func (*DefaultApiService) RedfishV1FabricsGet(w http.ResponseWriter, r *http.Request) {

	model := sf.FabricCollectionFabricCollection{
		OdataId:   "/redfish/v1/Fabrics",
		OdataType: "#FabricCollection.v1_0_0.FabricCollection",
		Name:      "Fabric Collection",
	}

	err := fabric.Get(&model)

	log.WithError(err).Info("RedfishV1FabricsGet")

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.FabricV120Fabric{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s", fabricId),
		OdataType: "#Fabric.v1_0_0.Fabric",
		Id:        fabricId,
		Name:      "Fabric",
	}

	err := fabric.FabricIdGet(fabricId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.SwitchCollectionSwitchCollection{}

	err := fabric.FabricIdSwitchesGet(fabricId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]

	model := sf.SwitchV140Switch{}

	err := fabric.FabricIdSwitchesSwitchIdGet(fabricId, switchId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]

	model := sf.PortCollectionPortCollection{}

	err := fabric.FabricIdSwitchesSwitchIdPortsGet(fabricId, switchId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]
	portId := params["PortId"]

	model := sf.PortV130Port{}

	err := fabric.FabricIdSwitchesSwitchIdPortsPortIdGet(fabricId, switchId, portId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdEndpointsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.EndpointCollectionEndpointCollection{}

	err := fabric.FabricIdEndpointsGet(fabricId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdEndpointsEndpointIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointsEndpointIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	endpointId := params["EndpointId"]

	model := sf.EndpointV150Endpoint{}

	err := fabric.FabricIdEndpointsEndpointIdGet(fabricId, endpointId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdConnectionsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdConnectionsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.ConnectionCollectionConnectionCollection{}

	err := fabric.FabricIdConnectionsGet(fabricId, &model)

	encodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdConnectionsPost -
func (*DefaultApiService) RedfishV1FabricsFabricIdConnectionsPost(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	// TODO: Move the read & unmarshal into a function
	var model sf.ConnectionV100Connection

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// TODO:
	}

	err = json.Unmarshal(body, &model)
	if err != nil {
		// TODO:
	}

	err = fabric.FabricIdConnectionsPost(fabricId, &model)

	encodeResponse(model, err, w) // TODO: Need to return header information for the ConnectionId
}

func encodeResponse(s interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		log.WithError(err).Warn("Element Controller Error")
	}

	var e *ec.ControllerError
	if errors.As(err, &e) {
		w.WriteHeader(e.StatusCode)
		return
	}

	response, err := json.Marshal(s)
	log.WithError(err).Infof("Writing response %s", string(response))
	if err != nil {
		log.WithError(err).Error("Failed to marshal json response")
		// TODO
	}
	_, err = w.Write(response)
	if err != nil {
		log.WithError(err).Error("Failed to write json response")
		// TODO
	}
	//log.WithError(err).Infof("Wrote response %s", string(w.Buffer.Bytes()))
}
