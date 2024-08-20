package launcher

import (
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/timer"
	"github.com/mylilcat/savior/util"
	"sync"
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
	iMonitor *idleMonitor
)

type idleMonitor struct {
	//read idle timeout. 连接读超时
	readIdle int64

	//write idle timeout. 连接写超时
	writeIdle int64

	//time unit,supports down to milliseconds. 超时时间单位，最小支持到毫秒。
	//time.Microsecond, time.Millisecond, time.Second, time.Minute, time.Hour
	unit time.Duration
}

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
	iMonitor = &idleMonitor{
		readIdle:  readIdle,
		writeIdle: writeIdle,
		unit:      unit,
	}
	net.IdleMonitoring = idleMonitoring
}

// ServerStart server start. 服務启动
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
	default:
		s := new(net.TCPServer)
		s.Port = port
		s.Start()
	}
}

// connection idle checking. 连接空闲检测方法
func idleMonitoring(connections *sync.Map) {
	if iMonitor == nil {
		return
	}
	if !util.IsTimeUnitValid(iMonitor.unit) {
		return
	}
	var period int64
	if iMonitor.readIdle >= iMonitor.writeIdle {
		period = iMonitor.readIdle
	} else {
		period = iMonitor.writeIdle
	}

	idleTimer := timer.NewTimer(period, iMonitor.unit, 1)
	idleTimer.Start()
	idleTimer.AddTask(func() {
		connections.Range(func(key, value any) bool {
			conn := value.(net.Connection)
			if iMonitor.readIdle > 0 && conn.GetLastReadTime().Add(time.Duration(iMonitor.readIdle)*iMonitor.unit).Before(time.Now()) {
				net.OnIdle(conn)
				return true
			}
			if iMonitor.writeIdle > 0 && conn.GetLastWriteTime().Add(time.Duration(iMonitor.writeIdle)*iMonitor.unit).Before(time.Now()) {
				net.OnIdle(conn)
				return true
			}
			return true
		})
	}, period, timer.IntervalTask)
}
