package net

import (
	"log"
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
}

func newReceiver() *receiver {
	return new(receiver)
}

func (r *receiver) receiverRunning(c Connection) {
	for {
		if !c.IsConnected() {
			break
		}
		buf := make([]byte, 4096)
		n, err := c.Read(buf)
		if err != nil {
			log.Println("receive err:", err)
			break
		}
		if OnRead != nil {
			OnRead(c, buf[:n])
		}
	}
	c.Close()
}

type sender struct {
	sendChan chan []byte
}

func newSender() *sender {
	s := new(sender)
	s.sendChan = make(chan []byte, 100)
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
		_, err := c.Write(bytes)
		if err != nil {
			log.Println("send err:", err)
			break
		}
	}
	c.Close()
}
