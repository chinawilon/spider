package engine

type Engine struct {
	Scheduler Scheduler
	WorkerCount int
	RequestProcess Processor
}

func (e *Engine) Run()  {
	e.Scheduler.Run()
	for i := 0; i < e.WorkerCount; i++ {
		e.createWorker(e.Scheduler.WorkerChan(), e.Scheduler)
	}
}

func (e *Engine) createWorker(in chan Request, ready ReadyNotifier)  {
	go func() {
		for {
			ready.WorkerReady(in)
			request := <- in
			e.RequestProcess(&request)
		}
	}()
}