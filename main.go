package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type proxy struct {
	registryHost string
	registryPort int
	portusHost   string
	portusPort   int
	port         int
	mux          *http.ServeMux
	registryURL  *url.URL
	portusURL    *url.URL
}

func main() {
	proxy := proxy{}
	proxy.setupVars()
	proxy.setupURLs()
	proxy.setupWeb()
	proxy.RunProxy()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (proxy *proxy) setupVars() {
	log.Print("Setting up variables")
	var err error
	proxy.registryHost = getEnv("REGISTRY_HOST", "registry")
	proxy.registryPort, err = strconv.Atoi(getEnv("REGISTRY_PORT", "5000"))
	if err != nil {
		log.Fatal("REGISTRY_PORT must be int")
	}
	proxy.portusHost = getEnv("PORTUS_HOST", "portus")
	proxy.portusPort, err = strconv.Atoi(getEnv("PORTUS_PORT", "3000"))
	if err != nil {
		log.Fatal("PORTUS_PORT must be int")
	}
	proxy.port, err = strconv.Atoi(getEnv("PROXY_PORT", "8080"))
	if err != nil {
		log.Fatal("PROXY_PORT must be int")
	}
	log.Print("Finished setting up variables")
}

func (proxy *proxy) setupWeb() {
	log.Print("Setting up proxy")
	proxy.mux = http.NewServeMux()
	proxy.mux.HandleFunc("/", proxy.handleRequest)
	log.Print("Finished setting up proxy")
}

func (proxy *proxy) setupURLs() {
	var err error
	proxy.portusURL, err = url.Parse(fmt.Sprintf("http://%s:%d", proxy.portusHost, proxy.portusPort))
	if err != nil {
		log.Fatal("Unable to create url for proxyToRegistry")
	}
	proxy.registryURL, err = url.Parse(fmt.Sprintf("http://%s:%d", proxy.registryHost, proxy.registryPort))
	if err != nil {
		log.Fatal("Unable to create url for proxy")
	}
}

func (proxy *proxy) RunProxy() {
	log.Print("Starting proxy.")
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", proxy.port),
		Handler: proxy.mux,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing proxy.")
}

func (proxy *proxy) handleRequest(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/v2") {
		if !strings.HasPrefix(request.URL.Path, "/v2/token") && !strings.HasPrefix(request.URL.Path, "/v2/webhooks") {
			httputil.NewSingleHostReverseProxy(proxy.registryURL).ServeHTTP(writer, request)
			return
		}
	}
	httputil.NewSingleHostReverseProxy(proxy.portusURL).ServeHTTP(writer, request)
}
