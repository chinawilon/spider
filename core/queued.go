package core


type QueuedScheduler struct {
	requestChan chan Request
	workerChan chan chan Request
}

func (s *QueuedScheduler) WorkerChan() chan Request {
	return make(chan Request)
}

func (s *QueuedScheduler) Submit(r Request) {
	s.requestChan <- r
}

func (s *QueuedScheduler) WorkerReady(w chan Request) {
	s.workerChan <- w
}

func (s *QueuedScheduler) Run()  {
	s.workerChan = make(chan chan Request)
	s.requestChan = make(chan Request)

	go func() {
		var requestQ []Request
		var workerQ []chan Request

		for {
			var activeRequest Request
			var activeWorker chan Request

			if len(requestQ) > 0 && len(workerQ) > 0 {
				activeRequest = requestQ[0]
				activeWorker = workerQ[0]
			}

			select {
			case r := <-s.requestChan:
				requestQ = append(requestQ, r)
			case w := <-s.workerChan:
				workerQ = append(workerQ, w)
			case activeWorker <- activeRequest: // do it
				workerQ = workerQ[1:]
				requestQ = requestQ[1:]
			}
		}

	}()
}