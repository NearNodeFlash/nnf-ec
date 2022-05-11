/* -----------------------------------------------------------------
 * evttest.go-
 *
 * Register for event and then: 1) send a test event post or wait for
 * something else to send the event post.
 *
 *
 * Author: Tim Morneau
 * Date: 20200804
 *
 * © Copyright 2020 Hewlett Packard Enterprise Development LP
 *
 * ----------------------------------------------------------------- */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	sf "github.com/nearnodeflash/nnf-ec/pkg/rfsf/pkg/common"
	openapi "github.com/nearnodeflash/nnf-ec/pkg/rfsf/pkg/models"
)

const (
	eventTargetUri   = "http://twm320:5001/eventreceiver/"
	eventPath        = "/eventreceiver/"
	eventRegUri      = "http://twm320:5000/redfish/v1/EventService/Subscriptions/1"
	eventTestPostUri = "http://twm320:5000/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
	// date -Iseconds for the event timestamp below.
	eventTestBody = `{
		"EventType": "StatusChange",
		"EventID": "myEventId",

		"EventTimestamp": "2020-08-05T13:56:20-0600",
		"Severity": "OK",
		"Message": "This is a test message",
		"MessageID": "iLOEvents.0.9.ResourceStatusChanged",
		"MessageArgs": [ "arg0", "arg1" ],
		"OriginOfCondition": "/rest/v1/Chassis/1/FooBar"
	}`
)

var totalEvents int = 0

func registerEvent() {
	var ed = sf.EventDestination{
		Context:         "evttest.go application event destination",
		Destination:     eventTargetUri,
		Name:            "Test Event Registration - from evttest.go",
		OriginResources: []openapi.OdataV4IdRef{{OdataId: "/redfish/v1/StorageServices/USE531239CBVS019"}},
		HttpHeaders:     []map[string]interface{}{{"Content-Type": "Application/JSON", "OData-Version": "4.0"}},
	}
	client := &http.Client{}
	data, _ := json.MarshalIndent(ed, "", "   ")
	// fmt.Printf("event destination = %s", data)
	req, err := http.NewRequest(http.MethodPut, eventRegUri, bytes.NewBuffer(data))

	if err != nil {
		fmt.Printf("PUT assembly operation failed\n")
	}
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("PUT operation failed\n")
	}
}

func eventRx(w http.ResponseWriter, _ *http.Request) {
	totalEvents++
	fmt.Printf("Received Event %d\n", totalEvents)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

}

func genTestEvent() {
	fmt.Printf("Entered genTestEvent\n")
	rB, _ := json.Marshal(eventTestBody)
	http.Post(eventTestPostUri, "application/json", bytes.NewBuffer(rB))
}
func handleEvents(cnt int) {
	http.HandleFunc(eventPath, eventRx)
	log.Fatal(http.ListenAndServe(":5001", nil))
}

/*
func main() {
	fmt.Printf("Entered main\n")
	registerEvent()
	go handleEvents(0)
	genTestEvent()
	for i := 0; ; i++ {

	}
}
*/
