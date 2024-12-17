package net

import (
	"github.com/mylilcat/savior/timer"
	"github.com/mylilcat/savior/util"
	"sync"
	"time"
)

type IdleMonitor struct {
	//read idle timeout. 连接读超时
	ReadIdle int64

	//write idle timeout. 连接写超时
	WriteIdle int64

	//time unit,supports down to milliseconds. 超时时间单位，最小支持到毫秒。
	//time.Microsecond, time.Millisecond, time.Second, time.Minute, time.Hour
	Unit time.Duration
}

// connection idle checking. 连接空闲检测方法
func (i *IdleMonitor) idleMonitoring(connections *sync.Map, onIdle func(connection Connection)) {
	if i == nil {
		return
	}
	if !util.IsTimeUnitValid(i.Unit) {
		return
	}
	var period int64
	if i.ReadIdle >= i.WriteIdle {
		period = i.ReadIdle
	} else {
		period = i.WriteIdle
	}

	idleTimer := timer.NewTimer(period, i.Unit, 1)
	idleTimer.Start()
	idleTimer.AddTask(func() {
		connections.Range(func(key, value any) bool {
			conn := value.(Connection)
			if i.ReadIdle > 0 && conn.GetLastReadTime().Add(time.Duration(i.ReadIdle)*i.Unit).Before(time.Now()) {
				if onIdle != nil {
					onIdle(conn)
				}
				return true
			}
			if i.WriteIdle > 0 && conn.GetLastWriteTime().Add(time.Duration(i.WriteIdle)*i.Unit).Before(time.Now()) {
				if onIdle != nil {
					onIdle(conn)
				}
				return true
			}
			return true
		})
	}, period, timer.IntervalTask)
}
