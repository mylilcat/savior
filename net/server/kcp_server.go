package server

import (
	"github.com/mylilcat/savior/net/kcp/connection"
	"github.com/xtaci/kcp-go/v5"
	"log"
	"sync"
)

type KCPServer struct {
	Port                string
	wgServer            sync.WaitGroup
	wgConn              sync.WaitGroup
	connLock            sync.Mutex
	listener            *kcp.Listener
	connections         sync.Map
	connCloseNotifyChan chan *connection.KCPConnection
}

func (server *KCPServer) Start() {
	server.connCloseNotifyChan = make(chan *connection.KCPConnection, 100)
	listener, err := kcp.ListenWithOptions("0.0.0.0:"+server.Port, nil, 0, 0)
	if err != nil {
		log.Println("server start err:", err)
		return
	}
	server.listener = listener
	go server.run()
	go server.closedConnWatcher()
}

func (server *KCPServer) run() {
	server.wgServer.Add(1)
	defer server.wgServer.Done()
	for {
		conn, err := server.listener.AcceptKCP()
		if err != nil {
			log.Println("server listening err:", err)
			conn.Close()
			return
		}
		kcpConn := connection.NewKCPConnection(conn, server.connCloseNotifyChan)
		server.connections.Store(kcpConn.GetConv(), kcpConn)
		server.wgConn.Add(1)
		if connection.OnConnect != nil {
			connection.OnConnect(kcpConn)
		}
	}
}

func (server *KCPServer) closedConnWatcher() {
	for {
		kcpConn := <-server.connCloseNotifyChan
		if !kcpConn.IsConnected() {
			if _, ok := server.connections.Load(kcpConn.GetConv()); ok {
				if connection.OnDisconnect != nil {
					connection.OnDisconnect(kcpConn)
				}
				server.wgConn.Done()
				server.connections.Delete(kcpConn.GetConv())
			}

		}
	}
}
