package main

import (
	"spider/core"
)

func main()  {
	// spider engine
	proc := make(core.RequestChan, 10000)
	e := &core.Engine{
		Scheduler: &core.QueuedScheduler{},
		WorkerCount: 10000,
		Processor: proc,
	}
	// Core Service
	s := core.Service{
		Engine: e,
		Server: &core.Server{
			Host: "0.0.0.0",
			Port: 8080,
			Protocol: &core.V1{
				Engine: e,
			},
		},
	}
	s.Run()
}

