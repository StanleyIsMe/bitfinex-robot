package utils

import (
	"context"
	"sync"
)

const (
	PoolSize     = 8
	inputChannel = 100
	jobChannel   = 100
)

var WorkerPoolInstance *WorkerPool

// 方法
type TaskMethod func(params []interface{}) interface{}

// 参数
type TaskParam struct {
	TaskMethod TaskMethod
	TaskParam  []interface{}
}

type WorkerPool struct {
	inputChan chan *TaskParam
	jobsChan  chan *TaskParam

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func InitWorkerPool() {
	WorkerPoolInstance = NewWorkerPool()
}

func (c *WorkerPool) listen() {
	defer close(c.jobsChan)
	for {
		select {
		case job, ok := <-c.inputChan:
			if c.ctx.Err() != nil && !ok {
				return
			}
			c.jobsChan <- job
		}
	}
}

func (c *WorkerPool) worker(num int) {
	defer c.wg.Done()
	for {
		select {
		case job, ok := <-c.jobsChan:
			if c.ctx.Err() != nil && !ok {
				return
			}
			job.TaskMethod(job.TaskParam)
		case <-c.ctx.Done():
			if len(c.jobsChan) > 0 {
				continue
			}
			return
		}
	}
}

func NewWorkerPool() *WorkerPool {
	pool := &WorkerPool{
		inputChan: make(chan *TaskParam, inputChannel),
		jobsChan:  make(chan *TaskParam, jobChannel),
		wg:        &sync.WaitGroup{},
	}

	pool.ctx, pool.cancel = context.WithCancel(context.Background())

	for i := 0; i < PoolSize; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	go pool.listen()
	return pool
}

func (c *WorkerPool) AddTask(task TaskMethod, params ...interface{}) {
	c.inputChan <- &TaskParam{TaskMethod: task, TaskParam: params}
}
