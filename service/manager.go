package service

import "sync"

type Manager struct {
	actor       *actor
	wg          sync.WaitGroup
	initializer *initializer
	finalizer   *finalizer
	stopChan    chan any
	running     bool
}

type initializer struct {
	initFunc func()
	executed bool
}

type finalizer struct {
	destroyFunc func()
	executed    bool
}

func NewManager() *Manager {
	manager := new(Manager)
	manager.actor = NewActor()
	manager.stopChan = make(chan any, 1)
	manager.initializer = new(initializer)
	manager.finalizer = new(finalizer)
	manager.running = false
	return manager
}

func (m *Manager) init() {
	m.wg.Add(1)
	m.actor.run()
}

func (m *Manager) run() {
	m.init()
	m.running = true
	<-m.stopChan
	m.wg.Done()
}

func (m *Manager) stop() {
	m.actor.stop()
	m.actor.wgWorker.Wait()
	close(m.stopChan)
	m.running = false
}
