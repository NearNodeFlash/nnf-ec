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

package nnf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	ec "github.com/NearNodeFlash/nnf-ec/pkg/ec"
	fabric "github.com/NearNodeFlash/nnf-ec/pkg/manager-fabric"
	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
)

var controllerOption = NewMockOptions(false)
var (
	MockController = NewController(controllerOption)
)

type TestRoute struct {
	Name  string
	URL   string
	Intf  interface{}
	Count int
	Func  func(t *testing.T, v interface{}, expectedCount int)
}

func Routes() []TestRoute {
	return []TestRoute{

		// This is a example on how to test any endpoint we have defined by
		// the NNF Element Controller. Define the URL you want to fetch and the
		// expected interface{} pointer to decode as a response. Finally, write
		// a verification function to verify the model data.
		//
		// Fabric Manager endpoints
		{
			Name:  "RedfishV1FabricsGet",
			URL:   "/redfish/v1/Fabrics",
			Intf:  &sf.FabricCollectionFabricCollection{},
			Count: 1,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.FabricCollectionFabricCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Fatalf("Wrong member count in fabric: %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}

				expectedOdataId := fmt.Sprintf("/redfish/v1/Fabrics/%s", fabric.FabricId)
				if model.Members[0].OdataId != expectedOdataId {
					t.Errorf("Wrong fabric id: %s expecting %s", model.Members[0].OdataId, expectedOdataId)
				}
			},
		},
		{
			Name: "RedfishV1FabricsFabricIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}",
			Intf: &sf.FabricV120Fabric{},
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.FabricV120Fabric)

				t.Logf("%v", model)

				if model.FabricType != sf.PC_IE_PP {
					t.Fatalf("Wrong FabricType %s expecting: %s", model.FabricType, sf.PC_IE_PP)
				}
			},
		},
		{
			Name:  "RedfishV1FabricsFabricIdSwitchesGet",
			URL:   "/redfish/v1/Fabrics/{FabricId}/Switches",
			Intf:  &sf.SwitchCollectionSwitchCollection{},
			Count: 2,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.SwitchCollectionSwitchCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		{
			Name: "RedfishV1FabricsFabricIdSwitchesSwitchIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}",
			Intf: &sf.SwitchV140Switch{},
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.SwitchV140Switch)

				t.Logf("%v", model)
				// if model.Id !=

			},
		},

		{
			Name:  "RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet",
			URL:   "/redfish/v1/Fabrics/{FabricId}/Switches/{SwitchId}/Ports",
			Intf:  &sf.PortCollectionPortCollection{},
			Count: 19,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.PortCollectionPortCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		{
			Name: "RedfishV1FabricsFabricIdSwitchesSwitchIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}/Switches/0/Ports/10",
			Intf: &sf.PortV130Port{},
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.PortV130Port)

				if model.Links.AssociatedEndpointsodataCount != 18 {
					t.Errorf("Wrong number of endpoints associated with DSP: %d expecting: %d", model.Links.AssociatedEndpointsodataCount, 18)
				}
			},
		},

		{
			Name:  "RedfishV1FabricsFabricIdEndpointsGet",
			URL:   "/redfish/v1/Fabrics/{FabricId}/Endpoints",
			Intf:  &sf.EndpointCollectionEndpointCollection{},
			Count: 341,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.EndpointCollectionEndpointCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		{
			Name: "RedfishV1FabricsFabricIdEndpointsEndpointIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}/Endpoints/{EndpointId}",
			Intf: &sf.EndpointV150Endpoint{},
			Func: nil,
		},

		{
			Name:  "RedfishV1FabricsFabricIdEndpointGroupsGet",
			URL:   "/redfish/v1/Fabrics/{FabricId}/EndpointGroups",
			Intf:  &sf.EndpointGroupCollectionEndpointGroupCollection{},
			Count: 17,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.EndpointGroupCollectionEndpointGroupCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},
		{
			Name: "RedfishV1FabricsFabricIdEndpointGroupsEndpointGroupIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}/EndpointGroups/{EndpointGroupId}",
			Intf: &sf.EndpointGroupV130EndpointGroup{},
			Func: nil,
		},
		{
			Name:  "RedfishV1FabricsFabricIdConnectionsGet",
			URL:   "/redfish/v1/Fabrics/{FabricId}/Connections",
			Intf:  &sf.ConnectionCollectionConnectionCollection{},
			Count: 17,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.ConnectionCollectionConnectionCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},
		{
			Name: "RedfishV1FabricsFabricIdConnectionsConnectionIdGet",
			URL:  "/redfish/v1/Fabrics/{FabricId}/Connections/{ConnectionId}",
			Intf: &sf.ConnectionV100Connection{},
			Func: nil,
		},
	}
}

func TestFabricManagerEndpoints(t *testing.T) {

	t.Skip("MockController doesn't shutdown until the entire test suite is finished.\n" +
		"Skipping test because TestWalkFabricEndpoints checks the endpoints. Consider deleting")

	opts := ec.Options{Http: true, Port: 8080, Log: true, Verbose: true}
	MockController.Init(&opts)
	defer MockController.Close()

	go MockController.Run()

	// Allow the MockController to get going before sending it anything.
	time.Sleep(time.Second)
	formatUrl := func(path string) string {
		path = strings.Replace(path, "{FabricId}", fabric.FabricId, 1)
		path = strings.Replace(path, "{SwitchId}", "0", 1)
		path = strings.Replace(path, "{EndpointId}", "0", 1)
		path = strings.Replace(path, "{EndpointGroupId}", "0", 1)
		path = strings.Replace(path, "{ConnectionId}", "0", 1)
		return fmt.Sprintf("http://localhost:8080%s", path)
	}

	t.Logf("Given the need to access all fabric endpoints.")
	for _, route := range Routes() {

		runTheRoute := func() {
			url := formatUrl(route.URL)

			t.Logf("--When checking endpoint %s URL:%s formattedURL:%s", route.Name, route.URL, url)
			// r, _ := http.NewRequest(string(ec.GET_METHOD), url, strings.NewReader(""))
			// r.RequestURI = url // Something about only the URI making it to the EC

			// w := ec.NewResponseWriter()
			// MockController.Send(w, r)

			// if err := json.Unmarshal(w.Buffer.Bytes(), route.Intf); err != nil {
			// 	test.Errorf("Failed to unmarshal response: +%v", err)
			// }

			// Fetch the endpoint
			rsp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Unable to GET %s, %s", url, err)
			}
			if rsp.StatusCode != http.StatusOK {
				t.Fatalf("Unable to find %s, %s", url, rsp.Status)
			}

			// Decode response into our specified interface
			json.NewDecoder(rsp.Body).Decode(route.Intf)
			rsp.Body.Close()

			// If there is a function for the route to verify the response
			if route.Func != nil {
				route.Func(t, route.Intf, route.Count)
			}
		}

		runTheRoute()
	}
	/*


		for _, route := range MockController.Router.Routes() {

			if route.Method != GET_METHOD {
				continue
			}
			url := formatUrl(route.Path)
			r, err := http.NewRequest(route.Method, url, strings.NewReader(""))
			r.RequestURI = url

			fmt.Printf("Send %s %s (%v)\n", route.Path, r.URL, err)

			w := ec.NewResponseWriter()
			MockController.Send(w, r)

			fmt.Printf("Response: %s\n", string(w.Buffer.Bytes()))
		}
	*/

	// Need some way to stop MockController
	// MockController.Halt
}

func RouteCollections() []TestRoute {
	return []TestRoute{

		// These routes return collections of endpoints.
		// First we retrieve the collection, then we retrieve each of the specified routes withih the collection
		//
		// Fabric Manager endpoints
		{
			Name:  "RedfishV1FabricsGet",
			URL:   "/redfish/v1/Fabrics",
			Intf:  &sf.FabricCollectionFabricCollection{},
			Count: 1,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.FabricCollectionFabricCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Fatalf("Wrong member count in fabric: %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}

				expectedOdataId := fmt.Sprintf("/redfish/v1/Fabrics/%s", fabric.FabricId)
				if model.Members[0].OdataId != expectedOdataId {
					t.Errorf("Wrong fabric id: %s expecting %s", model.Members[0].OdataId, expectedOdataId)
				}
			},
		},

		{
			Name:  "RedfishV1FabricsFabricIdSwitchesGet",
			URL:   "{FabricId}/Switches",
			Intf:  &sf.SwitchCollectionSwitchCollection{},
			Count: 2,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.SwitchCollectionSwitchCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		// This one is created inline within the function below.
		// {
		// 	Name:  "RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet",
		// 	URL:   "{FabricId}/Switches/{SwitchId}/Ports",
		// 	Intf:  &sf.PortCollectionPortCollection{},
		// 	Count: 19,
		// 	Func: func(t *testing.T, v interface{}, expectedCount int) {
		// 		model := v.(*sf.PortCollectionPortCollection)

		// 		expectedCount64 := int64(expectedCount)
		// 		if model.MembersodataCount != expectedCount64 {
		// 			t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
		// 		}

		// 		// Ensure we have records for all the members we think should be here
		// 		if len(model.Members) != expectedCount {
		// 			t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
		// 		}
		// 	},
		// },

		{
			Name:  "RedfishV1FabricsFabricIdEndpointsGet",
			URL:   "{FabricId}/Endpoints",
			Intf:  &sf.EndpointCollectionEndpointCollection{},
			Count: 341,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.EndpointCollectionEndpointCollection)

				t.Logf("%v", model)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		{
			Name:  "RedfishV1FabricsFabricIdEndpointGroupsGet",
			URL:   "{FabricId}/EndpointGroups",
			Intf:  &sf.EndpointGroupCollectionEndpointGroupCollection{},
			Count: 17,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.EndpointGroupCollectionEndpointGroupCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		{
			Name:  "RedfishV1FabricsFabricIdConnectionsGet",
			URL:   "{FabricId}/Connections",
			Intf:  &sf.ConnectionCollectionConnectionCollection{},
			Count: 17,
			Func: func(t *testing.T, v interface{}, expectedCount int) {
				model := v.(*sf.ConnectionCollectionConnectionCollection)

				expectedCount64 := int64(expectedCount)
				if model.MembersodataCount != expectedCount64 {
					t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, expectedCount64)
				}

				// Ensure we have records for all the members we think should be here
				if len(model.Members) != expectedCount {
					t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), expectedCount)
				}
			},
		},

		// Storage Services
		{
			Name:  "RedfishV1StorageServicesGet",
			URL:   "/redfish/v1/StorageServices",
			Count: 1,
			Func:  nil,
		},
	}
}

func TestWalkFabricEndpoints(t *testing.T) {

	opts := ec.Options{Http: true, Port: 8080, Log: true, Verbose: true}

	MockController.Init(&opts)
	defer MockController.Close()

	go MockController.Run()

	// Allow the MockController to get going before sending it anything.
	time.Sleep(time.Second)

	// Structure to capture the individual URLs in the collection
	// to visit
	type subURLs struct {
		// The number of items in a collection.
		MembersodataCount int64 `json:"Members@odata.count"`
		// The members of this collection.
		Members []sf.OdataV4IdRef `json:"Members"`
	}

	formatUrl := func(baseURL string, path string) string {
		path = strings.Replace(path, "{FabricId}", baseURL, 1)
		return fmt.Sprintf("http://localhost:8080%s", path)
	}

	runTheRoute := func(url string, model *subURLs) {
		t.Logf("--When checking endpoint URL:%s", url)

		// Fetch the endpoint
		rsp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Unable to GET %s, %s", url, err)
		}
		if rsp.StatusCode != http.StatusOK {
			t.Fatalf("Unable to find %s, %s", url, rsp.Status)
		}
		defer rsp.Body.Close()

		// Decode response into our specified interface
		json.NewDecoder(rsp.Body).Decode(model)
	}

	var fabricId string
	exploreTheCollection := func(route TestRoute, model *subURLs) {
		url := formatUrl(fabricId, route.URL)

		runTheRoute(url, model)
		// t.Logf("%v", model)

		// The first collection we get (Fabrics) only member is the fabricId URL
		// we use subsequently
		if fabricId == "" {
			fabricId = model.Members[0].OdataId
		}

		// Validate that the count we retrieved matches the expected count for the route
		if model.MembersodataCount != int64(route.Count) {
			t.Errorf("Wrong member count %d expecting: %d", model.MembersodataCount, int64(route.Count))
		}

		// Ensure we have records for all the members we think should be here
		if len(model.Members) != route.Count {
			t.Errorf("Wrong number of members: %d expecting: %d", len(model.Members), route.Count)
		}

		// Iterate through all of the members of this collection
		for i := range model.Members {
			// fabricId is already embedded in the model.Member
			subURL := formatUrl("", model.Members[i].OdataId)

			// The members retrieved from the collection are the full URL, and we don't expect subURLs here
			runTheRoute(subURL, nil)
		}
	}

	for _, nextRoute := range RouteCollections() {

		model := subURLs{}

		exploreTheCollection(nextRoute, &model)

		// If we just accessed the Switches collection, we have more work to do.
		// There are Ports attached to each switch, so we need to access them as well.
		if strings.Contains(model.Members[0].OdataId, "Switches") {

			for port := range model.Members {

				portURL := fmt.Sprintf("%s/Ports", model.Members[port].OdataId)

				portRoute := TestRoute{
					Name:  "RedfishV1FabricsFabricIdSwitchesSwitchIdPortsGet",
					URL:   portURL,
					Intf:  &sf.PortCollectionPortCollection{},
					Count: 19,
					Func:  nil,
				}

				portModel := subURLs{}

				exploreTheCollection(portRoute, &portModel)
			}
		}
	}
}
