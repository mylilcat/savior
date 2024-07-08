package timer

import (
	"container/list"
	"log"
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

func NewTimer(period int64, unit time.Duration, slotNum int) *Timer {
	if period <= 0 {
		period = 1
	}
	if unit <= 0 {
		unit = time.Second
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

	delayTime = delayTime * int64(t.unit)
	period := t.period * int64(t.unit)
	round := int(delayTime / (int64(len(t.slots)) * period))
	pos := int((int64(t.curSlot) + delayTime/period) % int64(len(t.slots)))
	tsk := &task{
		pos:   pos,
		round: round,
		f:     f,
	}
	switch typ[0] {
	case DelayTask:
		tsk.typ = DelayTask
	case IntervalTask:
		tsk.typ = IntervalTask
		tsk.delayTime = delayTime
	default:
		tsk.typ = DelayTask
	}
	t.taskChan <- tsk
}

func (t *Timer) Start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("timer run err:", r.(error))
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
	defer func() {
		t.curSlot = (t.curSlot + 1) % len(t.slots)
	}()
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
					log.Println("timer run err:", r.(error))
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
