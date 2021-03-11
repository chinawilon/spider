package engine

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

// Prevent having too many files open at the same time
var rateLimiter = time.Tick(1 * time.Millisecond)

func Worker(r Request){

	<- rateLimiter

	rp := &Response{
		UID: r.UID,
	}

	client := http.Client{
		Timeout: r.Timeout * time.Second,
	}

	request, err := http.NewRequest(r.Method, r.Uri, bytes.NewReader(r.Body))
	if err != nil {
		rp.Error = err
		Result.Push(rp)
		return
	}

	request.Header = http.Header{}
	for i, v := range r.Header {
		request.Header[i] = []string{v}
	}

	response, err := client.Do(request)
	if err != nil {
		rp.Error = err
		Result.Push(rp)
		return
	}

	if response.Body != nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		rp.Body = body
		rp.StatusCode = response.StatusCode
		Result.Push(rp)
		return
	}

	// other side
	Result.Push(rp)
}
