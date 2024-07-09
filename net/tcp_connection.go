package net

import (
	"net"
	"time"
)

type TCPConnection struct {
	id              any
	conn            net.Conn
	ioWorker        *worker
	isConnected     bool
	closeNotifyChan chan *TCPConnection
}

func NewTCPConnection(conn net.Conn, closeNotifyChan chan *TCPConnection) *TCPConnection {
	tcp := new(TCPConnection)
	tcp.conn = conn
	tcp.closeNotifyChan = closeNotifyChan
	tcp.isConnected = true
	tcp.ioWorker = newIOWorker(tcp)
	return tcp
}

func (t *TCPConnection) GetId() any {
	return t.id
}

func (t *TCPConnection) SetId(id any) {
	t.id = id
}

func (t *TCPConnection) IsConnected() bool {
	return t.isConnected
}

func (t *TCPConnection) Close() {
	t.isConnected = false
	t.conn.Close()
	t.closeNotifyChan <- t
}

func (t *TCPConnection) Read(b []byte) (n int, err error) {
	return t.conn.Read(b)
}

func (t *TCPConnection) Write(b []byte) (n int, err error) {
	return t.conn.Write(b)
}

func (t *TCPConnection) GetLastReadTime() time.Time {
	return t.ioWorker.receiver.lastReadTime
}

func (t *TCPConnection) GetLastWriteTime() time.Time {
	return t.ioWorker.sender.lastWriteTime
}
