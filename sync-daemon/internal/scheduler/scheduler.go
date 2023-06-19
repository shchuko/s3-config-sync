package scheduler

import (
	"errors"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"sync"
	"sync/atomic"
	"time"
)

type SyncSchedulerTask struct {
	Runnable func() error
	Rate     time.Duration
}

type Scheduler struct {
	mutex    sync.Mutex
	stopFlag *atomic.Bool
	wg       sync.WaitGroup
	error    error
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		mutex:    sync.Mutex{},
		stopFlag: nil,
		wg:       sync.WaitGroup{},
		error:    nil,
	}
}

func (s *Scheduler) Schedule(tasks []SyncSchedulerTask) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.stopFlag != nil {
		stoppedBefore := s.stopFlag.CompareAndSwap(false, true)
		s.wg.Wait() // wait for all tasks to finish

		if stoppedBefore {
			if s.error != nil {
				return errors.Join(errors.New("previously scheduled tasks are failed with error"), s.error)
			}
			return errors.New("scheduler is stopped, unable to schedule tasks")
		}
	}

	s.stopFlag = &atomic.Bool{}
	s.wg.Add(1)
	go s.scheduleTasks(tasks, s.stopFlag, &s.wg)

	return nil
}

func (s *Scheduler) IsRunning() bool {
	return s.stopFlag != nil && !s.stopFlag.Load()
}

func (s *Scheduler) Wait() error {
	s.wg.Wait()
	return s.error
}

func (s *Scheduler) Stop() {
	s.stopFlag.Store(true)
}

func (s *Scheduler) StopAndWait() error {
	s.Stop()
	return s.Wait()
}

type internalSyncSchedulerTask struct {
	SyncSchedulerTask
	lastStartTime time.Time
}

func (s *Scheduler) scheduleTasks(tasks []SyncSchedulerTask, stop *atomic.Bool, wg *sync.WaitGroup) {
	defer wg.Done()
	internalTasks := collections.Map(tasks, func(t SyncSchedulerTask) *internalSyncSchedulerTask {
		return &internalSyncSchedulerTask{t, time.UnixMilli(0)}
	})
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if stop.Load() {
				s.error = nil
				return
			}

			for _, task := range internalTasks {
				now := time.Now()
				if task.lastStartTime.Add(task.Rate).Before(now) {
					task.lastStartTime = now
					if err := task.Runnable(); err != nil {
						s.stopFlag.Store(true)
						s.error = err
						return
					}
				}
			}
		}
	}
}
