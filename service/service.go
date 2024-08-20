package service

import "sync"

type Service struct {
	name      string
	manager   *Manager
	wgManager sync.WaitGroup
}

// NewService make a service with actor,service name must be unique.
// 初始化一个actor服务类，服务名称要保证唯一。
func NewService(name string) *Service {
	if name == "" {
		panic("service name empty")
	}
	s := new(Service)
	s.name = name
	s.manager = NewManager()
	return s
}

// RegisterActorFunction register a service actor call function.
// 注册服务actor调用函数
func (s *Service) RegisterActorFunction(name string, function any) {
	s.manager.actor.RegisterFunction(name, function)
}

// SetInitFunc set service initialization function; for example,
// if this is a leaderboard service, you can load leaderboard data during the startup phase.
// 设置服务初始化函数，比如这是一个排行榜服务，就可以在启动阶段加载排行榜数据。
func (s *Service) SetInitFunc(fun func()) {
	s.manager.serviceInitFunc = fun
}

// SetDestroyFunc set a service shutdown handler; for example, save the in-memory leaderboard data after stopping the service.
// 设置服务关闭后处理函数，比如停服后保存内存中的排行榜数据。
func (s *Service) SetDestroyFunc(fun func()) {
	s.manager.serviceDestroyFunc = fun
}

func makeTask(funcName string, needResult bool, args []any) *taskInfo {
	task := new(taskInfo)
	task.functionName = funcName
	task.args = args
	if needResult {
		task.resultChan = make(chan []any)
	}
	return task
}

type taskInfo struct {
	functionName string
	args         []any
	resultChan   chan []any
}

// Call call other service actor methods with no return value.
// 调用其他服务actor方法，无返回值。
func Call(serviceName string, funcName string, args ...any) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, false, args)
	targetActor.send(task)
}

// Call1 call other service actor methods with 1 return value.
// 调用其他服务actor方法，有一个返回值。
func Call1[R any](serviceName string, funcName string, args ...any) (r R) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.send(task)
	result := <-task.resultChan
	r = result[0].(R)
	return
}

// Call2 call other service actor methods with 2 return value.
// 调用其他服务actor方法，有两个返回值。
func Call2[R1 any, R2 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.send(task)
	results := <-task.resultChan
	r1, r2 = results[0].(R1), results[1].(R2)
	return
}

func Call3[R1 any, R2 any, R3 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.send(task)
	results := <-task.resultChan
	r1, r2, r3 = results[0].(R1), results[1].(R2), results[2].(R3)
	return
}

func Call4[R1 any, R2 any, R3 any, R4 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3, r4 R4) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.send(task)
	results := <-task.resultChan
	r1, r2, r3, r4 = results[0].(R1), results[1].(R2), results[2].(R3), results[3].(R4)
	return
}

func Call5[R1 any, R2 any, R3 any, R4 any, R5 any](serviceName string, funcName string, args ...any) (r1 R1, r2 R2, r3 R3, r4 R4, r5 R5) {
	targetActor := getServiceActor(serviceName)
	if targetActor == nil {
		return
	}
	task := makeTask(funcName, true, args)
	targetActor.send(task)
	results := <-task.resultChan
	r1, r2, r3, r4, r5 = results[0].(R1), results[1].(R2), results[2].(R3), results[3].(R4), results[4].(R5)
	return
}
