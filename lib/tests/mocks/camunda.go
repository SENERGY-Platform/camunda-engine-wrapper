package mocks

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
)

type Request struct {
	Url     string
	Payload []byte
	Err     error
}

func CamundaServer(ctx context.Context, wg *sync.WaitGroup) (url string, requests chan Request) {
	requests = make(chan Request, 100)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		pl, err := io.ReadAll(r.Body)
		requests <- Request{
			Url:     r.URL.String(),
			Payload: pl,
			Err:     err,
		}
		if strings.Contains(r.URL.Path, "/process-definition") {
			json.NewEncoder(w).Encode([]interface{}{})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"id": strconv.Itoa(rand.Int())})
		}
	}))
	url = ts.URL
	wg.Add(1)
	go func() {
		<-ctx.Done()
		ts.Close()
		close(requests)
		wg.Done()
	}()
	return
}

func CamundaServerWithResponse(ctx context.Context, wg *sync.WaitGroup, resp interface{}) (url string, requests chan Request) {
	requests = make(chan Request, 100)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		pl, err := io.ReadAll(r.Body)
		requests <- Request{
			Url:     r.URL.String(),
			Payload: pl,
			Err:     err,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	url = ts.URL
	wg.Add(1)
	go func() {
		<-ctx.Done()
		ts.Close()
		close(requests)
		wg.Done()
	}()
	return
}
