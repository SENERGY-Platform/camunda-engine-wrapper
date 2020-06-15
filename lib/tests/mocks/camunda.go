package mocks

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		pl, err := ioutil.ReadAll(r.Body)
		requests <- Request{
			Url:     r.URL.String(),
			Payload: pl,
			Err:     err,
		}
		json.NewEncoder(w).Encode(map[string]interface{}{})
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
