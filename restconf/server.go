package restconf

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/stock"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/parser"
	"github.com/freeconf/yang/source"
)

type webApp struct {
	endpoint string
	homeDir  string
	homePage string
}

type Server struct {
	Web                     *stock.HttpServer
	webApps                 []webApp
	Ver                     string
	main                    device.Device
	devices                 device.Map
	ypath                   source.Opener
	UnhandledRequestHandler http.HandlerFunc
	Filters                 []RequestFilter
	OnlyStrictCompliance    bool
}

var ErrBadAddress = errors.New("expected format: http://server/restconf[=device]/operation/module:path")

type RequestFilter func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error)

func NewServer(d *device.Local) *Server {
	m := NewHttpServe(d)
	d.Add("fc-restconf", Node(m, d.SchemaSource()))
	return m
}

func NewHttpServe(d *device.Local) *Server {
	m := &Server{
		ypath: d.SchemaSource(),
	}
	m.ServeDevice(d)
	return m
}

func (srv *Server) ServeDevice(d device.Device) error {
	srv.main = d
	return nil
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acceptType := MimeType(r.Header.Get("Accept"))
	compliance := Simplified
	ctx := context.WithValue(r.Context(), ComplianceContextKey, compliance)

	if fc.DebugLogEnabled() && r.Body != nil {
		content, _ := ioutil.ReadAll(r.Body)
		fc.Debug.Print(string(content))
	}

	op1, deviceId, p := shiftOptionalParamWithinSegment(r.URL, '=', '/')
	device, _ := srv.findDevice(deviceId)

	switch op1 {
	case "restconf":
		op2, p := shift(p, '/')
		r.URL = p
		switch op2 {
		case "data":
			srv.serve(compliance, ctx, device, w, r, endpointData, acceptType)
		case "operations":
			srv.serve(compliance, ctx, device, w, r, endpointOperations, acceptType)
		case "schema":
			srv.serveSchema(compliance, ctx, w, r, device.SchemaSource(), acceptType)
		}
	}
}

const (
	endpointData = iota
	endpointOperations
	endpointStreams
	endpointSchema
)

func (srv *Server) serveSchema(compliance ComplianceOptions, ctx context.Context, w http.ResponseWriter, r *http.Request, ypath source.Opener, accept MimeType) {
	modName, p := shift(r.URL, '/')
	r.URL = p
	m, _ := parser.LoadModule(ypath, modName)
	ylib, _ := parser.LoadModule(ypath, "fc-yang")
	
	hndlr := &browserHandler{browser: nodeutil.Schema(m, ylib)}
	hndlr.ServeHTTP(compliance, ctx, w, r, endpointSchema)
}

func (srv *Server) serve(compliance ComplianceOptions, ctx context.Context, d device.Device, w http.ResponseWriter, r *http.Request, endpointId int, accept MimeType) {
	if hndlr, p := srv.shiftBrowserHandler(compliance, r, d, w, r.URL, accept); hndlr != nil {
		r.URL = p
		hndlr.ServeHTTP(compliance, ctx, w, r, endpointId)
	}
}

func (srv *Server) findDevice(deviceId string) (device.Device, error) {
	if deviceId == "" {
		return srv.main, nil
	}
	return srv.devices.Device(deviceId)
}

func (srv *Server) shiftBrowserHandler(compliance ComplianceOptions, r *http.Request, d device.Device, w http.ResponseWriter, orig *url.URL, accept MimeType) (*browserHandler, *url.URL) {
	if module, p := shift(orig, ':'); module != "" {
		if browser, _ := d.Browser(module); browser != nil {
			return &browserHandler{browser: browser}, p
		}
	}
	return nil, orig
}