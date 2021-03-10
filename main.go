package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"spider/core/engine"
	"spider/core/scheduler"

	//"io"
	"log"
	"net"
	"runtime"
	"strings"
)

func main()  {

	// default spider engine
	var e = &engine.Engine{
		Scheduler: &scheduler.QueuedScheduler{},
		WorkerCount: 10000,
		RequestProcess: engine.Worker,
	}
	e.Run()

	l, err := net.Listen("tcp", "127.0.0.1:9501")
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
		go handle(c, e)
	}
}

const defaultBufferSize = 16 * 1024
func handle(conn net.Conn, e *engine.Engine)  {

	log.Printf("TCP: new client(%s)", conn.RemoteAddr())
	reader := bufio.NewReaderSize(conn, defaultBufferSize)
	writer := bufio.NewWriterSize(conn, defaultBufferSize)

	typ := make([]byte, 3)
	n, err := reader.Read(typ)
	if n != 3 || err != nil {
		_, _ = writer.WriteString("Forbidden.")
		_ = writer.Flush()
		_ = conn.Close()
		return
	}

	switch string(typ) {
	case "PUB":
		for {
			length := make([]byte, 2)
			_, err := reader.Read(length)
			if err != nil {
				log.Printf("read conn err(%s) - %s", conn.RemoteAddr(), err)
				_ = conn.Close()
				return
			}
			dataLen := binary.BigEndian.Uint16(length)
			data := make([]byte, dataLen)
			_, err = reader.Read(data)
			request := engine.Request{}
			err = json.Unmarshal(data, &request)
			request.UID = uuid.NewV4().String()
			e.Scheduler.Submit(request)
			_, err = writer.WriteString(request.UID)
			_ = writer.Flush()
			if err != nil {
				log.Printf("PUB write conn err(%s) - %s", conn.RemoteAddr(), err)
				engine.Result.Push(&request)
				_ = conn.Close()
				return
			}
		}
	case "SUB":
		for {
			r := engine.Result.Pop()
			ret, err := json.Marshal(r)
			if err != nil {
				log.Printf("marshal err(%s) - %v", conn.RemoteAddr(), err)
				continue
			}
			_, err = writer.Write(ret)
			_ = writer.Flush()
			if err != nil {
				log.Printf("SUB write conn err(%s) - %s", conn.RemoteAddr(), err)
				engine.Result.Push(r)
				_ = conn.Close()
				return
			}
		}
	default:
		_, _ = writer.WriteString("Forbidden.")
		_ = writer.Flush()
		_ = conn.Close()
	}

}

