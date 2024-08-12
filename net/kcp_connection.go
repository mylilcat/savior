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
	kcp := new(KCPConnection)
	kcp.conn = conn
	kcp.closeNotifyChan = closeNotifyChan
	kcp.isConnected = true
	conn.SetNoDelay(1, 15, 2, 1)
	conn.SetWindowSize(1024, 1024)
	conn.SetWriteDelay(false)
	conn.SetACKNoDelay(true)
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

func (k *KCPConnection) Write(b []byte) (n int, err error) {
	err = k.conn.SetWriteDeadline(time.Now().Add(5 * time.Millisecond))
	if err != nil {
		return 0, err
	}
	return k.conn.Write(b)
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
