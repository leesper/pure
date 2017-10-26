package pure

import (
	"fmt"
	"net/http"
)

// Run starts a web server.
func Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, router)
}

// RunTLS starts a TLS-based web server.
func RunTLS(port int, cert, key string) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServeTLS(addr, cert, key, router)
}
