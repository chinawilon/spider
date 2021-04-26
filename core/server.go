package core

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
)


type Server struct {
	Host string
	Port int
	Protocol Protocol
}

func (s *Server) addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Server) Run()  {
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
		s.Protocol.Handle(c)
	}
}
