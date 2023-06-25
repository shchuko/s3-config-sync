package syncdaemon

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/scheduler"
	"reflect"
	"time"
)

const selfConfigReloadInterval = 30 * time.Second

type SyncConfigSource interface {
	IterateFiles(prefix string, processor func(path string) error) error
}

type SyncDaemon struct {
	configFilePath     string
	config             syncDaemonSettings
	autoReloadConfig   bool
	syncTasksScheduler *scheduler.Scheduler
	logger             zerolog.Logger
}

func NewSyncDaemon(configFilePath string, autoReloadConfig bool) *SyncDaemon {
	return &SyncDaemon{
		configFilePath:     configFilePath,
		autoReloadConfig:   autoReloadConfig,
		syncTasksScheduler: scheduler.NewScheduler(),
		logger:             log.With().Str("module", "syncdaemon").Logger(),
	}
}

func (s *SyncDaemon) Run() error {
	if s.syncTasksScheduler.IsRunning() {
		return errors.New("SyncDaemon is already running")
	}

	if err := s.runInternal(); err != nil {
		return err
	}

	return s.syncTasksScheduler.Wait()
}

func (s *SyncDaemon) runInternal() error {
	var config syncDaemonSettings
	if err := readConfig(s.configFilePath, &config); err != nil {
		return err
	}

	logger := s.logger.With().Str("action", "runInternal").Logger()
	if reflect.DeepEqual(config, s.config) {
		logger.Info().Msg("sync-daemon config is not changed")
		return nil
	}
	logger.Info().Msg("sync-daemon config loaded")

	s.config = config
	var tasks []scheduler.SyncSchedulerTask
	if s.autoReloadConfig {
		tasks = append(tasks, scheduler.SyncSchedulerTask{
			Runnable:   func() error { return s.runInternal() },
			Rate:       selfConfigReloadInterval,
			StartDelay: selfConfigReloadInterval})
	}

	sourcesTasks := collections.Map(s.config.sources, func(src syncSourceSettings) scheduler.SyncSchedulerTask {
		return scheduler.SyncSchedulerTask{
			Runnable: func() error {
				if err := src.doSync(); err != nil {
					return err
				}
				return nil
			},
			Rate:       src.pollInterval,
			StartDelay: 0,
		}
	})
	tasks = append(tasks, sourcesTasks...)
	s.syncTasksScheduler.Schedule(tasks)
	return nil
}
