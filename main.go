package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"log"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"spider/core/engine"
	"spider/core/processor"
	"spider/core/scheduler"
	"strings"
)

func main()  {


	f1, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f1.Close() // 该例子中暂忽略错误处理
	if err := pprof.StartCPUProfile(f1); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()


	f2, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f2.Close()
	if err := pprof.WriteHeapProfile(f2); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}


	// spider engine
	proc := make(processor.RequestChan, 100)
	var e = &engine.Engine{
		Scheduler: &scheduler.QueuedScheduler{},
		WorkerCount: 10000,
		Processor: proc,
	}
	e.Run()

	addr := "127.0.0.1:9501"
	l, err := net.Listen("tcp", addr)
	log.Printf("listen %v", addr)
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
	log.Printf("handle type : %v", string(typ))
	switch string(typ) {
	case "PUB":
		pub(reader, writer, conn, e)
	case "SUB":
		sub(writer, conn, e)
	default:
		shutdown(writer, conn)
	}
}

func shutdown(writer *bufio.Writer, conn net.Conn)  {
	_, _ = writer.WriteString("Forbidden.")
	_ = writer.Flush()
	_ = conn.Close()
}

func pub(reader *bufio.Reader, writer *bufio.Writer, conn net.Conn, e *engine.Engine) {
	for {
		length := make([]byte, 2)
		n, err := reader.Read(length)
		if n != 2 || err != nil {
			log.Printf("read conn err(%s) - %s", conn.RemoteAddr(), err)
			shutdown(writer, conn)
			return
		}
		dataLen := binary.BigEndian.Uint16(length)
		data := make([]byte, dataLen)
		n, err = reader.Read(data)
		if err != nil || n != int(dataLen) {
			log.Printf("payload length(%d) expect(%d)", n, dataLen)
			shutdown(writer, conn)
			return
		}
		request := engine.Request{}
		err = json.Unmarshal(data, &request)
		if err != nil {
			log.Printf("payload data err - %s", err)
			shutdown(writer, conn)
			return
		}
		request.UID = uuid.NewV4().String()
		_, err = writer.WriteString(request.UID)
		_ = writer.Flush()
		if err != nil {
			log.Printf("PUB write conn err(%s) - %s", conn.RemoteAddr(), err)
			_ = conn.Close()
			return
		}
		// every thing is ok
		e.Scheduler.Submit(request)
	}
}

func sub(writer *bufio.Writer, conn net.Conn, e *engine.Engine)  {
	for {
		r := e.Processor.Pop()
		payload, err := json.Marshal(r)
		if err != nil {
			log.Printf("marshal err(%s) - %v", conn.RemoteAddr(), err)
			continue
		}
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(len(payload)))
		_, err = writer.Write(b)
		_, err = writer.Write(payload)
		_ = writer.Flush()
		if err != nil {
			log.Printf("SUB write conn err(%s) - %s", conn.RemoteAddr(), err)
			e.Processor.Push(r)
			_ = conn.Close()
			return
		}
	}
}