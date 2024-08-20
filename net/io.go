package net

import (
	"log"
	"time"
)

// connection io worker
type worker struct {
	sender   *sender
	receiver *receiver
}

func newIOWorker(c Connection, connTyp string) *worker {
	ioWorker := new(worker)
	ioWorker.sender = newSender(c, connTyp)
	ioWorker.receiver = newReceiver()
	go ioWorker.receiver.receiverRunning(c)
	return ioWorker
}

type receiver struct {
	lastReadTime time.Time
}

func newReceiver() *receiver {
	r := new(receiver)
	r.lastReadTime = time.Now()
	return r
}

// read bytes
func (r *receiver) receiverRunning(c Connection) {
	for {
		if !c.IsConnected() {
			break
		}
		buf := make([]byte, 4096)
		n, err := c.Read(buf)
		if err != nil {
			log.Println("Savior receive err:", err)
			break
		}
		if OnRead != nil {
			OnRead(c, buf[:n])
		}
		r.lastReadTime = time.Now()
	}
	c.Close()
}

type sender struct {
	typ           string
	conn          Connection
	lastWriteTime time.Time
}

func newSender(c Connection, connType string) *sender {
	s := new(sender)
	s.conn = c
	s.typ = connType
	s.lastWriteTime = time.Now()
	return s
}

// send bytes
func (s *sender) send(bytes []byte) {
	switch s.typ {
	case "tcp":
		if !s.conn.IsConnected() {
			return
		}
		_, err := s.conn.Write(bytes)
		if err != nil {
			s.conn.Close()
			return
		}
		s.lastWriteTime = time.Now()
	case "kcp":
		if !s.conn.IsConnected() {
			return
		}
		go func() {
			_, err := s.conn.Write(bytes)
			if err != nil {
				s.conn.Close()
				return
			}
		}()
		s.lastWriteTime = time.Now()
	}
}
