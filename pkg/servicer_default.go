package nnf

import (
	"encoding/json"
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric"
	sf "stash.us.cray.com/sp/rfsf-openapi/pkg/models"
)

// DefaultApiService -
type DefaultApiService struct {
}

// NewDefaultApiService -
func NewDefaultApiService() DefaultApiServicer {
	return &DefaultApiService{}
}

func Params(r string) (params map[string]string) {

	json.Unmarshal([]byte(r), &params)

	return params
}

// RedfishV1FabricsGet -
func (*DefaultApiService) RedfishV1FabricsGet(r string) (string, error) {

	model := sf.FabricCollectionFabricCollection{
		OdataId:   "/redfish/v1/Fabrics",
		OdataType: "#FabricCollection.v1_0_0.FabricCollection",
		Name:      "Fabric Collection",
	}

	if err := fabric.Get(&model); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(model)
}

// RedfishV1FabricsFabricIdGet
func (*DefaultApiService) RedfishV1FabricsFabricIdGet(r string) (string, error) {
	params := Params(r)
	fabricId := params["fabricId"]

	model := sf.FabricV120Fabric{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s", fabricId),
		OdataType: "#Fabric.v1_0_0.Fabric",
		Id:        fabricId,
		Name:      "Fabric",
	}

	if err := fabric.FabricIdGet(fabricId, &model); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(model)
}

/*
// RedfishV1FabricsFabricIdSwitchesGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesGet(r string) (string, error) {
	if err := fabric.FabricIdSwitchesGet(&model); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(model)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdGet
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdGet(r string) (string, error) {
	model := sf.SwitchV140Switch{}

	if err := json.Unmarshal([]byte(r), &model); err != nil {
		return encodeErrorResponse(err)
	}

	return encodeJsonResponse(model)
}
*/

func encodeErrorResponse(err error) (string, error) {
	return "", err
}

func encodeJsonResponse(s interface{}) (string, error) {
	response, err := json.Marshal(s)
	return string(response), err
}
