package service

import (
	"log"
	"reflect"
	"sync"
)

type Actor struct {
	pool           *routinePool
	actorFunctions map[string]*FunctionInfo
	wgWorker       sync.WaitGroup
	running        bool
}

func NewActor() *Actor {
	actor := new(Actor)
	actor.actorFunctions = make(map[string]*FunctionInfo)
	actor.pool = new(routinePool)
	actor.pool.workers = newWorkers()
	return actor
}

type routineWorker struct {
	taskChan chan *TaskInfo
	stopChan chan any
}

type routinePool struct {
	workers []*routineWorker
}

func (a *Actor) poolStart() {
	for _, worker := range a.pool.workers {
		a.wgWorker.Add(1)
		worker.run(a)
	}
}

func newWorkers() []*routineWorker {
	var workers []*routineWorker
	for i := 0; i < 3; i++ {
		w := new(routineWorker)
		w.taskChan = make(chan *TaskInfo, 20)
		w.stopChan = make(chan any, 1)
		workers = append(workers, w)
	}
	return workers
}

func (w *routineWorker) run(actor *Actor) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Print(r)
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

func executeTask(actor *Actor, task *TaskInfo) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("executeTask panic:", r)
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

func (a *Actor) RegisterFunction(name string, function any) {

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

func (a *Actor) Send(task *TaskInfo) {
	a.pool.chooseWorker().Submit(task)
}

func (w *routineWorker) Submit(task *TaskInfo) {
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

func (a *Actor) run() {
	a.poolStart()
	a.running = true
}

func (a *Actor) stop() {
	for _, worker := range a.pool.workers {
		close(worker.stopChan)
	}
	a.wgWorker.Wait()
	a.running = false
}
