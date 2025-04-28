package util

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type TaskQueue struct {
	taskChan     chan func()
	workerSem    chan struct{}
	limiter      *rate.Limiter
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewTaskQueue(maxConcurrent int, rateLimit float64, burst int) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	q := &TaskQueue{
		taskChan:  make(chan func(), 64),
		ctx:       ctx,
		cancel:    cancel,
		workerSem: make(chan struct{}, maxConcurrent),
		limiter:   rate.NewLimiter(rate.Limit(rateLimit), burst),
	}

	go q.dispatcher()

	return q
}

func (q *TaskQueue) dispatcher() {
	for task := range q.taskChan {
		select {
		case q.workerSem <- struct{}{}:
		case <-q.ctx.Done():
			return
		}

		q.wg.Add(1)
		go func(t func()) {
			defer q.wg.Done()
			defer func() { <-q.workerSem }()

			if err := q.limiter.Wait(q.ctx); err != nil {
				return
			}

			t()
		}(task)
	}
}

func (q *TaskQueue) Enqueue(task func()) {
	if task == nil {
		panic("task cannot be nil")
	}

	select {
	case q.taskChan <- task:
	case <-q.ctx.Done():
	}
}

func (q *TaskQueue) Shutdown() {
	q.shutdownOnce.Do(func() {
		q.cancel()
		q.wg.Wait()
		close(q.taskChan)
	})
}
