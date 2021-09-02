package ec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
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
	GET_METHOD    = http.MethodGet
	POST_METHOD   = http.MethodPost
	PATCH_METHOD  = http.MethodPatch
	DELETE_METHOD = http.MethodDelete
)

// Route -
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
	Port    int
	Version string
	Routers Routers

	options   Options
	router    *mux.Router
	processor ControllerProcessor
}

// Options -
type Options struct {
	Http    bool
	Port    int
	Log     bool
	Verbose bool
}

func NewDefaultOptions() *Options {
	return &Options{Http: false, Port: 0, Log: false, Verbose: false}
}

func BindFlags(fs *flag.FlagSet) *Options {
	opts := NewDefaultOptions()
	fs.BoolVar(&opts.Http, "http", opts.Http, "Setup element controller as standard http server")
	fs.IntVar(&opts.Port, "port", opts.Port, "Override element controller port")
	fs.BoolVar(&opts.Log, "log", opts.Log, "Enable server logging")
	fs.BoolVar(&opts.Verbose, "verbose", opts.Verbose, "Enable verbose logging")

	return opts
}

// ResponseWriter -
type ResponseWriter struct {
	StatusCode int
	Hdr        http.Header
	Buffer     *bytes.Buffer
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		StatusCode: http.StatusOK,
		Hdr:        make(http.Header),
		Buffer:     new(bytes.Buffer),
	}
}

func (r *ResponseWriter) Header() http.Header {
	return r.Hdr
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	return r.Buffer.Write(b)
}

func (r *ResponseWriter) WriteHeader(code int) {
	r.StatusCode = code
}

// checkAPI -
func (c *Controller) checkApiVersion(api string) error {
	if c.Version != api {
		return status.Errorf(codes.Unimplemented, "%s: Unsupported API Version", c.Name)
	}
	return nil
}

// initialize - Initialize the controller with a new mux.Router and ensure all
// the controller routers are succesfully initialized.
func (c *Controller) initialize(opts *Options) error {
	c.options = *opts

	c.processor = NewControllerProcessor(opts.Http)
	c.router = mux.NewRouter().StrictSlash(true)

	for _, api := range c.Routers {
		if err := api.Init(); err != nil {
			return err
		}
	}

	for _, api := range c.Routers {
		if err := api.Start(); err != nil {
			return err
		}
	}

	return nil
}

type ControllerProcessor interface {
	Run(c *Controller, options Options) error
	Send(c *Controller, w http.ResponseWriter, r *http.Request)
}

func NewControllerProcessor(http bool) ControllerProcessor {
	if http {
		return &HttpControllerProcessor{}
	}

	return &GrpcControllerProcessor{}
}

type HttpControllerProcessor struct {
	client http.Client
}

func (*HttpControllerProcessor) Run(c *Controller, options Options) error {
	if options.Log {
		c.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				var rlog = log.WithFields(log.Fields{
					"Method": r.Method,
					"URL":    r.RequestURI,
				})

				if options.Verbose && r.Method == POST_METHOD {
					body, _ := ioutil.ReadAll(r.Body)
					r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
					rlog.WithField("Request", string(body)).Infof("Http Request: %s %s", r.Method, r.URL)
				}

				start := time.Now()

				recorder := httptest.NewRecorder()
				next.ServeHTTP(recorder, r)

				status := recorder.Result().StatusCode

				w.WriteHeader(status)
				w.Write(recorder.Body.Bytes())

				if options.Verbose {
					rlog = rlog.WithField("Response", recorder.Body.String())
				}

				rlog.WithFields(log.Fields{
					"Status":      status,
					"ElapsedTime": time.Since(start).String(),
				}).Infof("Http Response: %d (%s)", status, http.StatusText(status))
			})
		})
	}

	// Permissive handling of Cross Origin Resource Sharing
	// for debug. This allows us access the server from other
	// web hosting platforms.
	crs := cors.AllowAll()

	address := fmt.Sprintf(":%d", c.Port)
	log.Infof("Starting HTTP Server at %s", address)
	return http.ListenAndServe(address, crs.Handler(c.router))
}

func (p *HttpControllerProcessor) Send(c *Controller, w http.ResponseWriter, r *http.Request) {
	rsp, _ := p.client.Do(r)

	w.WriteHeader(rsp.StatusCode)
	w.Header().Set("Content-Type", rsp.Header.Get("Content-Type"))
	io.Copy(w, rsp.Body)

	rsp.Body.Close()
}

type GrpcControllerProcessor struct{}

func (*GrpcControllerProcessor) Run(c *Controller, options Options) error {
	address := fmt.Sprintf(":%d", c.Port)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen on address %s", address)
	}

	server := grpc.NewServer()

	pb.RegisterControllerServiceServer(server, c)

	log.Infof("Starting GRPC Server at %s", address)
	return server.Serve(listen)
}

// Send - Send will send the parameters in the http request to the element controller
// using the GRPC and then handle the response from GRPC back out the http response
func (*GrpcControllerProcessor) Send(c *Controller, w http.ResponseWriter, r *http.Request) {

	// Initialize ClientRequest object
	grpcReq := client.ClientRequest{
		ElmntCntrlName: c.Name,
		ElmntCntrlPort: fmt.Sprintf(":%d", c.Port),
		HTTPwriter:     w,
	}

	// Record timestamp of request and reminder
	t := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	pfx := t.Format(time.RFC3339Nano)

	// Construct and send element controller request
	data := []byte{}
	if r.Body != nil {
		data, _ = ioutil.ReadAll(r.Body)
	}

	request := pb.ECTaskRequest{
		Api:       c.Version,
		Sender:    msgs.DPAPIname,
		Uri:       r.URL.String(),
		Method:    r.Method,
		JsonMsg:   string(data),
		Reminder:  reminder,
		Timestamp: pfx,
	}

	grpcReq.ProcessRequest(&request)
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
	c.router.ServeHTTP(res, req)

	if res.StatusCode != http.StatusOK {
		log.Warnf("Request faild with status %d", res.StatusCode)
		err = http.ErrNotSupported // TODO: Should have a encoding map from StatusCode to Err
	}

	return &pb.ECTaskResponse{
		Api:      c.Version,
		JsonData: string(res.Buffer.Bytes()),
	}, err
}

// HandlerFunc defines an http handler for a controller's routes. By default
// the controller's routers define handlers, but by using a custom global
// handler function one can override this behavior.
type HandlerFunc func(c *Controller) http.HandlerFunc

// Forward provides the means to pass a http request to an element controller
func Forward(c *Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Send(w, r)
	}
}

// Reject will refuse any and all requests received by the element controller
func Reject(c *Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

// Initialize an element controller with the given options, or nil for default
func (c *Controller) Init(opts *Options) error {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	if opts.Port != 0 {
		c.Port = opts.Port
	}

	return c.initialize(opts)
}

// Run - Run a controller with standard behavior - that is with GRPC server and
// request handling that operates by unpacking the GRPC request and
// forwardining it to the element controller's handlers.
func (c *Controller) Run() {
	if c.processor == nil {
		log.Fatalf("Controller %s must call Init() prior to run", c.Name)
	}
	c.Attach(c.router, nil)

	if err := c.processor.Run(c, c.options); err != nil {
		log.WithError(err).Fatalf("%s failed to run", c.Name)
	}
}

// Send a request to the element controller
func (c *Controller) Send(w http.ResponseWriter, r *http.Request) {
	c.processor.Send(c, w, r)
}

// Attach - Attach this controller to a existing router mux with the provided handler function
// This will add the element controller's defined routes, but will override the prefered route
// handler - allowing for interception of predefiend routes.
func (c *Controller) Attach(router *mux.Router, handlerFunc HandlerFunc) {

	for _, api := range c.Routers {
		for _, r := range api.Routes() {
			route := router.
				Name(r.Name).
				Path(r.Path).
				Methods(r.Method).
				Handler(r.HandlerFunc)

			if handlerFunc != nil {
				route.Handler(handlerFunc(c))
			}
		}
	}
}

// EncodeResponse -
func EncodeResponse(s interface{}, err error, w http.ResponseWriter) {

	if err != nil {
		// If the supplied error is of an Element Controller Controller Error type,
		// encode the response to a new error response packet.
		var e *controllerError
		if errors.As(err, &e) {
			w.WriteHeader(e.statusCode)
			s = NewErrorResponse(e, s)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	if s != nil {
		w.Header().Set("Content-Type", "application/json")
		response, err := json.Marshal(s)
		if err != nil {
			log.WithError(err).Error("Failed to marshal json response")
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write(response)
		if err != nil {
			log.WithError(err).Error("Failed to write json response")
			w.WriteHeader(http.StatusInternalServerError)
		}

	}
}