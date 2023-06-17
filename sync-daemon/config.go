package main

import (
	"fmt"
	"github.com/creasty/defaults"
	"gopkg.in/dealancer/validate.v2"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

func ReadConfig(filename string, out *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

type Config struct {
	Sources []Source `validate:"empty=false" yaml:"sources"`
	Rules   []Rule   `validate:"empty=false" yaml:"rules"`
}

func (config *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(config); err != nil {
		return err
	}
	type T Config // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(config)); err != nil {
		return err
	}
	if err := validate.Validate(config); err != nil {
		return err
	}
	return nil
}

func (config *Config) Validate() error {
	sourcesIdsUnique, nonUniqueSourceId := checkUnique(config.Sources, func(item Source) string {
		return item.ID
	})
	if !sourcesIdsUnique {
		return fmt.Errorf("invalid sources: non-unique id='%v'", *nonUniqueSourceId)
	}

	rulesIdsUnique, nonUniqueRuleId := checkUnique(config.Rules, func(item Rule) string {
		return item.ID
	})
	if !rulesIdsUnique {
		return fmt.Errorf("invalid rules: non-unique id='%v'", *nonUniqueRuleId)
	}
	return nil
}

type Source struct {
	ID           string        `validate:"empty=false"  yaml:"id"`
	Kind         string        `validate:"empty=false" yaml:"kind"`
	PollInterval time.Duration `default:"30s" yaml:"poll_interval"`

	S3Config struct {
		BucketName string `default:"30s" validate:"empty=false" yaml:"bucket_name"`
	} `yaml:"s3_config"`
}

func (s *Source) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T Source // declare new type to prevent stackoverflow
	t := (*T)(s)
	if err := unmarshal(t); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type CommandEntry struct {
	Command   []string `validate:"empty=false" yaml:"command"`
	OnFailure string   `default:"fail_sync" validate:"one_of=ignore,fail_sync,panic" yaml:"on_failure"` // one_of
}

func (s *CommandEntry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T CommandEntry // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type Rule struct {
	ID          string         `validate:"empty=false" yaml:"id"`
	Source      string         `validate:"empty=false" yaml:"source"`
	Prefix      string         `default:"" yaml:"prefix"`
	MaxFailures int            `default:"-1" yaml:"max_failures"`
	AfterSync   []CommandEntry `yaml:"after_sync"`
	Includes    []RuleIncludes `validate:"empty=false" yaml:"includes"`
}

func (s *Rule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T Rule // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type RuleIncludes struct {
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Cleanup bool   `default:"false" yaml:"cleanup"`
}

func (s *RuleIncludes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T RuleIncludes // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}
