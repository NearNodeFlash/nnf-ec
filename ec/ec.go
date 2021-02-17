package elementcontroller

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	log "github.com/sirupsen/logrus"

	client "stash.us.cray.com/sp/dp-api/api/grpc/v2/grpc-client"
	msgs "stash.us.cray.com/sp/dp-common/api/proto/v2/dp-api_msgs"
	pb "stash.us.cray.com/sp/dp-common/api/proto/v2/dp-ec"
)

var (
	GET_METHOD    = strings.ToUpper("Get")
	POST_METHOD   = strings.ToUpper("Post")
	PATCH_METHOD  = strings.ToUpper("Patch")
	DELETE_METHOD = strings.ToUpper("Delete")
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

	Name() string
	Init() error
	Start() error
}

// Routers -
type Routers []Router

// Controller -
type Controller struct {
	Name    string
	Port    string
	Version string
	Routers Routers
	Mux     *mux.Router
}

// ResponseWriter -
type ResponseWriter struct {
	Hdr        http.Header
	StatusCode int
	Buffer     *bytes.Buffer
}

func NewResponseWriter() ResponseWriter {
	return ResponseWriter{
		Hdr:        http.Header{},
		StatusCode: http.StatusOK,
		Buffer:     bytes.NewBuffer([]byte{}),
	}
}

func (r ResponseWriter) Header() http.Header {
	return r.Hdr
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

// setupRouters -
func setupRouters(routers Routers) (*mux.Router, error) {

	m := mux.NewRouter().StrictSlash(true)

	for _, router := range routers {

		if err := router.Init(); err != nil {
			return nil, fmt.Errorf("%s failed to initialize", router.Name())
		}

		for _, route := range router.Routes() {
			m.
				Name(route.Name).
				Methods(route.Method).
				Path(route.Path).
				Handler(route.HandlerFunc)
		}
	}

	for _, router := range routers {
		if err := router.Start(); err != nil {
			return nil, fmt.Errorf("%s failed to start", router.Name())
		}
	}

	return m, nil
}

// ProcessTaskRequest -
func (c *Controller) ProcessTaskRequest(_ context.Context, in *pb.ECTaskRequest) (*pb.ECTaskResponse, error) {
	log.Infof("Received Task request from [%s] for method [%s] [%s]", in.Sender, in.Method, in.Uri)

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
func (c *Controller) Run() {
	log.Infof("%s starting on port %s", c.Name, c.Port)

	muxRouter, err := setupRouters(c.Routers)
	if err != nil {
		log.WithError(err).Fatalf("%s failed controller setup", c.Name)
	}

	c.Mux = muxRouter

	listen, err := net.Listen("tcp", "localhost:"+c.Port)
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on port %s", c.Port)
	}

	server := grpc.NewServer()

	pb.RegisterControllerServiceServer(server, c)

	if err := server.Serve(listen); err != nil {
		log.WithError(err).Fatalf("%s failed to serve", c.Name)
	}

	log.Warnf("%s terminated", c.Name)
}

// AttachAndForward - Attach will take an element controller and attach it to an
// existing router. The Routes will forward given requests to the Controller as
// they are received.
func (c *Controller) AttachAndForward(m *mux.Router) {
	log.Infof("%s attaching forwarder", c.Name)

	handlerFunc := func(c *Controller) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c.Send(w, r)
		}
	}

	for _, router := range c.Routers {
		for _, route := range router.Routes() {
			m.
				Name(route.Name).
				Methods(route.Method).
				Path(route.Path).
				Handler(handlerFunc(c))
		}
	}
}

// Serve - Serve will call on the controller to act as an HTTP server instead
// of a GRPC Server. This is helpful for testing the capabilities of the
// controller without DP-API hosting the server
func (c *Controller) ListenAndServe() {
	log.Info("%s setup listen and serve", c.Name)
	m, err := setupRouters(c.Routers)
	if err != nil {
		log.WithError(err).Fatalf("%s failed in setup", c.Name)
		return
	}

	c.Mux = m

	// Permissive handling of Cross Origin Resource Sharing
	// for debug. This allows us access the server from other
	// web hosting platforms.
	cr := cors.AllowAll()

	log.Infof("%s listen and serve", c.Name)
	log.Fatal(http.ListenAndServe(":8080", cr.Handler(m)))
}
