package net

import (
	"github.com/xtaci/kcp-go/v5"
	"time"
)

type KCPConnection struct {
	id              any
	conn            *kcp.UDPSession
	ioWorker        *worker
	isConnected     bool
	closeNotifyChan chan *KCPConnection
}

func NewKCPConnection(conn *kcp.UDPSession, closeNotifyChan chan *KCPConnection) *KCPConnection {
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetWindowSize(4096, 4096)
	conn.SetWriteDelay(false)
	conn.SetACKNoDelay(true)
	kcp := new(KCPConnection)
	kcp.conn = conn
	kcp.closeNotifyChan = closeNotifyChan
	kcp.isConnected = true
	kcp.ioWorker = newIOWorker(kcp)
	return kcp
}

func (k *KCPConnection) GetId() any {
	return k.id
}

func (k *KCPConnection) SetId(id any) {
	k.id = id
}

func (k *KCPConnection) IsConnected() bool {
	return k.isConnected
}

func (k *KCPConnection) GetConv() uint32 {
	return k.conn.GetConv()
}

func (k *KCPConnection) Close() {
	k.isConnected = false
	k.conn.Close()
	k.closeNotifyChan <- k
}

func (k *KCPConnection) Read(b []byte) (n int, err error) {
	return k.conn.Read(b)
}

// unofficial method
func (k *KCPConnection) Write(b []byte) (n int, err error) {
	nChan := make(chan int)
	eChan := make(chan error)
	go func() {
		num, e := k.conn.Write(b)
		nChan <- num
		eChan <- e
	}()
	n = <-nChan
	err = <-eChan
	return
}

func (k *KCPConnection) Send(b []byte) {
	if k.isConnected {
		k.ioWorker.sender.send(b)
	}

}

func (k *KCPConnection) GetLastReadTime() time.Time {
	return k.ioWorker.receiver.lastReadTime
}

func (k *KCPConnection) GetLastWriteTime() time.Time {
	return k.ioWorker.sender.lastWriteTime
}
