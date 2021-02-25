package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
	"stash.us.cray.com/rabsw/nnf-ec/internal/fabric-manager"
	nnf "stash.us.cray.com/rabsw/nnf-ec/pkg"
	sf "stash.us.cray.com/rabsw/rfsf-openapi/pkg/models"
)

var (
	MockController = nnf.NewController(true)
)

type TestRoute struct {
	URL  string
	Intf interface{}
	Func func(t *testing.T, v interface{})
}

func Routes() []TestRoute {
	return []TestRoute{

		// This is a simple example on how to test any endpoint we have defined by
		// the NNF Element Controller. Define the URL you want to fetch and the
		// expected interface{} pointer to decode as a response. Finally, write
		// a verification function to verify the model data.
		{
			URL:  "/redfish/v1/Fabrics",
			Intf: &sf.FabricCollectionFabricCollection{},
			Func: func(t *testing.T, v interface{}) {
				model := v.(*sf.FabricCollectionFabricCollection)

				if model.MembersodataCount != 1 {
					t.Errorf("Wrong member count in fabric")
				} else if model.Members[0].OdataId != fmt.Sprintf("/redfish/v1/Fabrics/%s", fabric.FabricId) {
					t.Errorf("Wrong fabric id")
				}
			},
		},
		{
			URL:  "/redfish/v1/Fabrics/{FabricId}/Switches/0/Ports/10",
			Intf: &sf.PortV130Port{},
			Func: func(t *testing.T, v interface{}) {
				model := v.(*sf.PortV130Port)

				if model.Links.AssociatedEndpointsodataCount != 18 {
					t.Errorf("Wrong number of endpoints associated with DSP")
				}
			},
		},
	}
}

func TestMockController(t *testing.T) {

	go MockController.Run()

	formatUrl := func(path string) string {
		path = strings.Replace(path, "{FabricId}", fabric.FabricId, 1)
		return fmt.Sprintf("http://localhost:8080%s", path)
	}

	for _, route := range Routes() {
		url := formatUrl(route.URL)

		r, _ := http.NewRequest(string(ec.GET_METHOD), url, strings.NewReader(""))
		r.RequestURI = url // Something about only the URI making it to the EC

		w := ec.NewResponseWriter()
		MockController.Send(w, r)

		if err := json.Unmarshal(w.Buffer.Bytes(), route.Intf); err != nil {
			t.Errorf("Failed to unmarshal response: +%v", err)
		}

		route.Func(t, route.Intf)
	}
	/*


		for _, route := range MockController.Router.Routes() {

			if route.Method != nnf.GET_METHOD {
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
}
