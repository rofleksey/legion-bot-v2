package taskq

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/time/rate"
)

type Queue struct {
	taskChan     chan any
	workerSem    chan struct{}
	limiter      *rate.Limiter
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	ctx          context.Context
	cancel       context.CancelFunc
}

type TaskWithResult struct {
	f    func() any
	resp chan any
}

func New(maxConcurrent int, rateLimit float64, burst int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	q := &Queue{
		taskChan:  make(chan any, 64),
		ctx:       ctx,
		cancel:    cancel,
		workerSem: make(chan struct{}, maxConcurrent),
		limiter:   rate.NewLimiter(rate.Limit(rateLimit), burst),
	}

	q.wg.Add(maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		go q.worker()
	}

	return q
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case task, ok := <-q.taskChan:
			if !ok {
				return
			}

			select {
			case q.workerSem <- struct{}{}:
			case <-q.ctx.Done():
				return
			}

			func() {
				defer func() { <-q.workerSem }()

				if err := q.limiter.Wait(q.ctx); err != nil {
					if twr, ok := task.(TaskWithResult); ok {
						close(twr.resp)
					}
					return
				}

				defer func() {
					if err := recover(); err != nil {
						slog.Error("Error executing task", slog.Any("error", err))
					}
				}()

				switch t := task.(type) {
				case func():
					t()
				case TaskWithResult:
					t.resp <- t.f()
					close(t.resp)
				}
			}()
		}
	}
}

func (q *Queue) Enqueue(task func()) {
	if task == nil {
		panic("task cannot be nil")
	}

	select {
	case q.taskChan <- task:
	case <-q.ctx.Done():
	}
}

func Compute[T any](q *Queue, f func() T) T {
	if f == nil {
		panic("function cannot be nil")
	}

	wrapper := func() any {
		return f()
	}

	task := TaskWithResult{
		f:    wrapper,
		resp: make(chan any, 1),
	}

	select {
	case q.taskChan <- task:
		select {
		case res := <-task.resp:
			return res.(T)
		case <-q.ctx.Done():
			var zero T
			return zero
		}
	case <-q.ctx.Done():
		var zero T
		return zero
	}
}

func (q *Queue) Shutdown() {
	q.shutdownOnce.Do(func() {
		q.cancel()
		q.wg.Wait()
		close(q.taskChan)
	})
}
