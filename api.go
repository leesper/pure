package pure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leesper/holmes"
	"github.com/urfave/negroni"
)

// definitions for HTTP headers.
const (
	ContentType       = "Content-Type"
	ApplicationJSON   = "application/json"
	MultipartFormData = "multipart/form-data"
	AcceptLanguage    = "Accept-Language"
	Authorization     = "Authorization"
	UserAgent         = "User-Agent"
	AllowOrigin       = "Access-Control-Allow-Origin"
)

var (
	apiMap = make(map[string]*api)
	router = mux.NewRouter()
)

// Handler defines the interface of JSON request handler.
type Handler interface {
	Handle(ctx context.Context) interface{}
}

// HandlerFunc is an adapter to allow the use of ordinary functions as handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context) interface{}

// Handle calls HandlerFunc itself.
func (hf HandlerFunc) Handle(ctx context.Context) interface{} {
	return hf(ctx)
}

// APIBuilder is responsible for building APIs. Please don't operate it directly,
// use chaining calls instead.
type APIBuilder struct {
	version     string
	class       string
	name        string
	method      string
	handlerVal  reflect.Value
	middlewares []negroni.Handler
}

type api struct {
	version     string
	class       string
	name        string
	method      string
	handlerVal  reflect.Value
	middlewares []negroni.Handler
}

func newAPI(version, class, name, method string, handlerVal reflect.Value, middlewares []negroni.Handler) *api {
	return &api{
		version:     version,
		class:       class,
		name:        name,
		method:      method,
		handlerVal:  handlerVal,
		middlewares: middlewares,
	}
}

func (a *api) uri() string {
	// FIXME: handle when version, class and name are empty string
	return fmt.Sprintf("/%s/%s/%s", a.version, a.class, a.name)
}

// API returns a new APIBuilder for buiding API.
func API(name string) *APIBuilder {
	return &APIBuilder{
		name: name,
	}
}

// Version defines the API version.
func (b *APIBuilder) Version(v string) *APIBuilder {
	b.version = v
	return b
}

// Class defines the API class.
func (b *APIBuilder) Class(c string) *APIBuilder {
	b.class = c
	return b
}

// Post marks API an HTTP POST request.
func (b *APIBuilder) Post() *APIBuilder {
	b.method = http.MethodPost
	return b
}

// Get marks API an HTTP GET request.
func (b *APIBuilder) Get() *APIBuilder {
	b.method = http.MethodGet
	return b
}

// Use adds a series of middleware plugins to the API.
func (b *APIBuilder) Use(handlers ...negroni.Handler) *APIBuilder {
	b.middlewares = append(b.middlewares, handlers...)
	return b
}

// HandleFunc defines the handler function of request.
func (b *APIBuilder) HandleFunc(f func(ctx context.Context) interface{}) *APIBuilder {
	hf := HandlerFunc(f)
	b.handlerVal = reflect.ValueOf(hf)
	return b
}

// Handle defines the handler of request.
func (b *APIBuilder) Handle(h Handler) *APIBuilder {
	b.handlerVal = reflect.ValueOf(h)
	return b
}

// Done triggers APIBuilder to build and register an API.
func (b *APIBuilder) Done() {
	a := newAPI(b.version, b.class, b.name, b.method, b.handlerVal, b.middlewares)
	apiMap[a.uri()] = a
	middles := negroni.New(b.middlewares...)
	if a.method == http.MethodPost {
	}
	switch a.method {
	case http.MethodGet, http.MethodPost:
		router.Handle(a.uri(), middles.With(negroni.Wrap(http.HandlerFunc(httpRequestDispatcher)))).Methods(http.MethodOptions, a.method)
	default:
		holmes.Infof("HTTP %s ignored", a.method)
	}
}

func httpRequestDispatcher(w http.ResponseWriter, r *http.Request) {
	a, ok := apiMap[r.URL.Path]
	if !ok {
		holmes.Errorln("cannot find handler for", r.URL.Path)
		return
	}

	// construct a brand new handler of a.handlerType
	var handler Handler
	switch a.handlerVal.Type().Kind() {
	case reflect.Func:
		handler = a.handlerVal.Interface().(Handler)
	case reflect.Struct:
		handlerVal := reflect.New(a.handlerVal.Type()).Elem()
		// inflate handler with data
		handler = handlerVal.Addr().Interface().(Handler)
		if handlerVal.NumField() > 0 {
			err := json.NewDecoder(r.Body).Decode(handler)
			if err != nil {
				holmes.Errorln(err)
				return
			}
		}
	}

	// collect all the context info.
	ctx := r.Context()
	for k, v := range r.Header {
		ctx = context.WithValue(ctx, k, v)
	}

	// perform the business logic, return the result.
	output := handler.Handle(ctx)
	if err := jsonify(w, output); err != nil {
		holmes.Errorln(err)
	}
}

func jsonify(w http.ResponseWriter, message interface{}) error {
	w.Header().Set(ContentType, ApplicationJSON)
	w.Header().Set(AllowOrigin, "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		return err
	}
	return nil
}
