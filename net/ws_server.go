package net

import (
	"github.com/gorilla/websocket"
	saviorLog "github.com/mylilcat/savior/log"
	"net"
	"net/http"
	"sync"
)

type WSServer struct {
	Port                string
	wgServer            sync.WaitGroup
	wgConn              sync.WaitGroup
	connections         sync.Map
	connCloseNotifyChan chan *WSConnection
	listener            net.Listener
	Handler             *Handler
	wsHandler           *wsHandler
	IdleMonitor         *IdleMonitor
	MessageType         int
}

type wsHandler struct {
	upgrader websocket.Upgrader
	s        *WSServer
}

func (server *WSServer) Start() {
	server.connections = sync.Map{}
	server.connCloseNotifyChan = make(chan *WSConnection, 100)
	server.wsHandler = &wsHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		s: server,
	}
	listener, err := net.Listen("tcp", "0.0.0.0:"+server.Port)
	if err != nil {
		saviorLog.Print("server start err: %v", err)
		return
	}
	server.listener = listener
	go server.run()
	go server.closedConnWatcher()
	if server.IdleMonitor != nil {
		server.IdleMonitor.idleMonitoring(&server.connections, server.Handler.OnIdle)
	}
}

func (server *WSServer) run() {
	httpServer := &http.Server{}
	httpServer.Handler = server.wsHandler
	httpServer.MaxHeaderBytes = 1024
	serveErr := httpServer.Serve(server.listener)
	if serveErr != nil {
		saviorLog.Print("savior server serve err: %v", serveErr)
		return
	}
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		saviorLog.Print("upgrade err: %v", err)
		return
	}

	wsConn := NewWSConnection(conn, h.s.connCloseNotifyChan, h.s.Handler, h.s.MessageType)
	h.s.connections.Store(wsConn.conn.RemoteAddr(), wsConn)
	h.s.wgConn.Add(1)
	if h.s.Handler.OnConnect != nil {
		h.s.Handler.OnConnect(wsConn)
	}
}

func (server *WSServer) closedConnWatcher() {
	for {
		wsConn := <-server.connCloseNotifyChan
		if !wsConn.IsConnected() {
			if _, loaded := server.connections.LoadAndDelete(wsConn.conn.RemoteAddr()); loaded {
				if server.Handler.OnDisconnect != nil {
					server.Handler.OnDisconnect(wsConn)
				}
				server.wgConn.Done()
			}

		}
	}
}
