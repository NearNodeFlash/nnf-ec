/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fabric

import (
	"fmt"
	"net/http"

	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"

	. "github.com/NearNodeFlash/nnf-ec/pkg/common"
	ec "github.com/NearNodeFlash/nnf-ec/pkg/ec"
)

// DefaultApiService -
type DefaultApiService struct {
}

// NewDefaultApiService -
func NewDefaultApiService() Api {
	return &DefaultApiService{}
}

// RedfishV1FabricsGet -
func (*DefaultApiService) RedfishV1FabricsGet(w http.ResponseWriter, r *http.Request) {

	model := sf.FabricCollectionFabricCollection{
		OdataId:   "/redfish/v1/Fabrics",
		OdataType: "#FabricCollection.v1_0_0.FabricCollection",
		Name:      "Fabric Collection",
	}

	err := Get(&model)

	EncodeResponse(model, err, w)
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

	err := FabricIdGet(fabricId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.SwitchCollectionSwitchCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches", fabricId),
		OdataType: "#SwitchCollection.v1_0_0.SwitchCollection",
		Name:      "Switch Collection",
	}

	err := FabricIdSwitchesGet(fabricId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]

	model := sf.SwitchV140Switch{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s", fabricId, switchId),
		OdataType: "#Switch.v1_4_0.Switch",
		Name:      "Switch",
	}

	err := FabricIdSwitchesSwitchIdGet(fabricId, switchId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]

	model := sf.PortCollectionPortCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s/Ports", fabricId, switchId),
		OdataType: "#PortCollection.v1_0_0.PortCollection",
		Name:      "Port Collection",
	}

	err := FabricIdSwitchesSwitchIdPortsGet(fabricId, switchId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdSwitchesSwitchIdPortsPortIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	switchId := params["SwitchId"]
	portId := params["PortId"]

	model := sf.PortV130Port{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Switches/%s/Ports/%s", fabricId, switchId, portId),
		OdataType: "#Port.v1_3_0.Port",
		Name:      "Port",
	}

	err := FabricIdSwitchesSwitchIdPortsPortIdGet(fabricId, switchId, portId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdEndpointsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.EndpointCollectionEndpointCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints", fabricId),
		OdataType: "#EndpointCollection.v1_0_0.EndpointCollection",
		Name:      "Endpoint Collection",
	}

	err := FabricIdEndpointsGet(fabricId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdEndpointsEndpointIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointsEndpointIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	endpointId := params["EndpointId"]

	model := sf.EndpointV150Endpoint{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Endpoints/%s", fabricId, endpointId),
		OdataType: "#Endpoint.v1_5_0.Endpoint",
		Name:      "Endpoint",
	}

	err := FabricIdEndpointsEndpointIdGet(fabricId, endpointId, &model)

	EncodeResponse(model, err, w)
}

func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointGroupsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.EndpointGroupCollectionEndpointGroupCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/EndpointGroups", fabricId),
		OdataType: "#EndpointGroupCollection.v1_0_0.EndpointGroupCollection",
		Name:      "Endpoint Group Collection",
	}

	err := FabricIdEndpointGroupsGet(fabricId, &model)

	EncodeResponse(model, err, w)
}

func (*DefaultApiService) RedfishV1FabricsFabricIdEndpointGroupsEndpointGroupIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	groupId := params["EndpointGroupId"]

	model := sf.EndpointGroupV130EndpointGroup{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/EndpointGroups/%s", fabricId, groupId),
		OdataType: "#EndpointGroup.v1_3_0.EndpointGroup",
		Name:      "Endpoint Group",
	}

	err := FabricIdEndpointGroupsEndpointIdGet(fabricId, groupId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdConnectionsGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdConnectionsGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]

	model := sf.ConnectionCollectionConnectionCollection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections", fabricId),
		OdataType: "#ConnectionCollection.v1_0_0.ConnectionCollection",
		Name:      "Connection Collection",
	}

	err := FabricIdConnectionsGet(fabricId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdConnectionsConnectionIdGet -
func (*DefaultApiService) RedfishV1FabricsFabricIdConnectionsConnectionIdGet(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	connectionId := params["ConnectionId"]

	model := sf.ConnectionV100Connection{
		OdataId:   fmt.Sprintf("/redfish/v1/Fabrics/%s/Connections/%s", fabricId, connectionId),
		OdataType: "#Connection.v1_0_0.Connection",
		Name:      "Connection",
	}

	err := FabricIdConnectionsConnectionIdGet(fabricId, connectionId, &model)

	EncodeResponse(model, err, w)
}

// RedfishV1FabricsFabricIdConnectionsConnectionIdPatch -
func (*DefaultApiService) RedfishV1FabricsFabricIdConnectionsConnectionIdPatch(w http.ResponseWriter, r *http.Request) {
	params := Params(r)
	fabricId := params["FabricId"]
	connectionId := params["ConnectionId"]

	model := sf.ConnectionV100Connection{}

	if err := UnmarshalRequest(r, &model); err != nil {
		EncodeResponse(model, ec.NewErrBadRequest().WithError(err), w)
		return
	}

	err := FabricIdConnectionsConnectionIdPatch(fabricId, connectionId, &model)

	EncodeResponse(model, err, w)
}
