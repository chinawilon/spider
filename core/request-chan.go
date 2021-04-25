package core

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

type RequestChan chan Response

func (rc RequestChan) Push(response Response)  {
	for {
		select {
		case rc <- response:
			return
		default:
			_ = <-rc
		}
	}
}

func (rc RequestChan) Pop() Response {
	return <- rc
}

// Prevent having too many files open at the same time
// var rateLimited = time.Tick(1 * time.Millisecond)

func (rc RequestChan) Work(r Request){

	// <- rateLimited

	rp := Response{
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
	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		rp.Error = err
		rc.Push(rp)
		return
	}

	body, _ := ioutil.ReadAll(response.Body)
	rp.Body = body
	rp.StatusCode = response.StatusCode
	rc.Push(rp)

}

