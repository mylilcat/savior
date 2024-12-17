package net

import (
	"github.com/mylilcat/savior/util"
	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go/v5"
	"log"
	"net"
	"sync"
	"time"
)

type KCPServer struct {
	Port                string
	wgServer            sync.WaitGroup
	wgConn              sync.WaitGroup
	listener            *kcp.Listener
	connections         sync.Map
	connCloseNotifyChan chan *KCPConnection
	Handler             *Handler
	IdleMonitor         *IdleMonitor
}

func (server *KCPServer) Start() {
	server.connCloseNotifyChan = make(chan *KCPConnection, 100)
	listener, err := kcp.ListenWithOptions("0.0.0.0:"+server.Port, nil, 0, 0)
	if err != nil {
		log.Println("server start err:", err)
		return
	}
	server.listener = listener
	util.KcpSendPoolInit()
	go server.run()
	go server.closedConnWatcher()
	if server.IdleMonitor != nil {
		server.IdleMonitor.idleMonitoring(&server.connections, server.Handler.onIdle)
	}
}

func (server *KCPServer) run() {
	server.wgServer.Add(1)
	defer server.wgServer.Done()
	var delay time.Duration
	for {
		conn, err := server.listener.AcceptKCP()
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
		kcpConn := NewKCPConnection(conn, server.connCloseNotifyChan, server.Handler)
		server.connections.Store(kcpConn.GetConv(), kcpConn)
		server.wgConn.Add(1)
		if server.Handler.onConnect != nil {
			server.Handler.onConnect(kcpConn)
		}
	}
}

func (server *KCPServer) closedConnWatcher() {
	for {
		kcpConn := <-server.connCloseNotifyChan
		if !kcpConn.IsConnected() {
			if _, loaded := server.connections.LoadAndDelete(kcpConn.GetConv()); loaded {
				if server.Handler.onDisconnect != nil {
					server.Handler.onDisconnect(kcpConn)
				}
				server.wgConn.Done()
			}
		}
	}
}
