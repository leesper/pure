# Pure

A small-yet-beautiful pure JSON API Web Framework, 1.0.1.

小而美的纯JSON API Web开发框架，版本1.0.1

[![GitHub stars](https://img.shields.io/github/stars/leesper/pure.svg)](https://github.com/leesper/pure/stargazers)
[![GitHub license](https://img.shields.io/github/license/leesper/pure.svg)](https://github.com/leesper/pure)
[![GoDoc](https://godoc.org/github.com/leesper/pure?status.svg)](http://godoc.org/github.com/leesper/pure)

## Features

1. Small yet beautiful, pure JSON API support by default;

    小而美，默认支持纯JSON API

2. Using Web middlewares to extend its funcionality;

    使用Web中间件扩展框架功能

3. Extremely light-weight, easy to use, write less, behave more elegant;

    超轻量级，容易使用，写更少的代码，表现更加优雅

4. Supporting TLS;

    支持TLS

## Requirements

* Golang 1.8 and above
* [mux](https://github.com/gorilla/mux)
* [negroni](https://github.com/urfave/negroni)

## Installation

`go get -u -v github.com/leesper/pure`

## Example

### 1. A simple "hello, world" example:

```go
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
```

Use `curl -H "Content-Type: application/json" -X POST -d '{"name": "Fiona"}' http://localhost:5050/apiv1/hello/greeting` to get a `{"greeting":"hello, Fiona"}` back.

Use `curl -H "Content-Type: application/json" -X GET http://localhost:5050/apiv1/hello/greeting` to get a `{"greeting":"hi, guest"}` back.

### 2. Web middlewares supported

* JSONMiddle: A Content-Type checker for application/json
* MultipartFormMiddle: For multipart/form-data requests
* CORSMiddle: Cross-origin resource sharing
* RecoverPanicMiddle: Recover from server panic
* LoggerMiddle: logging about HTTP request details

## License

[MIT](https://choosealicense.com/licenses/mit/)

## Changelog

### 1.0.1

1. Replace Use(...) with With(...)
2. One can now declare an API which can handle both GET and POST request
3. Make the framework depending on the official log package
4. Add utility functions for retrieving HTTP method or multi-part form from context
5. Add functions for closing server gracefully

## More Documentation

[Pure - 小而美的JSON Web API框架](http://www.jianshu.com/p/fe5db94d8f51)

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.
