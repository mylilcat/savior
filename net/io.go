package net

import (
	"log"
	"time"
)

type worker struct {
	sender   *sender
	receiver *receiver
}

func newIOWorker(c Connection) *worker {
	ioWorker := new(worker)
	ioWorker.sender = newSender()
	ioWorker.receiver = newReceiver()
	go ioWorker.receiver.receiverRunning(c)
	go ioWorker.sender.senderRunning(c)
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
	sendChan      chan []byte
	lastWriteTime time.Time
}

func newSender() *sender {
	s := new(sender)
	s.sendChan = make(chan []byte, 100)
	s.lastWriteTime = time.Now()
	return s
}

func (s *sender) send(bytes []byte) {
	s.sendChan <- bytes
}

func (s *sender) senderRunning(c Connection) {
	for bytes := range s.sendChan {
		if bytes == nil {
			break
		}
		if !c.IsConnected() {
			break
		}
		switch c.(type) {
		case *TCPConnection:
			_, err := c.Write(bytes)
			if err != nil {
				log.Println("Savior write err:", err)
				break
			}
			s.lastWriteTime = time.Now()
		case *KCPConnection:
			var errFlag bool
			go func() {
				_, err := c.Write(bytes)
				if err != nil {
					log.Println("Savior write err:", err)
					errFlag = true
				}
				s.lastWriteTime = time.Now()
			}()
			if errFlag {
				break
			}
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
