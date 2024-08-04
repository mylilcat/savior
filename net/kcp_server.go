package net

import (
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
	connLock            sync.Mutex
	listener            *kcp.Listener
	connections         sync.Map
	connCloseNotifyChan chan *KCPConnection
}

func (server *KCPServer) Start() {
	server.connCloseNotifyChan = make(chan *KCPConnection, 100)
	listener, err := kcp.ListenWithOptions("0.0.0.0:"+server.Port, nil, 0, 0)
	if err != nil {
		log.Println("server start err:", err)
		return
	}
	server.listener = listener
	go server.run()
	go server.closedConnWatcher()
	if IdleMonitoring != nil {
		IdleMonitoring(&server.connections)
	}
}

func (server *KCPServer) run() {
	server.wgServer.Add(1)
	defer server.wgServer.Done()
	var delay time.Duration
	for {
		conn, err := server.listener.AcceptKCP()
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
		kcpConn := NewKCPConnection(conn, server.connCloseNotifyChan)
		server.connections.Store(kcpConn.GetConv(), kcpConn)
		server.wgConn.Add(1)
		if OnConnect != nil {
			OnConnect(kcpConn)
		}
	}
}

func (server *KCPServer) closedConnWatcher() {
	for {
		kcpConn := <-server.connCloseNotifyChan
		if !kcpConn.IsConnected() {
			if _, loaded := server.connections.LoadAndDelete(kcpConn.GetConv()); loaded && OnDisconnect != nil {
				OnDisconnect(kcpConn)
				server.wgConn.Done()
			}
		}
	}
}
