package service

import (
	"log"
	"reflect"
	"runtime"
	"sync"
)

type actor struct {
	pool           *routinePool
	actorFunctions map[string]*FunctionInfo
	wgWorker       sync.WaitGroup
	running        bool
}

func NewActor() *actor {
	a := new(actor)
	a.actorFunctions = make(map[string]*FunctionInfo)
	a.pool = new(routinePool)
	a.pool.workers = newWorkers()
	return a
}

type routineWorker struct {
	taskChan chan *taskInfo
	stopChan chan any
}

type routinePool struct {
	workers []*routineWorker
}

func (a *actor) poolStart() {
	for _, worker := range a.pool.workers {
		a.wgWorker.Add(1)
		worker.run(a)
	}
}

func newWorkers() []*routineWorker {
	var workers []*routineWorker
	for i := 0; i < 3; i++ {
		w := new(routineWorker)
		w.taskChan = make(chan *taskInfo, 100)
		w.stopChan = make(chan any, 1)
		workers = append(workers, w)
	}
	return workers
}

func (w *routineWorker) run(actor *actor) {
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
			case task := <-w.taskChan:
				executeTask(actor, task)
			case <-w.stopChan:
				if len(w.taskChan) > 0 {
					for task := range w.taskChan {
						executeTask(actor, task)
					}
				}
				actor.wgWorker.Done()
				return
			}
		}
	}()
}

func executeTask(actor *actor, task *taskInfo) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			log.Printf("Recovered from panic: %v\nStack trace:\n%s", r, buf[:n])
		}
	}()
	if functionInfo, ok := actor.actorFunctions[task.functionName]; ok {
		var argValues []reflect.Value
		if functionInfo.funcType.NumIn() > 0 {
			for _, arg := range task.args {
				argValues = append(argValues, reflect.ValueOf(arg))
			}
		}
		resultValues := functionInfo.funcValue.Call(argValues)
		if task.resultChan != nil {
			var results []any
			for _, value := range resultValues {
				results = append(results, value.Interface())
			}
			task.resultChan <- results
		}
	}
	return
}

type FunctionInfo struct {
	funcValue reflect.Value
	funcType  reflect.Type
}

func (a *actor) RegisterFunction(name string, function any) {

	if name == "" {
		panic("Function name empty")
	}
	funcValue := reflect.ValueOf(function)
	funcType := reflect.TypeOf(function)
	if funcValue.Kind() != reflect.Func {
		panic("Error registering actor function: the item to be registered is not a function type.")
	}
	functionInfo := new(FunctionInfo)
	functionInfo.funcValue = funcValue
	functionInfo.funcType = funcType
	a.actorFunctions[name] = functionInfo

}

func (a *actor) send(task *taskInfo) {
	a.pool.chooseWorker().submit(task)
}

func (w *routineWorker) submit(task *taskInfo) {
	w.taskChan <- task
}

func (p *routinePool) chooseWorker() *routineWorker {
	w := p.workers[0]
	for _, worker := range p.workers {
		if len(worker.taskChan) == 0 {
			return worker
		}
		if len(worker.taskChan) < len(w.taskChan) {
			w = worker
		}
	}
	return w
}

func (a *actor) run() {
	a.poolStart()
	a.running = true
}

func (a *actor) stop() {
	for _, worker := range a.pool.workers {
		close(worker.stopChan)
	}
	a.wgWorker.Wait()
	a.running = false
}
