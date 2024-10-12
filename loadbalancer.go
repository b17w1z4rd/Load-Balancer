package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type LoadBalancer struct {
	backends []*url.URL
	current  int
	mutex    sync.Mutex
}

func NewLoadBalancer(backendURLs []string) *LoadBalancer {
	backends := make([]*url.URL, len(backendURLs))
	for i, backendURL := range backendURLs {
		u, err := url.Parse(backendURL)
		if err != nil {
			log.Fatalf("Invalid URL %s: %v", backendURL, err)
		}
		backends[i] = u
	}
	return &LoadBalancer{backends: backends}
}

func (lb *LoadBalancer) getNextBackend() *url.URL {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	backend := lb.backends[lb.current]
	lb.current = (lb.current + 1) % len(lb.backends)
	return backend
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextBackend()
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.ServeHTTP(w, r)
}

func main() {
	backendURLs := []string{
		"http://localhost:3001",
		"http://localhost:3002",
	}
	lb := NewLoadBalancer(backendURLs)

	server := http.Server{
		Addr: ":3000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lb.handleRequest(w, r)
		}),
	}

	fmt.Println("Load balancer running on http://localhost:3000")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
