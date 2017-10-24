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

// APIBuilder is responsible for building APIs. Please don't operate it directly,
// use chaining calls instead.
type APIBuilder struct {
	version     string
	class       string
	name        string
	method      string
	handlerType reflect.Type
	middlewares []negroni.Handler
}

type api struct {
	version     string
	class       string
	name        string
	method      string
	handlerType reflect.Type
	middlewares []negroni.Handler
}

func newAPI(version, class, name, method string, handlerType reflect.Type, middlewares []negroni.Handler) *api {
	return &api{
		version:     version,
		class:       class,
		name:        name,
		method:      method,
		handlerType: handlerType,
		middlewares: middlewares,
	}
}

func (a *api) uri() string {
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

// Post makes API an HTTP POST request.
func (b *APIBuilder) Post() *APIBuilder {
	b.method = http.MethodPost
	return b
}

// Handle defines the handler of request.
func (b *APIBuilder) Handle(h Handler) *APIBuilder {
	b.handlerType = reflect.TypeOf(h)
	fmt.Println(reflect.TypeOf(h))
	return b
}

// Done triggers APIBuilder to build and register an API.
func (b *APIBuilder) Done() {
	a := newAPI(b.version, b.class, b.name, b.method, b.handlerType, b.middlewares)
	apiMap[a.uri()] = a
	middles := negroni.New(b.middlewares...)
	if a.method == http.MethodPost {
	}
	switch a.method {
	case http.MethodPost:
		router.Handle(a.uri(), middles.With(negroni.Wrap(http.HandlerFunc(postHandler)))).Methods(http.MethodOptions, a.method)
	case http.MethodGet:
		router.Handle(a.uri(), middles.With(negroni.Wrap(http.HandlerFunc(getHandler)))).Methods(http.MethodOptions, a.method)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	a, ok := apiMap[r.URL.Path]
	if !ok {
		holmes.Errorln("cannot find handler for", r.URL.Path)
		return
	}

	// construct a brand new handler of a.handlerType
	handlerVal := reflect.New(a.handlerType).Elem()

	// check and construct the input parameter of handler
	if handlerVal.NumField() <= 0 {
		holmes.Errorln("no fields found in type", a.handlerType)
		return
	}
	input := reflect.New(handlerVal.Field(0).Type()).Elem().Addr().Interface()
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		holmes.Errorln(err)
		return
	}
	handlerVal.Field(0).Set(reflect.ValueOf(input).Elem())

	// collect all the context info.
	ctx := r.Context()
	for k, v := range r.Header {
		ctx = context.WithValue(ctx, k, v)
	}

	// perform the business logic, return the result.
	handler := handlerVal.Interface().(Handler)
	output := handler.Handle(ctx)
	if err = jsonify(w, output); err != nil {
		holmes.Errorln(err)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {}

func jsonify(w http.ResponseWriter, message interface{}) error {
	w.Header().Set(ContentType, ApplicationJSON)
	w.Header().Set(AllowOrigin, "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		return err
	}
	return nil
}
