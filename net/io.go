package net

import (
	"github.com/mylilcat/savior/util"
	"log"
	"time"
)

// connection io worker
type worker struct {
	sender   *sender
	receiver *receiver
}

func newIOWorker(c Connection, connTyp string, handler *Handler) *worker {
	ioWorker := new(worker)
	ioWorker.sender = newSender(c, connTyp)
	ioWorker.receiver = newReceiver()
	go ioWorker.sender.senderRunning(c)
	go ioWorker.receiver.receiverRunning(c, handler.onRead)
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
func (r *receiver) receiverRunning(c Connection, onRead func(conn Connection, data []byte)) {
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
		if onRead != nil {
			onRead(c, buf[:n])
		}
		r.lastReadTime = time.Now()
	}
	c.Close()
}

type sender struct {
	typ           string
	conn          Connection
	sendChan      chan []byte
	lastWriteTime time.Time
}

func newSender(c Connection, connType string) *sender {
	s := new(sender)
	s.conn = c
	s.typ = connType
	s.sendChan = make(chan []byte, 100)
	s.lastWriteTime = time.Now()
	return s
}

func (s *sender) senderRunning(c Connection) {
	for bytes := range s.sendChan {
		if bytes != nil {
			break
		}
		if !c.IsConnected() {
			break
		}
		switch s.typ {
		case "tcp":
			_, err := s.conn.Write(bytes)
			if err != nil {
				s.conn.Close()
				break
			}
			s.lastWriteTime = time.Now()
		case "kcp":
			util.KcpSend(func() {
				_, err := s.conn.Write(bytes)
				if err != nil {
					s.conn.Close()
					return
				}
			})
			s.lastWriteTime = time.Now()
		}
	}
	if len(s.sendChan) > 0 {
		for range s.sendChan {
			continue
		}
	}
	close(s.sendChan)
	c.Close()
}

func (s *sender) send(data []byte) {
	s.sendChan <- data
}
