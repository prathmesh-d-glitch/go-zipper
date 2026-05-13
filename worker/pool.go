package worker

import "sync"

type Pool struct {
	tasks   chan Task
	results chan Result
	wg      sync.WaitGroup
	once    sync.Once
}
