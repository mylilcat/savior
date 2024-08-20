package timer

import (
	"container/list"
	"github.com/mylilcat/savior/util"
	"log"
	"runtime"
	"time"
)

const (
	DelayTask int = iota
	IntervalTask
)

type task struct {
	f         func()
	round     int
	pos       int
	typ       int
	delayTime int64
}

type Timer struct {
	period   int64
	unit     time.Duration
	ticker   *time.Ticker
	stopChan chan any
	taskChan chan *task
	slots    []*list.List
	curSlot  int
}

// NewTimer Initialize a time wheel timer. 初始化时间轮定时器
// time unit,supports down to milliseconds. 时间间隔单位，最小到毫秒。
// slot,number of slots in the time wheel. 时间轮的槽位数量
// period,attention pls,if you set unit milliseconds,the minimum time interval is 15 milliseconds!!!!!!!!!!!!!!
// period，是定时器时间间隔，如果你设置的时间单位是毫秒，那period最小支持到15毫秒。
func NewTimer(period int64, unit time.Duration, slotNum int) *Timer {

	if !util.IsTimeUnitValid(unit) {
		panic("timer unit is invalid")
	}

	if period == 0 {
		period = 1
	}

	if unit == time.Millisecond && period < 15 {
		period = 15
	}

	if slotNum <= 0 {
		slotNum = 60
	}
	t := &Timer{
		period:   period,
		unit:     unit,
		ticker:   time.NewTicker(time.Duration(period) * unit),
		slots:    make([]*list.List, slotNum),
		taskChan: make(chan *task),
		stopChan: make(chan any),
	}
	for i := range t.slots {
		t.slots[i] = list.New()
	}
	return t
}

func (t *Timer) AddTask(f func(), delayTime int64, typ ...any) {
	if delayTime <= 0 {
		delayTime = 1
	}
	round := int(delayTime / (int64(len(t.slots)) * t.period))
	pos := int((int64(t.curSlot) + delayTime/t.period) % int64(len(t.slots)))
	tsk := &task{
		pos:   pos,
		round: round,
		f:     f,
	}

	if len(typ) == 0 || typ[0] == DelayTask {
		tsk.typ = DelayTask
	}

	if len(typ) > 0 && typ[0] == IntervalTask {
		tsk.typ = IntervalTask
		tsk.delayTime = delayTime
	}

	t.taskChan <- tsk
}

func (t *Timer) Start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, buf[:n])
			}
		}()
		for {
			select {
			case <-t.stopChan:
				return
			case tsk := <-t.taskChan:
				t.slots[tsk.pos].PushBack(tsk)
			case <-t.ticker.C:
				t.tick()
			}
		}
	}()
}

func (t *Timer) Stop() {
	t.ticker.Stop()
	close(t.stopChan)
}

func (t *Timer) tick() {
	t.curSlot = (t.curSlot + 1) % len(t.slots)
	list := t.slots[t.curSlot]
	for e := list.Front(); e != nil; {
		tsk := e.Value.(*task)
		if tsk.round > 0 {
			tsk.round--
			e = e.Next()
			continue
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 1024)
					n := runtime.Stack(buf, false)
					log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, buf[:n])
				}
			}()
			if tsk.typ == IntervalTask {
				defer t.AddTask(tsk.f, tsk.delayTime, tsk.typ)
			}
			tsk.f()
		}()
		next := e.Next()
		list.Remove(e)
		e = next
	}
}
