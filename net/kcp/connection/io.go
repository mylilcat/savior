package connection

import "log"

type Worker struct {
	sender   *Sender
	receiver *Receiver
	kcp      *KCPConnection
}

func NewIOWorker(kcp *KCPConnection) *Worker {
	ioWorker := new(Worker)
	ioWorker.sender = NewSender()
	ioWorker.receiver = NewReceiver()
	go ioWorker.receiver.receiverRunning(kcp)
	go ioWorker.sender.senderRunning(kcp)
	return ioWorker
}

type Receiver struct {
}

func NewReceiver() *Receiver {
	return new(Receiver)
}

func (r *Receiver) receiverRunning(k *KCPConnection) {
	for {
		if !k.isConnected {
			break
		}
		buf := make([]byte, 4096)
		if k.isConnected {
			n, err := k.conn.Read(buf)
			if err != nil {
				log.Println("receive err:", err)
				break
			}
			if OnRead != nil {
				OnRead(k, buf[:n])
			}
		}
	}
	k.Close()
}

type Sender struct {
	sendChan chan []byte
}

func NewSender() *Sender {
	sender := new(Sender)
	sender.sendChan = make(chan []byte, 100)
	return sender
}

func (s *Sender) send(bytes []byte) {
	s.sendChan <- bytes
}

func (s *Sender) senderRunning(kcp *KCPConnection) {
	for bytes := range s.sendChan {
		if bytes == nil {
			break
		}
		if !kcp.isConnected {
			break
		}
		_, err := kcp.conn.Write(bytes)
		if err != nil {
			log.Println("send err:", err)
			break
		}
	}
	kcp.Close()
}
