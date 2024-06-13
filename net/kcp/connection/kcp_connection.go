package connection

import "github.com/xtaci/kcp-go/v5"

type KCPConnection struct {
	ID              interface{}
	conn            *kcp.UDPSession
	ioWorker        *Worker
	isConnected     bool
	closeNotifyChan chan *KCPConnection
}

func NewKCPConnection(conn *kcp.UDPSession, closeNotifyChan chan *KCPConnection) *KCPConnection {
	kcp := new(KCPConnection)
	kcp.conn = conn
	kcp.closeNotifyChan = closeNotifyChan
	kcp.isConnected = true
	kcp.ioWorker = NewIOWorker(kcp)
	return kcp
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
