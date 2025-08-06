package service

import (
	saviorLog "github.com/mylilcat/savior/log"
	"sync"
)

var services = make(map[string]*Service)
var wg sync.WaitGroup
var lock sync.Mutex

func getServiceActor(name string) *actor {
	if service, ok := services[name]; ok {
		return service.manager.actor
	}
	return nil
}

func Register(service *Service) {
	if _, ok := services[service.name]; ok {
		panic("[SAVIOR] service name repeated")
	}
	services[service.name] = service
	saviorLog.Print("service register, name: %s", service.name)
}

func ServicesRun() {
	lock.Lock()
	defer lock.Unlock()
	for _, service := range services {
		if service.manager.running {
			continue
		}
		wg.Add(1)
		go service.manager.run()
	}
	for _, service := range services {
		if service.manager.initializer.initFunc != nil && !service.manager.initializer.executed {
			service.manager.initializer.initFunc()
			service.manager.initializer.executed = true
		}
	}
}

func ServicesStop() {
	lock.Lock()
	defer lock.Unlock()
	for _, service := range services {
		if !service.manager.running {
			continue
		}
		service.manager.stop()
		service.manager.wg.Wait()
	}
	for _, service := range services {
		if service.manager.finalizer.destroyFunc != nil && !service.manager.finalizer.executed {
			service.manager.finalizer.destroyFunc()
			service.manager.finalizer.executed = true
		}
	}
}
