package engine

type RequestChan chan *Request

var Result = make(RequestChan, 10000)

func (r RequestChan) Push(request *Request)  {
	for {
		select {
		case r <- request:
			return
		default:
			<- r // if block remove the oldest one
		}
	}
}

func (r RequestChan) Pop() *Request {
	return <- r
}

