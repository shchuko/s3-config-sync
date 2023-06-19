package syncdaemon

import (
	"errors"
	"fmt"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/scheduler"
	"time"
)

const selfConfigReloadInterval = 30 * time.Second

type SyncDaemon struct {
	configFilePath        string
	configReloadScheduler *scheduler.Scheduler
	syncTasksScheduler    *scheduler.Scheduler
}

func NewSyncDaemon(configFilePath string) *SyncDaemon {
	return &SyncDaemon{
		configFilePath:        configFilePath,
		configReloadScheduler: scheduler.NewScheduler(),
		syncTasksScheduler:    scheduler.NewScheduler(),
	}
}

func (s *SyncDaemon) Run() error {
	if s.configReloadScheduler.IsRunning() {
		return errors.New("SyncDaemon is already running")
	}
	err := s.configReloadScheduler.Schedule([]scheduler.SyncSchedulerTask{
		{Runnable: func() error { return s.configReload() }, Rate: selfConfigReloadInterval},
	})
	if err != nil {
		return err
	}

	return s.configReloadScheduler.Wait()
}

func (s *SyncDaemon) configReload() error {
	var config syncDaemonConfig

	if err := loadConfig(s.configFilePath, &config); err != nil {
		fmt.Println("Error loading syncConfig:", err)
		return err
	}
	fmt.Printf("Config reloaded: %#v\n", config)
	return nil
}
