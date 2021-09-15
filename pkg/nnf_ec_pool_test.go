package nnf

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "io/ioutil"
	// "log"
	"net/http"
	"testing"
	"time"

	ec "stash.us.cray.com/rabsw/nnf-ec/internal/ec"
	sf "stash.us.cray.com/rabsw/nnf-ec/internal/rfsf/pkg/models"
)

func formatUrl(path string) string {
	return fmt.Sprintf("http://localhost:8080%s", path)
}

func retrieveUrl(url string, model *sf.StorageServiceV150StorageService) (status int) {

	// Fetch the endpoint
	rsp, err := http.Get(url)
	if err != nil {
		return rsp.StatusCode
	}
	if rsp.StatusCode != http.StatusOK {
		return rsp.StatusCode
	}
	defer rsp.Body.Close()

	// Decode response into our specified interface
	json.NewDecoder(rsp.Body).Decode(model)

	return rsp.StatusCode
}

func retrieveStorageService(t *testing.T, model *sf.StorageServiceV150StorageService) {

	url := formatUrl(StorageServiceRootURL)
	status := retrieveUrl(url, model)

	if status != http.StatusOK {
		t.Fatalf("Unable to find %s, %d", url, status)
	}
}

func startMockController() {
	opts := ec.Options{Http: true, Port: 8080, Log: true, Verbose: true}
	MockController.Init(&opts)

	go MockController.Run()

	// Allow the MockController to get started before sending it anything.
	time.Sleep(time.Second)
}

const (
	StorageServiceRootURL = "/redfish/v1/StorageServices/NNF"
)

func TestStorageServiceCreatePool(t *testing.T) {

	t.Skip("MockController doesn't shutdown until the entire test suite is finished.\n" +
		"Skipping test because TestWalkFabricEndpoints checks the endpoints. Consider deleting")

	startMockController()

	var storageServices sf.StorageServiceV150StorageService
	retrieveStorageService(t, &storageServices)
	fmt.Printf("%v\n", storageServices)

	storagePoolsUrl := fmt.Sprintf("%s%s", StorageServiceRootURL, "/StoragePools")

	postStoragePool := func() {
		url := formatUrl(storagePoolsUrl)

		storagePool := sf.StoragePoolV150StoragePool{
			CapacityBytes: 1024 * 1024 * 10,
		}

		jsonReq, _ := json.Marshal(storagePool)

		resp, err := http.NewRequest(ec.POST_METHOD, url, bytes.NewBuffer(jsonReq))
		if err != nil {
			t.Fatal(err)
		}

		w := ec.NewResponseWriter()

		MockController.Send(w, resp)

		if w.StatusCode != http.StatusOK {
			t.Errorf("Test Endpoint Failed %d", w.StatusCode)
		}
	}

	postStoragePool()
}

// ------------------ Original post code
// resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
// if err != nil {
// 	log.Fatalln(err)
// }

// defer resp.Body.Close()
// bodyBytes, _ := ioutil.ReadAll(resp.Body)

// bodyString := string(bodyBytes)
// fmt.Println(bodyString)

/// Retrieve/query StorageServices
///

/// Create StorageServices
///

/// Destroy StorageServices
///
