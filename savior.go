package savior

import (
	"github.com/gorilla/websocket"
	"github.com/mylilcat/savior/log"
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/service"
	"github.com/mylilcat/savior/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	TCP = "tcp"
	KCP = "kcp"
	WS  = "ws"
)

var enableDebugLog bool

type Savior struct {
	Port                 string
	Proto                string
	IdleMonitor          *net.IdleMonitor
	Handler              *net.Handler
	WebSocketMessageType int
}

func New() *Savior {
	s := new(Savior)
	s.Handler = &net.Handler{}
	return s
}

func (s *Savior) Start(services ...*service.Service) {
	if enableDebugLog {
		log.NewSaviorLogger()
	}
	for _, srv := range services {
		service.Register(srv)
	}
	service.ServicesRun()
	s.ServerStart()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	service.ServicesStop()
}

func EnableDebugLog() {
	enableDebugLog = true
}

func (s *Savior) BindPort(port string) {
	s.Port = port
}

func (s *Savior) SetProto(proto string) {
	s.Proto = proto
}

func (s *Savior) SetWebSocketMessageType(webSocketMessageType int) {
	s.WebSocketMessageType = webSocketMessageType
}

func (s *Savior) SetIdleMonitor(readIdle int64, writeIdle int64, unit time.Duration) {
	if readIdle < 0 || writeIdle < 0 || !util.IsTimeUnitValid(unit) {
		return
	}
	s.IdleMonitor = &net.IdleMonitor{
		ReadIdle:  readIdle,
		WriteIdle: writeIdle,
		Unit:      unit,
	}
}

func (s *Savior) SetOnConnectHandler(f func(c net.Connection)) {
	s.Handler.OnConnect = f
}

func (s *Savior) SetOnDisconnectHandler(f func(c net.Connection)) {
	s.Handler.OnDisconnect = f
}

func (s *Savior) SetOnReadHandler(f func(c net.Connection, data []byte)) {
	s.Handler.OnRead = f
}

func (s *Savior) SetOnIdleHandler(f func(c net.Connection)) {
	s.Handler.OnIdle = f
}
func (s *Savior) ServerStart() {
	if enableDebugLog {
		log.NewSaviorLogger()
	}
	switch s.Proto {
	case TCP:
		server := new(net.TCPServer)
		server.Port = s.Port
		server.Handler = s.Handler
		server.IdleMonitor = s.IdleMonitor
		server.Start()
	case KCP:
		server := new(net.KCPServer)
		server.Port = s.Port
		server.Handler = s.Handler
		server.IdleMonitor = s.IdleMonitor
		server.Start()
	case WS:
		server := new(net.WSServer)
		server.Port = s.Port
		server.Handler = s.Handler
		server.IdleMonitor = s.IdleMonitor
		if s.WebSocketMessageType > 0 {
			server.MessageType = s.WebSocketMessageType
		} else {
			server.MessageType = websocket.TextMessage
		}
		server.Start()
	default:
		server := new(net.TCPServer)
		server.Port = s.Port
		server.Handler = s.Handler
		server.IdleMonitor = s.IdleMonitor
		server.Start()
	}
}
