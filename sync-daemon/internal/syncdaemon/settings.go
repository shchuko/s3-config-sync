package syncdaemon

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"time"
)

type syncDaemonSettings struct {
	sources []syncSourceSettings
}

type syncSourceSettings struct {
	source       SyncConfigSource
	id           string
	pollInterval time.Duration
	rules        []syncRule
}

func (s *syncSourceSettings) doSync() error {
	for _, rule := range s.rules {
		if err := rule.evaluate(s); err != nil {
			return err
		}
	}

	return nil
}

type syncRule struct {
	id          string
	prefix      string
	maxFailures int
	afterSync   []syncRuleCommand
	includes    []syncRuleInclude
}

func (s *syncRule) evaluate(source *syncSourceSettings) error {
	var err error = nil
	i := 0
	for {
		log.Info().
			Str("source", source.id).
			Str("rule", s.id).
			Int("attempt", i+1).
			Int("maxAttempts", s.maxFailures+1).
			Msg("Rule evaluation started")

		err = s.evaluateAttempt(source)
		if err == nil {
			log.Info().
				Str("source", source.id).
				Str("rule", s.id).
				Int("attempt", i+1).
				Int("maxAttempts", s.maxFailures+1).
				Msg("Rule evaluation complete")
			return nil
		}

		if i == s.maxFailures {
			log.Err(err).
				Str("source", source.id).
				Str("rule", s.id).
				Int("attempt", i+1).
				Int("maxAttempts", s.maxFailures+1).
				Msg("Rule evaluation failed, exiting")
			return err
		}

		sleepDuration := 5 * time.Second
		log.Err(err).
			Str("source", source.id).
			Str("rule", s.id).
			Int("attempt", i+1).
			Int("maxAttempts", s.maxFailures+1).
			Msg(fmt.Sprintf("Rule evaluation failed, will retry in %v", sleepDuration))
		time.Sleep(sleepDuration) // wait 5 seconds for recover
		i++
	}
}

func (s *syncRule) evaluateAttempt(source *syncSourceSettings) error {
	err := source.source.IterateFiles(s.prefix, func(path string) error {
		for _, include := range s.includes {
			if err := include.doSync(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, command := range s.afterSync {
		if err := command.invoke(source.id, s.id); err != nil {
			return err
		}
	}
	return nil
}

type syncRuleCommand struct {
	command   []string
	onFailure errorCardinality
}

type errorCardinality int8

var errorCardinalityMap = map[string]errorCardinality{
	"fail_sync": errorCardinality_failSync,
	"ignore":    errorCardinality_ignore,
}

//goland:noinspection GoSnakeCaseUsage
const (
	errorCardinality_failSync = iota
	errorCardinality_ignore   = iota
)

func (s *syncRuleCommand) invoke(sourceId string, ruleId string) error {
	log.Info().
		Str("source", sourceId).
		Str("rule", ruleId).
		Str("command", fmt.Sprintf("%v", s.command)).
		Msg("External process started")

	stdout, stderr, err := runExternalProcess(s.command)
	if err != nil {
		logger := log.Err(err).
			Str("source", sourceId).
			Str("rule", ruleId).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Str("command", fmt.Sprintf("%v", s.command))

		switch s.onFailure {
		case errorCardinality_ignore:
			logger.Str("onError", "ignore").Msg("External process failed")
			return nil
		case errorCardinality_failSync:
			logger.Str("onError", "fail_sync").Msg("External process failed")
			return err
		}
	}

	log.Info().
		Str("source", sourceId).
		Str("rule", ruleId).
		Str("command", fmt.Sprintf("%v", s.command)).
		Str("stdout", stdout).
		Str("stderr", stderr).
		Msg("External process finished")
	return nil
}

type syncRuleInclude struct {
	rawFrom     string
	destination string
	cleanup     bool
	matcher     regexp.Regexp
}

func newSyncRuleInclude(from string, to string, cleanup bool) *syncRuleInclude {
	return &syncRuleInclude{
		cleanup:     cleanup,
		rawFrom:     from,
		matcher:     regexp.Regexp{}, // TODO build regexp
		destination: to,
	}
}

func (s *syncRuleInclude) doSync(source string) error {
	// TODO implement sync
	return nil
}

func runExternalProcess(args []string) (string, string, error) {
	var cmd *exec.Cmd
	if len(args) > 1 {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0])
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := stdout.String()
	errStr := stderr.String()
	return outStr, errStr, err
}
