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

package ec

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"encoding/json"
)

type TestApiRouter struct{}

func NewTestApiRouter() Router {
	return &TestApiRouter{}
}

func (*TestApiRouter) Name() string      { return "TestRouter" }
func (*TestApiRouter) Init(Logger) error { return nil }
func (*TestApiRouter) Start() error      { return nil }
func (*TestApiRouter) Close() error      { return nil }

func (*TestApiRouter) Routes() Routes {
	return Routes{{
		Name:        "TestGet",
		Method:      GET_METHOD,
		Path:        "/test",
		HandlerFunc: testHandlerFuncGet,
	}, {
		Name:        "TestPost",
		Method:      POST_METHOD,
		Path:        "/test",
		HandlerFunc: testHandlerFuncPost,
	}, {
		Name:        "TestFail",
		Method:      GET_METHOD,
		Path:        "/testFail",
		HandlerFunc: testHandlerFuncFail,
	}, {
		Name:        "TestFailNoError",
		Method:      GET_METHOD,
		Path:        "/testFailNoError",
		HandlerFunc: testHandlerFuncFailNoError,
	}}
}

const testMessage = "Element Controller Test"

func testHandlerFuncGet(w http.ResponseWriter, r *http.Request) {

	rsp, err := json.Marshal(&testModel{Message: testMessage})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(rsp)
}

func testHandlerFuncPost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	model := testModel{}
	if err := json.Unmarshal(body, &model); err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	if model.Message != testMessage {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(body)
}

func testHandlerFuncFail(w http.ResponseWriter, r *http.Request) {
	model := testModel{}

	EncodeResponse(
		model,
		NewErrNotAcceptable().WithError(fmt.Errorf("Test Error")).WithCause("Test Fail Func"),
		w)
}

func testHandlerFuncFailNoError(w http.ResponseWriter, r *http.Request) {
	model := testModel{}

	EncodeResponse(
		model,
		NewErrNotAcceptable().WithCause("Test Fail Func"),
		w)
}

type testModel struct {
	Message string
}

func TestHttp(t *testing.T) {

	var c = Controller{
		Name: "TestController",
		Port: 8081,
		Routers: Routers{
			NewTestApiRouter(),
		},
	}

	c.Init(&Options{Http: true, Log: true, Verbose: true})
	defer c.Close()

	go c.Run()
	time.Sleep(1 * time.Second)

	Request(t, &c, GET_METHOD)
	Request(t, &c, POST_METHOD)

	RequestFail(t, &c)
}

func Request(t *testing.T, c *Controller, method string) {
	url := fmt.Sprintf("http://localhost:%d/test", c.Port)

	body := []byte{}
	if method == POST_METHOD {
		body, _ = json.Marshal(&testModel{Message: testMessage})
	}

	r, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	w := NewResponseWriter()

	c.Send(w, r)

	if w.StatusCode != http.StatusOK {
		t.Errorf("Test Endpoint Failed %d", w.StatusCode)
	}

	rsp := testModel{}
	if err := json.Unmarshal(w.Buffer.Bytes(), &rsp); err != nil {
		t.Error(err)
	}

	if rsp.Message != testMessage {
		t.Errorf("Test Response not received")
	}
}

func RequestFail(t *testing.T, c *Controller) {

	for _, path := range []string{"testFail", "testFailNoError"} {
		url := fmt.Sprintf("http://localhost:%d/%s", c.Port, path)

		r, err := http.NewRequest(GET_METHOD, url, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := NewResponseWriter()

		c.Send(w, r)

		if w.StatusCode != http.StatusNotAcceptable {
			t.Errorf("Test Endpoint did not fail as expected %d", w.StatusCode)
		}

		rsp := ErrorResponse{}
		if err := json.Unmarshal(w.Buffer.Bytes(), &rsp); err != nil {
			t.Error(err)
		}

		if rsp.Status != http.StatusNotAcceptable {
			t.Errorf("Response Status Incorrect: Expected: %d Actual: %d", http.StatusNotAcceptable, rsp.Status)
		}

		rspm := testModel{}
		if err := json.Unmarshal([]byte(rsp.Model), &rspm); err != nil {
			t.Errorf("Unable to marshal model response %s", err)
		}
	}

}
