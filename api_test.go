package pure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/leesper/holmes"
)

type helloInput struct {
	Name string `json:"name"`
}

type helloOutput struct {
	Greetings string `json:"greetings"`
}

type helloHandler struct {
	Input helloInput
}

func (h helloHandler) Handle(ctx context.Context) interface{} {
	return helloOutput{fmt.Sprintf("hello, %s", h.Input.Name)}
}

func TestAPIBuilding(t *testing.T) {
	defer holmes.Start().Stop()
	API("hello").Version("apiv1").Class("hello").Post().Handle(helloHandler{}).Done()
	expected := "/apiv1/hello/hello"
	a, ok := apiMap[expected]
	if !ok {
		t.Fatalf("API %s not registered", expected)
	}

	if a.uri() != expected {
		t.Fatalf("returned: %s, expected: %s", a.uri(), expected)
	}

	if !strings.Contains(a.uri(), a.version) {
		t.Fatalf("returned: %s, expected: %s", a.version, expected)
	}

	if !strings.Contains(a.uri(), a.class) {
		t.Fatalf("returned: %s, expected: %s", a.version, expected)
	}

	if !strings.Contains(a.uri(), a.name) {
		t.Fatalf("returned: %s, expected: %s", a.version, expected)
	}

	if a.method != http.MethodPost {
		t.Fatalf("returned: %s, expected: %s", a.method, http.MethodPost)
	}

	if a.handlerType != reflect.TypeOf(helloHandler{}) {
		t.Fatalf("returned: %s, expected: %s", a.handlerType, reflect.TypeOf(helloHandler{}))
	}

	if len(a.middlewares) != 0 {
		t.Fatalf("returned: %v, expected: %v", a.middlewares, []string{})
	}

	var hasMethodPost, hasMethodOptions, hasPath bool
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		if path == expected {
			hasPath = true
		}

		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		for _, m := range methods {
			if m == http.MethodOptions {
				hasMethodOptions = true
			}
			if m == http.MethodPost {
				hasMethodPost = true
			}
		}

		return nil
	})

	if !hasMethodPost {
		t.Fatal("router should have method POST")
	}

	if !hasMethodOptions {
		t.Fatal("router should have method OPTIONS")
	}

	if !hasPath {
		t.Fatal("router should have url", expected)
	}

	input := helloInput{"Foo"}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, a.uri(), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	postHandler(rec, req)

	if status := rec.Code; status != http.StatusOK {
		t.Fatalf("returned status code: %d, expected: %d", status, http.StatusOK)
	}

	if contentType := rec.Header().Get(ContentType); contentType != ApplicationJSON {
		t.Fatalf("returned content type: %s, expected: %s", contentType, ApplicationJSON)
	}

	output := helloOutput{}
	if err = json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
		t.Fatal(err)
	}

	if output.Greetings != "hello, Foo" {
		t.Errorf("returned: %s, expected: %s", output.Greetings, "hello, Foo")
	}
}
