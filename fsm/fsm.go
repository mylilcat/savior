package fsm

import (
	"github.com/mylilcat/savior/util"
	"log"
	"runtime"
	"time"
)

// State
type State struct {
	Name           string        // state name 状态名称
	EnterTimestamp int64         // enter this state timestamp,seconds or milliseconds. 进入到当前状态的时间戳，单位是秒，或者毫秒。
	onEnter        func()        // the method executed upon entering this state. 进入该状态执行的方法。
	onExecute      func()        // the method that continue to execute while in this state, such as enemy detection. 出入当前状态，要持续执行的方法，比如寻敌。
	onExit         func()        // the method to be executed when exiting this state. 退出该状态执行的方法。
	transitions    []*Transition // transition condition list. 准换条件列表。
}

// NewState create a new state. 新建一个状态
func NewState(name string) *State {
	return &State{
		Name:        name,
		transitions: []*Transition{},
	}
}

func (s *State) SetOnEnter(f func()) {
	s.onEnter = f
}

func (s *State) SetOnExist(f func()) {
	s.onExit = f
}

func (s *State) SetOnExecute(f func()) {
	s.onExecute = f
}

// AddTransitions add transition conditions. 添加状态转换条件
func (s *State) AddTransitions(t ...*Transition) {
	for _, transition := range t {
		s.transitions = append(s.transitions, transition)
	}
}

// NewTransition create new transition. 创建转换条件
// nextState the next state to transition to. 要转换的下一个状态
func NewTransition(nextState *State) *Transition {
	return &Transition{
		nextState: nextState,
	}
}

// Transition transition
type Transition struct {
	nextState        *State      // next state 符合转换条件后要进入到的下一个状态
	isCanBeConverted func() bool // checking transition conditions 检查是否符合转换条件
}

// SetConvertCondition set convert condition function 设置转换条件检查方法
func (t *Transition) SetConvertCondition(f func() bool) {
	t.isCanBeConverted = f
}

// FiniteStateMachine finite state machine
type FiniteStateMachine struct {
	currentState *State
	stopChan     chan any
	period       int64
	unit         time.Duration
	ticker       *time.Ticker
	running      bool
}

// NewFiniteStateMachine create new finite state machine. 创建一个状态机
// initialState 初始状态
func NewFiniteStateMachine(initialState *State) *FiniteStateMachine {
	return &FiniteStateMachine{
		running:      false,
		currentState: initialState,
	}
}

func (f *FiniteStateMachine) CurrentState() *State {
	return f.currentState
}

func (f *FiniteStateMachine) IsRunning() bool {
	return f.running
}

// SetPeriodAndUnit set the execution interval and time unit for the state machine; the minimum supported interval is 15 milliseconds.
// 设置状态机的执行间隔和时间单位，最小时间间隔为15毫秒。
// setting is optional; the default execution interval is 15 milliseconds.
// 可以不进行设置，默认为15毫秒间隔执行。
func (f *FiniteStateMachine) SetPeriodAndUnit(period int64, unit time.Duration) {
	if !util.IsTimeUnitValid(unit) {
		panic("fsm unit is invalid")
	}
	f.period = period
	f.unit = unit
}

func (f *FiniteStateMachine) Start() {

	if f.running {
		return
	}

	f.stopChan = make(chan any)

	if f.period == 0 {
		f.period = 1
	}

	if !util.IsTimeSecondOrTimeMillisecond(f.unit) {
		f.unit = time.Millisecond
	}

	if f.unit == time.Millisecond && f.period < 15 {
		f.period = 15
	}

	f.ticker = time.NewTicker(time.Duration(f.period) * f.unit)

	if f.currentState.onEnter != nil {
		f.currentState.EnterTimestamp = func() int64 {
			switch f.unit {
			case time.Second:
				return time.Now().Unix()
			default:
				return time.Now().UnixMilli()
			}
		}()
		f.currentState.onEnter()
	}
	f.running = true
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
			case <-f.ticker.C:
				f.update()
			case <-f.stopChan:
				f.running = false
				return
			}
		}
	}()
}

func (f *FiniteStateMachine) Stop() {
	if !f.running {
		return
	}
	f.ticker.Stop()
	close(f.stopChan)
}

func (f *FiniteStateMachine) update() {

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, buf[:n])
		}
	}()

	for _, transition := range f.currentState.transitions {
		if transition.isCanBeConverted != nil && transition.isCanBeConverted() {
			ns := transition.nextState
			if f.currentState.onExit != nil {
				f.currentState.onExit()
			}
			f.currentState = ns
			if f.currentState.onEnter != nil {
				f.currentState.onEnter()
				f.currentState.EnterTimestamp = time.Now().UnixMilli()
			}
			break
		}
	}
	if f.currentState.onExecute != nil {
		f.currentState.onExecute()
	}
}
