package pool

import "sync"

type Task func()

type WorkerPool struct {
	wg   sync.WaitGroup
	jobs chan Task
}

func New(size int) *WorkerPool {
	p := &WorkerPool{jobs: make(chan Task, size*4)}
	for i := 0; i < size; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for task := range p.jobs {
				task()
			}
		}()
	}
	return p
}

func (p *WorkerPool) Submit(task Task) {
	p.jobs <- task
}

func (p *WorkerPool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
