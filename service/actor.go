package service

import (
	saviorLog "github.com/mylilcat/savior/log"
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
	for i := 0; i < 10; i++ {
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
				saviorLog.Print("routineWorker panicked: %v\nStack trace:\n%s", r, buf[:n])
			}
		}()
		for {
			select {
			case task := <-w.taskChan:
				executeTask(actor, task)
			case <-w.stopChan:
				for {
					select {
					case task := <-w.taskChan:
						executeTask(actor, task)
					default:
						actor.wgWorker.Done()
						return
					}
				}
			}
		}
	}()
}

func executeTask(actor *actor, task *taskInfo) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, false)
			saviorLog.Print("executeTask panicked: %v\nStack trace:\n%s", r, buf[:n], "function name:", task.functionName)
		}
	}()
	saviorLog.Print("actor start the task, task function: %v", task.functionName)
	if functionInfo, ok := actor.actorFunctions[task.functionName]; ok {
		var argValues []reflect.Value
		if functionInfo.funcType.NumIn() > 0 {
			for _, arg := range task.args {
				if arg == nil {
					argValues = append(argValues, reflect.Zero(functionInfo.funcType.In(len(argValues))))
				} else {
					argValues = append(argValues, reflect.ValueOf(arg))
				}
			}
		}
		resultValues := functionInfo.funcValue.Call(argValues)
		if task.resultChan != nil {
			defer func() {
				if r := recover(); r != nil {
					saviorLog.Print("send to resultChan panic: %v", r)
				}
			}()
			var results []any
			for _, value := range resultValues {
				results = append(results, value.Interface())
			}
			saviorLog.Print("actor task completed, task function: %v | return values: %v", task.functionName, len(results))
			if len(results) > 0 {
				task.resultChan <- results
			} else {
				close(task.resultChan)
			}
		} else {
			saviorLog.Print("actor async task completed, task function: %v", task.functionName)
		}
	} else {
		panic("[SAVIOR] actor function not found: " + task.functionName)
	}
	return
}

type FunctionInfo struct {
	funcValue reflect.Value
	funcType  reflect.Type
}

func (a *actor) RegisterFunction(name string, function any) {

	if name == "" {
		panic("[SAVIOR] function name empty")
	}
	funcValue := reflect.ValueOf(function)
	funcType := reflect.TypeOf(function)
	if funcValue.Kind() != reflect.Func {
		panic("[SAVIOR] error registering actor function: the item to be registered is not a function type.")
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
	select {
	case w.taskChan <- task:
	default:
		saviorLog.Print("taskChan full, task dropped,task function name: %v", task.functionName)
	}
}

func (p *routinePool) chooseWorker() *routineWorker {
	w := p.workers[0]
	for i, worker := range p.workers {
		if len(worker.taskChan) == 0 {
			saviorLog.Print("actor choose worker: %v", i)
			return worker
		}
		if len(worker.taskChan) < len(w.taskChan) {
			saviorLog.Print("actor current worker channel length: %v", w.taskChan)
			saviorLog.Print("actor switch worker %v | new worker channel length: %v", i, len(worker.taskChan))
			saviorLog.Print("actor choose worker: %v", i)
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
