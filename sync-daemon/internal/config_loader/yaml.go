package config_parser

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"gopkg.in/dealancer/validate.v2"
	"time"
)

type SyncConfigYaml struct {
	Sources []SourceYaml `validate:"empty=false" yaml:"sources"`
	Rules   []RuleYaml   `validate:"empty=false" yaml:"rules"`
}

func (yamlConfig *SyncConfigYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(yamlConfig); err != nil {
		return err
	}
	type T SyncConfigYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(yamlConfig)); err != nil {
		return err
	}
	if err := validate.Validate(yamlConfig); err != nil {
		return err
	}
	return nil
}

func (yamlConfig *SyncConfigYaml) Validate() error {
	sourcesIdsUnique, nonUniqueSourceId := collections.CheckUnique(yamlConfig.Sources, func(item SourceYaml) string {
		return item.ID
	})
	if !sourcesIdsUnique {
		return fmt.Errorf("invalid sources: non-unique id='%v'", *nonUniqueSourceId)
	}

	rulesIdsUnique, nonUniqueRuleId := collections.CheckUnique(yamlConfig.Rules, func(item RuleYaml) string {
		return item.ID
	})
	if !rulesIdsUnique {
		return fmt.Errorf("invalid rules: non-unique id='%v'", *nonUniqueRuleId)
	}
	return nil
}

type SourceYaml struct {
	ID           string        `validate:"empty=false"  yaml:"id"`
	Kind         string        `validate:"empty=false" yaml:"kind"`
	PollInterval time.Duration `default:"30s" yaml:"poll_interval"`

	S3Config struct {
		BucketName string `default:"30s" validate:"empty=false" yaml:"bucket_name"`
	} `yaml:"s3_config"`
}

func (s *SourceYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T SourceYaml // declare new type to prevent stackoverflow
	t := (*T)(s)
	if err := unmarshal(t); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type CommandEntryYaml struct {
	Command   []string `validate:"empty=false" yaml:"command"`
	OnFailure string   `default:"fail_sync" validate:"one_of=ignore,fail_sync,panic" yaml:"on_failure"` // one_of
}

func (s *CommandEntryYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T CommandEntryYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type RuleYaml struct {
	ID          string             `validate:"empty=false" yaml:"id"`
	Source      string             `validate:"empty=false" yaml:"source"`
	Prefix      string             `default:"" yaml:"prefix"`
	MaxFailures int                `default:"-1" yaml:"max_failures"`
	AfterSync   []CommandEntryYaml `yaml:"after_sync"`
	Includes    []RuleIncludesYaml `validate:"empty=false" yaml:"includes"`
}

func (s *RuleYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T RuleYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type RuleIncludesYaml struct {
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Cleanup bool   `default:"false" yaml:"cleanup"`
}

func (s *RuleIncludesYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T RuleIncludesYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}
