package service

import "sync"

type Service struct {
	name      string
	manager   *Manager
	wgManager sync.WaitGroup
}

func NewService(name string) *Service {
	if name == "" {
		panic("service name empty")
	}
	s := new(Service)
	s.name = name
	s.manager = NewManager()
	return s
}

func (s *Service) RegisterActorFunction(name string, function any) {
	s.manager.actor.RegisterFunction(name, function)
}

func (s *Service) SetInitFunc(fun func()) {
	s.manager.serviceInitFunc = fun
}

func (s *Service) SetDestroyFunc(fun func()) {
	s.manager.serviceDestroyFunc = fun
}

func makeTask(funcName string, needResult bool, args []any) *TaskInfo {
	task := new(TaskInfo)
	task.functionName = funcName
	task.args = args
	if needResult {
		task.resultChan = make(chan []any)
	}
	return task
}

type TaskInfo struct {
	functionName string
	args         []any
	resultChan   chan []any
}

func Call(serviceName string, funcName string, args ...any) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, false, args)
	targetActor.Send(task)
}

func Call1[R any](serviceName string, funcName string, args ...any) (r R) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.Send(task)
	result := <-task.resultChan
	r = result[0].(R)
	return
}

func Call2[R1 any, R2 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.Send(task)
	results := <-task.resultChan
	r1, r2 = results[0].(R1), results[1].(R2)
	return
}

func Call3[R1 any, R2 any, R3 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.Send(task)
	results := <-task.resultChan
	r1, r2, r3 = results[0].(R1), results[1].(R2), results[2].(R3)
	return
}

func Call4[R1 any, R2 any, R3 any, R4 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3, r4 R4) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.Send(task)
	results := <-task.resultChan
	r1, r2, r3, r4 = results[0].(R1), results[1].(R2), results[2].(R3), results[3].(R4)
	return
}

func Call5[R1 any, R2 any, R3 any, R4 any, R5 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3, r4 R4, r5 R5) {
	targetActor := GetServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.Send(task)
	results := <-task.resultChan
	r1, r2, r3, r4, r5 = results[0].(R1), results[1].(R2), results[2].(R3), results[3].(R4), results[4].(R5)
	return
}
