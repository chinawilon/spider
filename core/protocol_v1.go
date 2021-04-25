package core

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"io"
	"log"
	"net"
)

const defaultBufferSize = 16 * 1024

type V1 struct {
	Conn net.Conn
	Writer *bufio.Writer
	Reader *bufio.Reader
	Engine *Engine
}

func (p *V1) Handle(conn net.Conn)  {
	p.Reader = bufio.NewReaderSize(conn, defaultBufferSize)
	p.Writer = bufio.NewWriterSize(conn, defaultBufferSize)
	p.Conn = conn
	p.Run()
}

func (p *V1) Run()  {
	log.Printf("TCP: new client(%s)", p.Conn.RemoteAddr())

	typ := make([]byte, 4)
	_, err := io.ReadFull(p.Reader, typ)
	if err != nil {
		_, _ = p.Writer.WriteString("Forbidden.")
		_ = p.Writer.Flush()
		_ = p.Conn.Close()
		return
	}
	log.Printf("handle type : %v", string(typ))
	switch string(typ) {
	case "PUB":
		p.Pub()
	case "SUB":
		p.Sub()
	default:
		p.Shutdown()
	}
}

func (p *V1) Shutdown()  {
	_, _ = p.Writer.WriteString("Forbidden.")
	_ = p.Writer.Flush()
	_ = p.Conn.Close()
}

func (p *V1) Pub() {
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.Reader, length)
		if err != nil {
			log.Printf("read conn err(%s) - %s", p.Conn.RemoteAddr(), err)
			p.Shutdown()
			return
		}
		dataLen := binary.BigEndian.Uint16(length)
		data := make([]byte, dataLen)
		n, err := io.ReadFull(p.Reader, data)
		if err != nil || n != int(dataLen) {
			log.Printf("payload length(%d) expect(%d)", n, dataLen)
			p.Shutdown()
			return
		}
		request := Request{}
		err = json.Unmarshal(data, &request)
		if err != nil {
			log.Printf("payload data err - %s", err)
			p.Shutdown()
			return
		}
		request.UID = uuid.NewV4().String()
		_, err = p.Writer.WriteString(request.UID)
		_ = p.Writer.Flush()
		if err != nil {
			log.Printf("PUB write conn err(%s) - %s", p.Conn.RemoteAddr(), err)
			_ = p.Conn.Close()
			return
		}
		// every thing is ok
		p.Engine.Submit(request)
	}
}

func (p *V1) Sub()  {
	for {
		r := p.Engine.Pop()
		payload, err := json.Marshal(r)
		if err != nil {
			log.Printf("marshal err(%s) - %v", p.Conn.RemoteAddr(), err)
			continue
		}
		b := make([]byte, 4)
		binary.BigEndian.PutUint16(b, uint16(len(payload)))
		_, err = p.Writer.Write(b)
		_, err = p.Writer.Write(payload)
		_ = p.Writer.Flush()
		if err != nil {
			log.Printf("SUB write conn err(%s) - %s", p.Conn.RemoteAddr(), err)
			p.Engine.Push(r)
			_ = p.Conn.Close()
			return
		}
	}
}