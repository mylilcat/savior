package launcher

import (
	"github.com/mylilcat/savior/net"
)

const (
	TCP int = iota
	KCP
)

var (
	port  string
	proto int
)

func SetProto(p int) {
	proto = p
}

func SetPort(serverPort string) {
	port = serverPort
}

func ServerStart() {
	switch proto {
	case TCP:
		s := new(net.TCPServer)
		s.Port = port
		s.Start()
	case KCP:
		s := new(net.KCPServer)
		s.Port = port
		s.Start()
	}
}
