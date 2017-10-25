package pure

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/leesper/holmes"
	"github.com/urfave/negroni"
)

var (
	// JSONPlugin checks whether it is a JSON request.
	JSONPlugin = negroni.HandlerFunc(jsonMiddleware)
	// MultipartFormPlugin checks whether it is a multipart/form-data request.
	MultipartFormPlugin = negroni.HandlerFunc(multipartFormMiddleware)
	// CORSPlugin handle CORS request, see https://github.com/rs/cors
	CORSPlugin = negroni.HandlerFunc(corsMiddleware)
	// RecoverPanicPlugin recovers and records the stack info when panic occurred.
	// This prevents web server from crashing.
	RecoverPanicPlugin = negroni.HandlerFunc(recoverPanicMiddleWare)
	// LoggerPlugin adds statistic information for every request.
	LoggerPlugin = negroni.HandlerFunc(loggerMiddleware)
)

func jsonMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get(ContentType) != ApplicationJSON {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		fmt.Fprintf(w, "%s not %s", ContentType, ApplicationJSON)
	} else {
		next(w, r)
	}
}

// FIXME: add special treatments for MultipartFormData
func multipartFormMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Header.Get(ContentType) != MultipartFormData {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		fmt.Fprintf(w, "%s not %s", ContentType, MultipartFormData)
	} else {
		next(w, r)
	}
}

func corsMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if "OPTIONS" == r.Method {
		headers := w.Header()
		origin := r.Header.Get("Origin")
		// Always set Vary headers
		// see https://github.com/rs/cors/issues/10,
		//     https://github.com/rs/cors/commit/dbdca4d95feaa7511a46e6f1efb3b3aa505bc43f#commitcomment-12352001
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")

		reqMethod := r.Header.Get("Access-Control-Request-Method")
		reqHeaders := strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",")
		headers.Set("Access-Control-Allow-Origin", origin) // trust all sources
		// Spec says: Since the list of methods can be unbounded, simply returning the method indicated
		// by Access-Control-Request-Method (if supported) can be enough
		headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod)) // methods allowed
		if len(reqHeaders) > 0 {
			// Spec says: Since the list of headers can be unbounded, simply returning supported headers
			// from Access-Control-Request-Headers can be enough
			headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", ")) // custom headers allowed
		}
		w.WriteHeader(http.StatusOK)
	} else {
		next(w, r)
	}
}

func recoverPanicMiddleWare(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			traceInfo := make([]byte, 0<<15)
			n := runtime.Stack(traceInfo, true)
			w.WriteHeader(http.StatusInternalServerError)
			m := fmt.Sprintf("%s %s", err, string(traceInfo[:n]))
			fmt.Fprint(w, m)
		}
	}()

	next(w, r)
}

func loggerMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	rw := newResponseWriterWrapper(w)
	next(rw, r)

	logging := fmt.Sprintf("%s -- %v %s %s %s %s - %s %v",
		r.RemoteAddr,
		start,
		r.Method,
		r.URL.Path,
		r.Proto,
		http.StatusText(rw.StatusCode()),
		r.Header.Get("User-Agent"),
		time.Since(start))

	holmes.Infoln(logging)
}

// a wrapper to get HTTP status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{w, http.StatusOK}
}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriterWrapper) Write(bs []byte) (int, error) {
	return rw.ResponseWriter.Write(bs)
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterWrapper) StatusCode() int {
	return rw.statusCode
}
