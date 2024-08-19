package fsm

import (
	"github.com/mylilcat/savior/util"
	"log"
	"runtime"
	"time"
)

type State struct {
	Name           string
	EnterTimestamp int64
	onEnter        func()
	onExecute      func()
	onExit         func()
	transitions    []*Transition
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

func (s *State) AddTransitions(t ...*Transition) {
	for _, transition := range t {
		s.transitions = append(s.transitions, transition)
	}
}

func NewState(name string) *State {
	return &State{
		Name:        name,
		transitions: []*Transition{},
	}
}

type Transition struct {
	nextState        *State
	isCanBeConverted func() bool
	onTransition     func()
}

func (t *Transition) SetConvertCondition(f func() bool) {
	t.isCanBeConverted = f
}

func NewTransition(nextState *State) *Transition {
	return &Transition{
		nextState: nextState,
	}
}

type FiniteStateMachine struct {
	states       []*State
	currentState *State
	stopChan     chan any
	period       int64
	unit         time.Duration
	ticker       *time.Ticker
	running      bool
}

func (f *FiniteStateMachine) CurrentState() *State {
	return f.currentState
}

func (f *FiniteStateMachine) IsRunning() bool {
	return f.running
}

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

	if !util.IsTimeUnitValid(f.unit) {
		f.unit = time.Millisecond
	}

	if f.unit == time.Millisecond && f.period < 15 {
		f.period = 15
	}

	f.ticker = time.NewTicker(time.Duration(f.period) * f.unit)

	if f.currentState.onEnter != nil {
		f.currentState.EnterTimestamp = time.Now().UnixMilli()
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

func NewFiniteStateMachine(initialState *State, s ...*State) *FiniteStateMachine {
	return &FiniteStateMachine{
		states:       s,
		running:      false,
		currentState: initialState,
	}
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
