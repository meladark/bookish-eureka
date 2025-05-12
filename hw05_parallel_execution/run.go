package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
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
	// Инициализируем каналы и воркеры
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskCh := make(chan Task)
	errCh := make(chan error, n)
	var wg sync.WaitGroup

	for i := 0; i < n || i < len(tasks); i++ {
		wg.Add(1)
		go worker(ctx, &wg, taskCh, errCh)
	}

	// Отправляем задачи в канал
	go func() {
		defer close(taskCh)
		// Заполняем канал задачами
		for _, task := range tasks {
			select {
			// Отправляем задачу в канал
			case taskCh <- task:
			// Отменяем выполнение задач при достижении лимита ошибок
			case <-ctx.Done():
				return
			}
		}
	}()

	// Следим за ошибками
	var errCount int

	for range tasks {
		select {
		case err := <-errCh:
			if err != nil {
				errCount++
				if errCount >= m {
					cancel()
					return ErrErrorsLimitExceeded
				}
			}
		// Сюда в целом не попасть, при компиляции не собирается?
		case <-ctx.Done():
		}
	}

	// Ждем завершения всех воркеров
	wg.Wait()

	// Проверяем, не было ли ошибок после отмены
	// if firstErr != nil {
	// 	return firstErr
	// }

	return nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan Task, errCh chan<- error) {
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
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
