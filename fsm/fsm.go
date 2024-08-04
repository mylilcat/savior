package fsm

import (
	"log"
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
}

func (f *FiniteStateMachine) CurrentState() *State {
	return f.currentState
}

func (f *FiniteStateMachine) SetPeriodAndUnit(period int64, unit time.Duration) {
	f.period = period
	f.unit = unit
}

func (f *FiniteStateMachine) Start() {
	f.stopChan = make(chan any)
	if f.period == 0 {
		f.period = int64(15)
	}

	if f.unit == 0 {
		f.unit = time.Millisecond
	}
	ticker := time.NewTicker(time.Duration(f.period) * f.unit)

	if f.currentState.onEnter != nil {
		f.currentState.EnterTimestamp = time.Now().UnixMilli()
		f.currentState.onEnter()
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
			}
		}()
		for {
			select {
			case <-ticker.C:
				f.update()
			case <-f.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (f *FiniteStateMachine) Stop() {
	close(f.stopChan)
}

func NewFiniteStateMachine(initialState *State, s ...*State) *FiniteStateMachine {
	return &FiniteStateMachine{
		states:       s,
		currentState: initialState,
	}
}

func (f *FiniteStateMachine) update() {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in update: %v", r)
		}
	}()

	for _, transition := range f.currentState.transitions {
		if transition.isCanBeConverted() {
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
