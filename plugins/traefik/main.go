package traefik

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
)

type SablierMiddleware struct {
	client  *http.Client
	request *http.Request
	next    http.Handler
}

// New function creates the configuration
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	req, err := config.BuildRequest(name)

	if err != nil {
		return nil, err
	}

	return &SablierMiddleware{
		request: req,
		client:  &http.Client{},
		next:    next,
	}, nil
}

func (sm *SablierMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	sablierRequest := sm.request.Clone(context.TODO())

	resp, err := sm.client.Do(sablierRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	backendReady := false

	if resp.Header.Get("X-Sablier-Session-Status") == "ready" {
		// Check if the backend already received request data
		trace := &httptrace.ClientTrace{
			WroteHeaders: func() {
				backendReady = true
				fmt.Println("------------- WroteHeaders")
			},
			WroteRequest: func(httptrace.WroteRequestInfo) {
				backendReady = true
				fmt.Println("------------- WroteRequest")
			},
		}
		newCtx := httptrace.WithClientTrace(req.Context(), trace)

		sm.next.ServeHTTP(rw, req.WithContext(newCtx))
	}

	if backendReady == false {
		fmt.Println("------------- backend not Ready")
		forward(resp, rw)
	}
}

func forward(resp *http.Response, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(rw, resp.Body)
}
