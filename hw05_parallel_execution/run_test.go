package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)
		var runTasksCount int32
		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}
		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)
		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)
		var runTasksCount int32
		var sumTime time.Duration
		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}
		workersCount := 5
		maxErrorsCount := 1
		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)
		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
	t.Run("tasks without errors", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := 0; i < 10; i++ {
			tasks[i] = func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}
		}
		err := Run(tasks, 4, 2)
		require.NoError(t, err)
	})

	t.Run("tasks with errors exceeding limit", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return errors.New("error")
				}
			} else {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}
			}
		}

		err := Run(tasks, 4, 2)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
	})

	t.Run("tasks with errors not exceeding limit", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := 0; i < 10; i++ {
			if i < 1 {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return errors.New("error")
				}
			} else {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}
			}
		}

		err := Run(tasks, 4, 2)
		require.NoError(t, err)
	})

	t.Run("m <= 0", func(t *testing.T) {
		tasks := make([]Task, 10)
		for i := 0; i < 10; i++ {
			tasks[i] = func() error {
				return nil
			}
		}

		err := Run(tasks, 4, 0)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
	})

	t.Run("no tasks", func(t *testing.T) {
		err := Run([]Task{}, 4, 2)
		require.NoError(t, err)
	})

	t.Run("tasks less than workers", func(t *testing.T) {
		tasks := make([]Task, 2)
		for i := 0; i < 2; i++ {
			tasks[i] = func() error {
				return nil
			}
		}

		err := Run(tasks, 4, 2)
		require.NoError(t, err)
	})

	t.Run("tasks with errors exceeding limit and 1000000", func(t *testing.T) {
		tasks := make([]Task, 1000000)
		for i := 0; i < 100000; i++ {
			if i%2 == 0 {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return errors.New("error")
				}
			} else {
				tasks[i] = func() error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}
			}
		}

		err := Run(tasks, 4, 2)
		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
	})
}
