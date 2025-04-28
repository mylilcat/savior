package net

import (
	saviorLog "github.com/mylilcat/savior/log"
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Port                string
	wgServer            sync.WaitGroup
	wgConn              sync.WaitGroup
	listener            net.Listener
	connections         sync.Map
	connCloseNotifyChan chan *TCPConnection
	Handler             *Handler
	IdleMonitor         *IdleMonitor
}

func (server *TCPServer) Start() {
	server.connections = sync.Map{}
	server.connCloseNotifyChan = make(chan *TCPConnection, 100)
	listener, err := net.Listen("tcp", "0.0.0.0:"+server.Port)
	if err != nil {
		saviorLog.Print("server start err:", err)
		return
	}
	server.listener = listener
	go server.run()
	go server.closedConnWatcher()
	if server.IdleMonitor != nil {
		server.IdleMonitor.idleMonitoring(&server.connections, server.Handler.onIdle)
	}
}

func (server *TCPServer) run() {
	server.wgServer.Add(1)
	defer server.wgServer.Done()
	var delay time.Duration
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			var e net.Error
			if errors.As(err, &e) && err.(net.Error).Timeout() {
				if delay == 0 {
					delay = 2 * time.Millisecond
				} else {
					delay *= 2
				}
				if duration := 1 * time.Second; delay > duration {
					delay = duration
				}
				time.Sleep(delay)
				continue
			}
			return
		}
		tcpConn := NewTCPConnection(conn, server.connCloseNotifyChan, server.Handler)
		server.connections.Store(tcpConn.conn.RemoteAddr(), tcpConn)
		server.wgConn.Add(1)
		if server.Handler.onConnect != nil {
			server.Handler.onConnect(tcpConn)
		}
	}
}

func (server *TCPServer) closedConnWatcher() {
	for {
		tcpConn := <-server.connCloseNotifyChan
		if !tcpConn.IsConnected() {
			if _, loaded := server.connections.LoadAndDelete(tcpConn.conn.RemoteAddr()); loaded {
				if server.Handler.onDisconnect != nil {
					server.Handler.onDisconnect(tcpConn)
				}
				server.wgConn.Done()
			}
		}
	}
}
