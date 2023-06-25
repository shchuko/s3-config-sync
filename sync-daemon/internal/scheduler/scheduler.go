package scheduler

import (
	"errors"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"sync"
	"sync/atomic"
	"time"
)

type SyncSchedulerTask struct {
	Runnable   func() error
	Rate       time.Duration
	StartDelay time.Duration
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

func (s *Scheduler) Schedule(tasks []SyncSchedulerTask) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.stopFlag != nil {
		if !s.stopFlag.CompareAndSwap(false, true) { // if stopped by some other goroutine
			if s.error != nil {
				s.error = errors.Join(errors.New("previously scheduled tasks are failed with error"), s.error)
			}
			s.error = errors.New("scheduler is stopped, unable to schedule tasks")
		}
	}

	s.stopFlag = &atomic.Bool{}
	s.wg.Add(1)
	go s.scheduleTasks(tasks, s.stopFlag, &s.wg)
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
		return &internalSyncSchedulerTask{t, time.Now().Add(-t.Rate + t.StartDelay)}
	})
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if stop.Load() {
				return
			}

			for _, task := range internalTasks {
				now := time.Now()
				if task.lastStartTime.Add(task.Rate).Before(now) {
					task.lastStartTime = now
					if err := task.Runnable(); err != nil {
						s.mutex.Lock()
						s.stopFlag.Store(true)
						s.error = err
						s.mutex.Unlock()
						return
					}
				}
			}
		}
	}
}
