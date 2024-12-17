package launcher

import (
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/util"
	"time"
)

const (
	TCP = "tcp"
	KCP = "kcp"
)

var (
	//server port 服务端口
	port string

	//server prototype 服务器协议类型
	proto string

	//connection idle detection 连接空闲检测
	IMonitor *net.IdleMonitor
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

// ServerStart server start. 服務启动
func ServerStart() {
	switch proto {
	case TCP:
		s := new(net.TCPServer)
		s.Port = port
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		s.Start()
	case KCP:
		s := new(net.KCPServer)
		s.Port = port
		s.Handler = net.NewHandler()
		s.IdleMonitor = IMonitor
		s.Start()
	default:
		s := new(net.TCPServer)
		s.Port = port
		s.Handler = net.NewHandler()
		s.Start()
	}
}
