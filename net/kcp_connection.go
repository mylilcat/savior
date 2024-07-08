package net

import (
	"github.com/xtaci/kcp-go/v5"
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
	return k.conn.Write(b)
}
