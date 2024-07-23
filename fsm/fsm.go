package fsm

import "time"

type State struct {
	Name        string
	onEnter     func()
	onExecute   func()
	onExit      func()
	transitions []*Transition
}

func (s *State) setOnEnter(f func()) {
	s.onEnter = f
}

func (s *State) setOnExist(f func()) {
	s.onExit = f
}

func (s *State) setOnExecute(f func()) {
	s.onExecute = f
}

func (s *State) addTransitions(t ...*Transition) {
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

func (t *Transition) setConvertCondition(f func() bool) {
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
}

func (f *FiniteStateMachine) CurrentState() *State {
	return f.currentState
}

func (f *FiniteStateMachine) Start(options ...any) {
	f.stopChan = make(chan any)
	period := int64(10)
	unit := time.Millisecond
	if len(options) > 0 {
		period = options[0].(int64)
	}
	if len(options) > 1 {
		unit = options[1].(time.Duration)
	}
	ticker := time.NewTicker(time.Duration(period) * unit)

	go func() {
		for {
			select {
			case <-ticker.C:
				f.update()
			case <-f.stopChan:
				ticker.Stop()
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
	for _, transition := range f.currentState.transitions {
		if transition.isCanBeConverted() {
			ns := transition.nextState
			f.currentState.onExit()
			f.currentState = ns
			f.currentState.onEnter()
			break
		}
	}
	f.currentState.onExecute()
}
