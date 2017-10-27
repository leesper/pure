package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/leesper/holmes"
	"github.com/leesper/pure"
)

func main() {
	defer holmes.Start().Stop()
	pure.API("greeting").Version("apiv1").Class("hello").Get().Post().Handle(hello{}).With(
		pure.JSONMiddle,
		pure.LoggerMiddle,
	).Done()
	holmes.Errorln(pure.Run(5050))
}

type hello struct {
	Name string `json:"name"`
}

func (h hello) Handle(ctx context.Context) interface{} {
	switch pure.HTTPMethod(ctx) {
	case http.MethodGet:
		return helloRsp{fmt.Sprintf("hi, guest")}
	case http.MethodPost:
		return helloRsp{fmt.Sprintf("hello, %s", h.Name)}
	}
	return helloRsp{fmt.Sprintf("Sorry, I don't know your %s", pure.HTTPMethod(ctx))}
}

type helloRsp struct {
	Greeting string `json:"greeting"`
}

// command: curl -H "Content-Type: application/json" -X POST -d '{"name": "Fiona"}' http://localhost:5050/apiv1/hello/greeting
// command: curl -H "Content-Type: application/json" -X GET http://localhost:5050/apiv1/hello/greeting
