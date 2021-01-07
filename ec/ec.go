package elementcontroller

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	client "stash.us.cray.com/sp/dp-api/api/grpc/v2/grpc-client"
	msgs "stash.us.cray.com/sp/dp-common/api/proto/v2/dp-api_msgs"
	pb "stash.us.cray.com/sp/dp-common/api/proto/v2/dp-ec"
)

// Route
type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

// Routes -
type Routes []Route

// Router -
type Router interface {
	Routes() Routes
}

// Controller -
type Controller struct {
	Name    string
	Port    string
	Version string
	Router  Router
	Mux     *mux.Router
}

// ResponseWriter -
type ResponseWriter struct {
	Buffer     bytes.Buffer
	StatusCode int
}

func NewResponseWriter() ResponseWriter {
	return ResponseWriter{}
}

func (r ResponseWriter) Header() http.Header {
	return http.Header{}
}

func (r ResponseWriter) Write(b []byte) (int, error) {
	return r.Buffer.Write(b)
}

func (r ResponseWriter) WriteHeader(s int) {
	r.StatusCode = s
}

// Send -
func (c *Controller) Send(w http.ResponseWriter, r *http.Request) {

	// Initialize ClientRequest object
	grpcReq := client.ClientRequest{
		ElmntCntrlName: c.Name,
		ElmntCntrlPort: ":" + c.Port,
		HTTPwriter:     w,
	}

	// Record timestamp of request and reminder
	t := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	pfx := t.Format(time.RFC3339Nano)

	// Construct and send element controller request
	funcName := getParentFunctionName()
	data, _ := ioutil.ReadAll(r.Body)

	request := pb.ECTaskRequest{
		Api:       c.Version,
		Sender:    msgs.DPAPIname,
		Name:      funcName,
		Uri:       r.RequestURI,
		Method:    r.Method,
		JsonMsg:   string(data),
		Reminder:  reminder,
		Timestamp: pfx,
	}

	grpcReq.ProcessRequest(&request)
}

// Returns the parent's function name at the point of calling this function.
// i.e. If the call chain is A()->B()->getParentFunctionName() then 'A' is returned.
func getParentFunctionName() string {
	var pcs [4]uintptr
	n := runtime.Callers(3, pcs[:])
	frame, _ := runtime.CallersFrames(pcs[:n]).Next()
	return frame.Function[strings.LastIndex(frame.Function, ".")+1:]
}

// checkAPI -
func (c *Controller) checkApiVersion(api string) error {
	if c.Version != api {
		return status.Errorf(codes.Unimplemented, "%s: Unsupported API Version", c.Name)
	}
	return nil
}

// ProcessTaskRequest -
func (c *Controller) ProcessTaskRequest(_ context.Context, in *pb.ECTaskRequest) (*pb.ECTaskResponse, error) {
	log.Infof("Received Task request from [%s] for method [%s]", in.Sender, in.Method)

	if err := c.checkApiVersion(in.Api); err != nil {
		log.WithError(err).Warnf("API Version incorrect %s", in.Api)
		return nil, err
	}

	// Rebuild the HTTP Request
	req, err := http.NewRequest(in.Method, in.Uri, strings.NewReader(in.JsonMsg))
	if err != nil {
		log.WithError(err).Errorf("Could not build http request")
		return nil, status.Error(codes.Internal, "Could not build http request")
	}

	res := NewResponseWriter()
	c.Mux.ServeHTTP(res, req)

	if res.StatusCode != http.StatusOK {
		log.Warnf("Request faild with status %d", res.StatusCode)
		err = http.ErrNotSupported // TODO: Should have a encoding map from StatusCode to Err
	}

	return &pb.ECTaskResponse{
		Api:      c.Version,
		JsonData: string(res.Buffer.Bytes()),
	}, err
}

// Run -
func Run(c *Controller) {
	log.Infof("Starting %s Element Controller on Port %s", c.Name, c.Port)

	c.Mux = mux.NewRouter().StrictSlash(true)
	for _, route := range c.Router.Routes() {
		c.Mux.
			Name(route.Name).
			Methods(route.Method).
			Path(route.Path).
			Handler(route.HandlerFunc)
	}

	listen, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on port %s", c.Port)
	}

	server := grpc.NewServer()

	pb.RegisterControllerServiceServer(server, c)

	if err := server.Serve(listen); err != nil {
		log.WithError(err).Fatalf("Failed to serve %s", c.Name)
	}

	log.Warnf("%s Element Controller Terminated", c.Name)
}
