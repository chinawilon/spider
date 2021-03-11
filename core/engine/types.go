package engine

import "time"

type Request struct {
	UID string `json:"uid"`
	Uri string `json:"uri"`
	Method string `json:"method"`
	Header map[string]string `json:"header"`
	Body []byte `json:"body"`
	Timeout time.Duration `json:"timeout"`
}

type Response struct {
	UID string `json:"uid"`
	StatusCode int `json:"status_code"`
	Error error `json:"error"`
	Body []byte `json:"body"`
}


type Processor func(Request)

type Scheduler interface {
	ReadyNotifier
	Submit(Request)
	WorkerChan() chan Request
	Run()
}

type ReadyNotifier interface {
	WorkerReady(chan Request)
}
