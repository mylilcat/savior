package service

import "sync"

type Manager struct {
	actor              *actor
	wg                 sync.WaitGroup
	serviceInitFunc    func()
	serviceDestroyFunc func()
	stopChan           chan any
}

func NewManager() *Manager {
	manager := new(Manager)
	manager.actor = NewActor()
	manager.stopChan = make(chan any, 1)
	return manager
}

func (m *Manager) init() {
	m.wg.Add(1)
	m.actor.run()
}

func (m *Manager) run() {
	m.init()
	<-m.stopChan
	m.wg.Done()
}

func (m *Manager) stop() {
	m.actor.stop()
	m.actor.wgWorker.Wait()
	close(m.stopChan)
}
