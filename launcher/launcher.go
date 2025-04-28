package launcher

import (
	"github.com/gorilla/websocket"
	"github.com/mylilcat/savior/log"
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/util"
	"time"
)

const (
	TCP = "tcp"
	KCP = "kcp"
	WS  = "ws"
)

var (
	//server port 服务端口
	port string

	//server prototype 服务器协议类型
	proto string

	//connection idle detection 连接空闲检测
	IMonitor *net.IdleMonitor

	webSocketMessageType int

	enableDebugLog bool
)

// SetProto set server proto. 设置服务协议
func SetProto(p string) {
	proto = p
}

// SetPort set server port. 设置服务端口
func SetPort(serverPort string) {
	port = serverPort
}

// SetIdleMonitor set idle timeout,and time unit. time unit,supports down to milliseconds.
// time.Microsecond, time.Millisecond, time.Second, time.Minute, time.Hour
// 设置连接空闲检测，读超时时间，写超时时间，超时时间单位。时间单位最小支持到毫秒。
func SetIdleMonitor(readIdle int64, writeIdle int64, unit time.Duration) {
	if readIdle < 0 || writeIdle < 0 || !util.IsTimeUnitValid(unit) {
		return
	}
	IMonitor = &net.IdleMonitor{
		ReadIdle:  readIdle,
		WriteIdle: writeIdle,
		Unit:      unit,
	}
}

func EnableDebugLog() {
	enableDebugLog = true
}

func SetWebSocketMessageType(t int) {
	webSocketMessageType = t
}

// ServerStart server start. 服务启动
func ServerStart() {
	if enableDebugLog {
		log.NewSaviorLogger()
	}
	switch proto {
	case TCP:
		s := new(net.TCPServer)
		s.Port = port
		port = ""
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		IMonitor = nil
		s.Start()
	case KCP:
		s := new(net.KCPServer)
		s.Port = port
		port = ""
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		IMonitor = nil
		s.Start()
	case WS:
		s := new(net.WSServer)
		s.Port = port
		port = ""
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		IMonitor = nil
		if webSocketMessageType > 0 {
			s.MessageType = webSocketMessageType
		} else {
			s.MessageType = websocket.TextMessage
		}
		s.Start()
	default:
		s := new(net.TCPServer)
		s.Port = port
		port = ""
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		IMonitor = nil
		s.Start()
	}
}
