package worker

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/prathmesh-d-glitch/go-zipper/compressor/huffman"
	"github.com/prathmesh-d-glitch/go-zipper/compressor/lz77"
	"github.com/prathmesh-d-glitch/go-zipper/utils"
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

// goroutine body
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
		result = p.processCompress(workerID, task)
	case TaskDecompress:
		result = p.processDecompress(workerID, task)
	default:
		result.Err = fmt.Errorf("worker %d: unknown task type %q", workerID, task.Type)
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return
}

func (p *Pool) processCompress(workerID int, task Task) Result {
	res := Result{TaskID: task.ID, WorkerID: workerID}

	data, err := os.ReadFile(task.InputPath)
	if err != nil {
		res.Err = fmt.Errorf("reading %q: %w", task.InputPath, err)
	}

	res.BytesIn = int64(len(data))

	crc := utils.Checksum(data)

	tokens := lz77.Encode(data)
	serialised := lz77.SerializeTokens(tokens)
	compressed, err := huffman.Encode(serialised)
	if err != nil {
		res.Err = fmt.Errorf("comressing %q: %w", task.InputPath, err)
		return res
	}

	res.Output = compressed
	res.BytesOut = int64(len(compressed))

	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata["crc32"] = fmt.Sprintf("%d", crc)

	return res
}

func (p *Pool) processDecompress(workerID int, task Task) Result {
	res := Result{TaskID: task.ID, WorkerID: workerID}
	res.BytesIn = int64(len(task.InputData))

	serialised, err := huffman.Decode(task.InputData)
	if err != nil {
		res.Err = fmt.Errorf("huffman decode: %w", err)
		return res
	}

	tokens, err := lz77.DeserializeTokens(serialised)
	if err != nil {
		res.Err = fmt.Errorf("token deserialise: %w", err)
		return res
	}

	original, err := lz77.Decode(tokens)
	if err != nil {
		res.Err = fmt.Errorf("lz77 decode: %w", err)
		return res
	}

	res.Output = original
	res.BytesOut = int64(len(original))
	return res
}

// Enqueue a task
func (p *Pool) Submit(task Task) error {
	select {
	case p.tasks <- task:
		return nil
	default:
		return errors.New("worker: task queue is full")
	}
}

func (p *Pool) Results() <-chan Result {
	return p.results
}

func (p *Pool) Shutdown() {
	p.once.Do(func() {
		close(p.tasks)
		p.wg.Wait()
		close(p.results)
	})
}
