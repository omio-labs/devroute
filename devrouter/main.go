package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
)

// DevRouter is a wrapper around *http.Server
type DevRouter struct {
	*http.Server
}

// Start starts DevRouter server
func (d DevRouter) Start() {
	d.ListenAndServe()
}

// Stop shutdowns DevRouter server
func (d DevRouter) Stop() {
	d.Shutdown(context.Background())
}

// NewDevRouter creates a new DevRouter
func NewDevRouter() DevRouter {
	port := ":8080"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = ":" + p
	}

	http.HandleFunc("/", ProxyToDev)
	http.HandleFunc("/_healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	return DevRouter{
		&http.Server{
			Addr:    port,
			Handler: http.DefaultServeMux,
		}}
}

// ProxyToDev parses the contract contained in `X-Devroute` HTTP header
// and proxies the request to a developer laptop.
func ProxyToDev(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Header["X-Devroute"]; !ok {
		fmt.Fprintln(os.Stderr, `Missing "X-Devroute" header`)
		http.Error(w, `Missing "X-Devroute" header`, http.StatusBadRequest)
		return
	}

	var contract map[string]string
	err := json.Unmarshal([]byte(r.Header.Get("X-Devroute")), &contract)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Failed to parse "X-Devroute" header: %v\n`, err)
		http.Error(w, fmt.Sprintf(`Failed to parse "X-Devroute" header: %v`, err),
			http.StatusBadRequest)
		return
	}

	matched := r.Header.Get("X-Devroute-Matched")
	if _, ok := contract[matched]; !ok {
		fmt.Fprintf(os.Stderr, `Matched service "%s" was not found in contract: %v\n`, matched, contract)
		http.Error(w, fmt.Sprintf(`Matched service "%s" was not found in contract: %v`, matched, contract),
			http.StatusBadRequest)
		return
	}

	host, _, err := net.SplitHostPort(contract[matched])
	if err != nil {
		fmt.Fprintf(os.Stderr, `Failed to parse host:port on key "%s": %v\n`, matched, err)
		http.Error(w, fmt.Sprintf(`Failed to parse host:port on key "%s": %v`, matched, err),
			http.StatusBadRequest)
		return
	}
	if !isPrivateIP(host) {
		fmt.Fprintf(os.Stderr, "Proxying to non-private IPs (%s) is forbidden\n", host)
		http.Error(w, fmt.Sprintf("Proxying to non-private IPs (%s) is forbidden", host),
			http.StatusForbidden)
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = contract[matched]
			r.URL.Scheme = "http"
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			fmt.Fprintf(os.Stderr, "http: proxy error: %v\n", err)
			rw.WriteHeader(http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
}

func isPrivateIP(host string) bool {
	networks := []string{
		// IPv4:
		"192.168.0.0/16",
		"172.16.0.0/12",
		"10.0.0.0/8",
		// IPv6:
		"fd00::/8",
		// loopback: (for unit test & debug)
		"127.0.0.0/8",
	}

	a := net.ParseIP(host)
	if a == nil {
		return false
	}

	for _, network := range networks {
		_, sub, _ := net.ParseCIDR(network)
		if sub.Contains(a) {
			// match
			return true
		}
	}

	// no match
	return false
}

func main() {
	devrouter := NewDevRouter()
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go devrouter.Start()
	<-stop
	devrouter.Stop()
}
