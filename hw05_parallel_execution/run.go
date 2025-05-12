package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n int, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}
	if len(tasks) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	taskCh := make(chan Task, n)
	errCh := make(chan error, m)
	errorTokens := make(chan struct{}, m)
	for i := 0; i < m; i++ {
		errorTokens <- struct{}{}
	}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(ctx, &wg, taskCh, errCh, errorTokens)
	}
	go func() {
		defer close(taskCh)
		for _, task := range tasks {
			select {
			case taskCh <- task:
			case <-ctx.Done():
				return
			}
		}
	}()
	var errCount int64
	for range tasks {
		select {
		case err := <-errCh:
			if err != nil {
				if atomic.AddInt64(&errCount, 1) >= int64(m) {
					cancel()
					wg.Wait()
					return ErrErrorsLimitExceeded
				}
			}
		case <-ctx.Done():
		}
	}
	return nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan Task, errCh chan<- error, errTokens chan struct{}) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-taskCh:
			if !ok {
				return
			}
			err := task()
			select {
			case errCh <- err:
				if err != nil {
					select {
					case <-errTokens:
					default:
						return
					}
				}
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
