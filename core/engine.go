package core

type Engine struct {
	Scheduler   Scheduler
	WorkerCount int
	Processor   Processor
}

func (e *Engine) Run()  {
	e.Scheduler.Run()
	for i := 0; i < e.WorkerCount; i++ {
		e.createWorker(e.Scheduler.WorkerChan(), e.Scheduler)
	}
}

func (e *Engine) Submit(r Request) {
	e.Scheduler.Submit(r)
}

func (e *Engine) Pop() Response {
	return e.Processor.Pop()
}

func (e *Engine) Push(r Response)  {
	e.Processor.Push(r)
}

func (e *Engine) createWorker(in chan Request, ready ReadyNotifier)  {
	go func() {
		for {
			ready.WorkerReady(in)
			request := <- in
			e.Processor.Work(request)
		}
	}()
}