package utils

import (
	"context"
	"sync"
	"sync/atomic"
)

var (
	InstanceRoutinePool *RoutinePool
)

const (
	DefaultPoolSize = 16
)

// 任務
type TaskMethod func(params []interface{}) interface{}

///任務参数
type TaskParam struct {
	TaskMethod TaskMethod
	TaskParam  []interface{}
}

///协程上下文
type Worker struct {
	sync.RWMutex
	idleFlag int32 //是否閒置
	ctx      context.Context
	cancel   context.CancelFunc
	taskCh   chan *TaskParam
	wg       *sync.WaitGroup
	parentWG *sync.WaitGroup
}

///协程方法
func (object *Worker) run() {
	object.wg.Add(1)
loop:
	for {
		select {
		case <-object.ctx.Done():
			break loop
		case taskParam, ok := <-object.taskCh:
			if !ok {
				continue
			}

			// cancel worker
			if object.ctx.Err() != nil {
				break loop
			}

			object.parentWG.Add(1)
			taskParam.TaskMethod(taskParam.TaskParam)
			object.parentWG.Done()
			atomic.StoreInt32(&object.idleFlag, 1)
		}
	}
	object.wg.Done()
}

// worker pool
type RoutinePool struct {
	sync.RWMutex
	minRoutine int64
	workerPool []*Worker
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
}

func NewRoutinePool(minRouting int64) *RoutinePool {
	object := &RoutinePool{
		minRoutine: minRouting,
		wg:         &sync.WaitGroup{},
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	object.workerPool = make([]*Worker, minRouting)

	for i := int64(0); i < minRouting; i++ {
		worker := &Worker{
			idleFlag: 1,
			taskCh:   make(chan *TaskParam, 16),
			parentWG: object.wg,
			wg:       &sync.WaitGroup{},
		}
		worker.ctx, worker.cancel = context.WithCancel(context.Background())
		object.workerPool[i] = worker
		go worker.run()
	}

	return object
}

///回收worker
func (object *RoutinePool) recycleWorker(worker *Worker) {
	worker.cancel()
	worker.wg.Wait()

	worker.Lock()
	close(worker.taskCh)
	worker.taskCh = nil
	worker.Unlock()
}

///提交任务
func (object *RoutinePool) PostTask(task TaskMethod, params ...interface{}) {
	object.Lock()
	defer object.Unlock()

	i := 0
	for i < len(object.workerPool) {
		worker := object.workerPool[i]

		if atomic.LoadInt32(&(worker.idleFlag)) == 0 {
			i++
			continue
		}

		atomic.StoreInt32(&(worker.idleFlag), 0)

		worker.RLock()
		if worker.taskCh != nil && cap(worker.taskCh) > 0 {
			worker.taskCh <- &TaskParam{TaskMethod: task, TaskParam: params}
		}
		worker.RUnlock()
		break
	}
}

func (object *RoutinePool) Close() {
	object.cancel()
	object.wg.Wait()
	object.Lock()
	defer object.Unlock()
	for _, worker := range object.workerPool {
		object.recycleWorker(worker)
	}
	object.workerPool = nil
}

///初始化
func InitWorkerPool() {
	InstanceRoutinePool = NewRoutinePool(DefaultPoolSize)
}
