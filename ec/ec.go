package elementcontroller

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	client "stash.us.cray.com/sp/dp-api/api/grpc/v1/grpc-client"
	pb "stash.us.cray.com/sp/dp-common/api/proto/v1/dp-element_cntrl"
)

// Controller -
type Controller struct {
	Name     string
	Port     string
	Version  string
	Servicer interface{}
}

// Send -
func (c *Controller) Send(w http.ResponseWriter, payload interface{}) {

	arg, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Warnf("Element Controller %s: Send Failed", c.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	funcName := getFunctionName()
	r := client.ClientRequest{
		ElmntCntrlFunc:        funcName,
		ElmntCntrlFuncJsonArg: arg,
		ElmntCntrlPort:        ":" + c.Port,
		ElmntCntrlName:        c.Name,

		HTTPwriter: w,
	}

	log.Warnf("Element Controller %s: Process Function: %s", c.Name, funcName)
	r.ProcessRequest()
}

func getFunctionName() string {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frame, _ := runtime.CallersFrames(pcs[:n]).Next()
	return strings.SplitAfter(path.Base(frame.Function), ".")[1]
}

// checkAPI -
func (c *Controller) checkAPI(api string) error {
	return nil
}

// SendElmntCntrlTaskRequest -
func (c *Controller) SendElmntCntrlTaskRequest(_ context.Context, in *pb.CreateElmntCntrlTaskRequest) (*pb.CreateElmntCntrlTaskResponse, error) {
	log.Info("%s Received Task Request: %s: %s", c.Name, in.Sender, in.Task.Name)

	if err := c.checkAPI(in.Api); err != nil {
		log.WithError(err).Warnf("Unsupported API version %s", in.Api)
		return nil, err
	}

	// Use reflection to call the method requested and pass in the JSON message as args
	method := reflect.ValueOf(c.Servicer).MethodByName((in.Task.Name))
	if !method.IsValid() {
		log.Errorf("Method %s not found in ", in.Task.Name, c.Name)
		return nil, nil
	}

	values := make([]reflect.Value, method.Type().NumIn())
	values[0] = reflect.ValueOf(in.Task.JsonMsg)

	// Call the method with input values of the form [0]jsonRequest.([]byte), and
	// expect response in form [0]jsonResponse, [1]error
	response := method.Call(values)

	if !response[1].IsNil() {
		return nil, response[1].Interface().(error)
	}

	return &pb.CreateElmntCntrlTaskResponse{
		Api:      c.Version,
		JsonData: response[0].Interface().(string),
	}, nil
}

// Run -
func Run(c *Controller) {
	log.Infof("Starting %s Element Controller on Port %s", c.Name, c.Port)
	listen, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on port %s", c.Port)
	}

	server := grpc.NewServer()

	pb.RegisterDP_Elmnt_CntrlServer(server, c)

	if err := server.Serve(listen); err != nil {
		log.WithError(err).Fatalf("Failed to server %s", c.Name)
	}

	log.Warn(" %s Element Controller Terminated", c.Name)
}
