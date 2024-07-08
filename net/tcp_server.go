package net

import (
	"log"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Port                string
	wgServer            sync.WaitGroup
	wgConn              sync.WaitGroup
	connLock            sync.Mutex
	listener            net.Listener
	connections         sync.Map
	connCloseNotifyChan chan *TCPConnection
}

func (server *TCPServer) Start() {
	server.connCloseNotifyChan = make(chan *TCPConnection, 100)
	listener, err := net.Listen("tcp", "0.0.0.0:"+server.Port)
	if err != nil {
		log.Println("server start err:", err)
		return
	}
	server.listener = listener
}

func (server *TCPServer) run() {
	server.wgServer.Add(1)
	defer server.wgServer.Done()
	var delay time.Duration
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if _, ok := err.(net.Error); ok && err.(net.Error).Timeout() {
				if delay == 0 {
					delay = 2 * time.Millisecond
				} else {
					delay *= 2
				}
				if max := 1 * time.Second; delay > max {
					delay = max
				}
				time.Sleep(delay)
				continue
			}
			return
		}
		tcpConn := NewTCPConnection(conn, server.connCloseNotifyChan)
		server.connections.Store(tcpConn.conn.RemoteAddr(), tcpConn)
		server.wgConn.Add(1)
		if OnConnect != nil {
			OnConnect(tcpConn)
		}
	}
}

func (server *TCPServer) closedConnWatcher() {
	for {
		tcpConn := <-server.connCloseNotifyChan
		if !tcpConn.IsConnected() {
			if _, ok := server.connections.Load(tcpConn.conn.RemoteAddr()); ok {
				if OnDisconnect != nil {
					OnDisconnect(tcpConn)
				}
				server.wgConn.Done()
				server.connections.Delete(tcpConn.conn.RemoteAddr())
			}
		}
	}
}
