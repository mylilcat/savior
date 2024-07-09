package launcher

import (
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/timer"
	"github.com/mylilcat/savior/util"
	"sync"
	"time"
)

const (
	TCP int = iota
	KCP
)

var (
	port     string
	proto    int
	iMonitor *idleMonitor
)

type idleMonitor struct {
	readIdle  int64
	writeIdle int64
	unit      time.Duration
}

func SetProto(p int) {
	proto = p
}

func SetPort(serverPort string) {
	port = serverPort
}

func SetIdleMonitor(readIdle int64, writeIdle int64, unit time.Duration) {
	if readIdle < 0 || writeIdle < 0 || !util.IsTimeUnitValid(unit) {
		return
	}
	iMonitor = &idleMonitor{
		readIdle:  readIdle,
		writeIdle: writeIdle,
		unit:      unit,
	}
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

func StartIdleMonitoring(connections *sync.Map) {
	if iMonitor == nil {
		return
	}
	if net.OnIdle == nil {
		return
	}
	if util.IsTimeUnitValid(iMonitor.unit) {
		return
	}
	var period int64
	if iMonitor.readIdle >= iMonitor.writeIdle {
		period = iMonitor.readIdle
	}
	idleTimer := timer.NewTimer(period, iMonitor.unit, 1)
	idleTimer.Start()
	idleTimer.AddTask(func() {
		connections.Range(func(key, value any) bool {
			conn := value.(net.Connection)
			if conn.GetLastReadTime().Add(time.Duration(iMonitor.readIdle) * iMonitor.unit).Before(time.Now()) {
				net.OnIdle(conn)
				return true
			}
			if conn.GetLastWriteTime().Add(time.Duration(iMonitor.writeIdle) * iMonitor.unit).Before(time.Now()) {
				net.OnIdle(conn)
				return true
			}
			return true
		})
	}, 1, timer.IntervalTask)
}
