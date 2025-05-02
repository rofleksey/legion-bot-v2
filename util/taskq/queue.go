package taskq

import (
	"context"
	"github.com/alitto/pond/v2"
	"golang.org/x/time/rate"
	"log/slog"
)

type Queue struct {
	pool    pond.ResultPool[any]
	limiter *rate.Limiter
}

type TaskWithResult struct {
	f    func() any
	resp chan any
}

func New(maxConcurrent int, rateLimit float64, burst int) *Queue {
	if maxConcurrent < 1 {
		panic("maxConcurrent must be at least 1")
	}
	if rateLimit <= 0 {
		panic("rateLimit must be positive")
	}
	if burst < 1 {
		panic("burst must be at least 1")
	}

	q := &Queue{
		pool:    pond.NewResultPool[any](maxConcurrent),
		limiter: rate.NewLimiter(rate.Limit(rateLimit), burst),
	}

	return q
}

func (q *Queue) Enqueue(f func()) {
	q.pool.Submit(func() any {
		if err := q.limiter.Wait(context.Background()); err != nil {
			return nil
		}

		f()
		return nil
	})
}

func Compute[T any](q *Queue, f func() T) T {
	task := q.pool.Submit(func() any {
		if err := q.limiter.Wait(context.Background()); err != nil {
			return nil
		}

		var anyRes any
		anyRes = f()
		return anyRes
	})

	result, err := task.Wait()
	if err != nil {
		slog.Error("Compute error",
			slog.Any("error", err),
		)

		var zero T
		return zero
	}

	if result == nil {
		var zero T
		return zero
	}

	return result.(T)
}

type ResWithErr struct {
	Res any
	Err error
}

func ComputeWithError[T any](q *Queue, f func() (T, error)) (T, error) {
	task := q.pool.Submit(func() any {
		if err := q.limiter.Wait(context.Background()); err != nil {
			return nil
		}

		res, err := f()

		return ResWithErr{res, err}
	})

	result, err := task.Wait()
	if err != nil {
		slog.Error("Compute error",
			slog.Any("error", err),
		)

		var zero T

		return zero, err
	}

	resWithErr := result.(ResWithErr)

	if resWithErr.Res == nil {
		var zero T

		return zero, resWithErr.Err
	}

	return resWithErr.Res.(T), resWithErr.Err
}

func (q *Queue) Shutdown() {
	q.pool.StopAndWait()
}
