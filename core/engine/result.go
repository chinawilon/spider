package engine

import "log"

type RequestChan chan *Response

var Result = make(RequestChan, 10000)

func (r RequestChan) Push(response *Response)  {
	for {
		select {
		case r <- response:
			return
		default:
			log.Printf("remove oldest one: %v", <-r)
		}
	}
}

func (r RequestChan) Pop() *Response {
	return <- r
}

