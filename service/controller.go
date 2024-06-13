package service

import (
	"log"
	"sync"
)

var services = make(map[string]*Service)
var wg sync.WaitGroup

func GetServiceActor(name string) *Actor {
	if service, ok := services[name]; ok {
		return service.manager.actor
	}
	return nil
}

func Register(service *Service) {
	if _, ok := services[service.name]; ok {
		panic("service name repeated")
	}
	services[service.name] = service
	log.Print("service register,name:", service.name)
}

func ServicesRun() {
	for _, service := range services {
		wg.Add(1)
		log.Println("service name:", service.name)
		go service.manager.run()
	}
	for _, service := range services {
		if service.manager.serviceInitFunc != nil {
			service.manager.serviceInitFunc()
		}

	}
}

func ServicesStop() {
	for _, service := range services {
		service.manager.stop()
		service.manager.wg.Wait()
		if service.manager.serviceDestroyFunc != nil {
			service.manager.serviceDestroyFunc()
		}
	}
}
