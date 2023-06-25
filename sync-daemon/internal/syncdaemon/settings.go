package syncdaemon

import (
	"bytes"
	"errors"
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
			Int("attempt", i).
			Msg("Rule evaluation started")

		err = s.evaluateAttempt(source.source)
		if err == nil {
			log.Info().
				Str("source", source.id).
				Str("rule", s.id).
				Msg("Rule evaluation complete")
			return nil
		}

		if i == s.maxFailures {
			log.Err(err).
				Str("source", source.id).
				Str("rule", s.id).
				Msg(fmt.Sprintf("Rule evaluation failed, exiting"))
			return err
		}
		log.Err(err).
			Str("source", source.id).
			Str("rule", s.id).
			Int("attempt", i).
			Msg(fmt.Sprintf("Rule evaluation failed"))
		i++
	}
}

func (s *syncRule) evaluateAttempt(source SyncConfigSource) error {
	return source.IterateFiles(s.prefix, func(path string) error {
		for _, include := range s.includes {
			if err := include.doSync(path); err != nil {
				return err
			}

			for _, command := range s.afterSync {
				if err := command.invoke(); err != nil {
					return err
				}
			}
		}
		return nil
	})
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

func (s *syncRuleCommand) invoke() error {
	if err := runExternalProcess(s.command); err != nil {
		switch s.onFailure {
		case errorCardinality_ignore:
			return nil
		case errorCardinality_failSync:
			return err
		}
	}
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

func runExternalProcess(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := stdout.String()
	errStr := stderr.String()
	if err != nil {
		detailedError := fmt.Errorf("failed to execute command '%v': stdout='%s', stderr='%s'", args, outStr, errStr)
		return errors.Join(err, detailedError)
	}

	return nil
}
