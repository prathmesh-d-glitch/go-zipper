package worker

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
	tasks   chan Task
	results chan Result
	wg      sync.WaitGroup
	once    sync.Once
}

func NewPool(newWorkers, queueSize int) *Pool {
	if newWorkers < 1 {
		newWorkers = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}

	p := &Pool{
		tasks:   make(chan Task, queueSize),
		results: make(chan Result, queueSize),
	}

	for i := 0; i < newWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for task := range p.tasks {
		result := p.process(id, task)
		p.results <- result
	}
}

func (p *Pool) process(workerID int, task Task) (result Result) {
	result.TaskID = task.ID
	result.WorkerID = workerID

	defer func() {
		if r := recover(); r != nil {
			result.Err = fmt.Errorf("worker %d: panic: %v", workerID, r)
		}
	}()

	start := time.Now()

	switch task.Type {
	case TaskCompress:

	case TaskDecompress:

	default:
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return
}
