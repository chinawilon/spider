package core

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
)

type Service struct {
	Engine Engineer
	Protocol Protocol
	Host string
	Port int
}

func (s *Service) addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Service) Run()  {
	// Engine run
	s.Engine.Run()
	// Server run
	s.Server()
}

func (s *Service) Server()  {
	l, err := net.Listen("tcp", s.addr())
	log.Printf("listen %v", s.addr())
	if err != nil {
		log.Fatalf("net listen err - %s", err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				runtime.Gosched()
				continue
			}
			if !strings.Contains(err.Error(), "use of closed network connection") {
				_ = fmt.Errorf("listener.Accept() error - %s", err)
			}
			break
		}
		s.p.Handle(c)
	}
}

