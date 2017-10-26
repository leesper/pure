package main

import (
	"context"
	"fmt"

	"github.com/leesper/holmes"
	"github.com/leesper/pure"
)

func main() {
	defer holmes.Start().Stop()
	pure.API("greeting").Version("apiv1").Class("hello").Post().Handle(hello{}).Use(
		pure.JSONMiddle,
		pure.LoggerMiddle,
	).Done()
	holmes.Errorln(pure.Run(5050))
}

type hello struct {
	Name string `json:"name"`
}

func (h hello) Handle(ctx context.Context) interface{} {
	return helloRsp{fmt.Sprintf("hello, %s", h.Name)}
}

type helloRsp struct {
	Greeting string `json:"greeting"`
}

// command: curl -H "Content-Type: application/json" -X POST -d '{"name": "Fiona"}' http://localhost:5050/apiv1/hello/greeting
