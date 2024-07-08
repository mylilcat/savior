package launcher

import (
	"github.com/mylilcat/savior/net"
)

const (
	KCP int = iota
	TCP
)

var (
	port      string
	netProto  int
	kcpServer *net.KCPServer
)

func SetPort(serverPort string) {
	port = serverPort
}

func ServerStart() {
	if netProto == KCP {
		kcpServer = new(net.KCPServer)
		kcpServer.Port = port
		kcpServer.Start()
	}
}
