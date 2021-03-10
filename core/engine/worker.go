package engine

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Prevent having too many files open at the same time
var rateLimiter = time.Tick(1 * time.Millisecond)

func Worker(r Request){

	<- rateLimiter

	client := http.Client{
		Timeout: r.Timeout * time.Second,
	}

	request, err := http.NewRequest(r.Method, r.Uri, bytes.NewReader(r.Body))
	if err != nil {
		log.Printf("new request err : %v", err)
		r.Response.Error = err
		Result.Push(&r)
		return
	}

	request.Header = http.Header{}
	for i, v := range r.Header {
		request.Header[i] = []string{v}
	}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("client do request err : %v", err)
		r.Response.Error = err
		Result.Push(&r)
		return
	}

	if response.Body != nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		r.Response.Body = body
		r.Response.StatusCode = response.StatusCode
		Result.Push(&r)
		return
	}

	// other side
	Result.Push(&r)
}
