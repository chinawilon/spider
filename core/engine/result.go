package engine

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RequestChan chan *Response

func (rc RequestChan) Push(response *Response)  {
	for {
		select {
		case rc <- response:
			return
		default:
			log.Printf("remove oldest one: %v", <-rc)
		}
	}
}

func (rc RequestChan) Pop() *Response {
	return <- rc
}


func (rc RequestChan) Worker(r *Request){

	// Prevent having too many files open at the same time
	<- time.Tick(1 * time.Millisecond)

	rp := &Response{
		UID: r.UID,
	}

	client := http.Client{
		Timeout: r.Timeout * time.Second,
	}

	request, err := http.NewRequest(r.Method, r.Uri, bytes.NewReader(r.Body))
	if err != nil {
		rp.Error = err
		rc.Push(rp)
		return
	}

	request.Header = http.Header{}
	for i, v := range r.Header {
		request.Header[i] = []string{v}
	}

	response, err := client.Do(request)
	if err != nil {
		rp.Error = err
		rc.Push(rp)
		return
	}

	if response.Body != nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		rp.Body = body
		rp.StatusCode = response.StatusCode
		rc.Push(rp)
		return
	}

	// other side
	rc.Push(rp)
}
