package pure

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run starts a web server.
func Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	srv := newServer(addr, router)
	return srv.ListenAndServe()
}

// RunTLS starts a TLS-based web server.
func RunTLS(port int, cert, key string) error {
	addr := fmt.Sprintf(":%d", port)
	srv := newServer(addr, router)
	return srv.ListenAndServeTLS(cert, key)
}

// RunElegant starts a web server which can be shutdown gracefully with a timout.
func RunElegant(port int, timeout time.Duration) error {
	addr := fmt.Sprintf(":%d", port)
	srv := newServer(addr, router)
	go handleSignal(srv, timeout)
	return srv.ListenAndServe()
}

// RunTLSElegant starts a TLS-based web server which can be shutdown gracefully with a timeout.
func RunTLSElegant(port int, timeout time.Duration, cert, key string) error {
	addr := fmt.Sprintf(":%d", port)
	srv := newServer(addr, router)
	go handleSignal(srv, timeout)
	return srv.ListenAndServeTLS(cert, key)
}

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

func handleSignal(srv *http.Server, timeout time.Duration) {
	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
