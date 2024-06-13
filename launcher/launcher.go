package launcher

import (
	"github.com/mylilcat/savior/net/kcp/connection"
	"github.com/mylilcat/savior/net/server"
)

const (
	KCP int = iota
	TCP
)

var (
	port      string
	netProto  int
	kcpServer *server.KCPServer
)

func SetKcpConnectHandler(fun func(connection *connection.KCPConnection)) {
	connection.OnConnect = fun
}

func SetKcpDisconnectHandler(fun func(connection *connection.KCPConnection)) {
	connection.OnDisconnect = fun
}

func SetKcpReadHandler(fun func(connection *connection.KCPConnection, data []byte)) {
	connection.OnRead = fun
}

func SetKcpIdleHandler(fun func(connection *connection.KCPConnection)) {
	connection.OnIdle = fun
}

func SetPort(serverPort string) {
	port = serverPort
}

func ServerStart() {
	if netProto == KCP {
		kcpServer = new(server.KCPServer)
		kcpServer.Port = port
		kcpServer.Start()
	}
}
