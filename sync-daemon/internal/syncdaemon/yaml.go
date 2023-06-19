package syncdaemon

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"gopkg.in/dealancer/validate.v2"
	"time"
)

type syncDaemonConfigYaml struct {
	Sources []syncDaemonConfigSourceYaml `validate:"empty=false" yaml:"sources"`
	Rules   []syncDaemonConfigRuleYaml   `validate:"empty=false" yaml:"rules"`
}

func (yamlConfig *syncDaemonConfigYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(yamlConfig); err != nil {
		return err
	}
	type T syncDaemonConfigYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(yamlConfig)); err != nil {
		return err
	}
	if err := validate.Validate(yamlConfig); err != nil {
		return err
	}
	return nil
}

func (yamlConfig *syncDaemonConfigYaml) Validate() error {
	sourcesIdsUnique, nonUniqueSourceId := collections.CheckUnique(yamlConfig.Sources, func(item syncDaemonConfigSourceYaml) string {
		return item.ID
	})
	if !sourcesIdsUnique {
		return fmt.Errorf("invalid sources: non-unique id='%v'", *nonUniqueSourceId)
	}

	rulesIdsUnique, nonUniqueRuleId := collections.CheckUnique(yamlConfig.Rules, func(item syncDaemonConfigRuleYaml) string {
		return item.ID
	})
	if !rulesIdsUnique {
		return fmt.Errorf("invalid rules: non-unique id='%v'", *nonUniqueRuleId)
	}
	return nil
}

type syncDaemonConfigSourceYaml struct {
	ID           string        `validate:"empty=false"  yaml:"id"`
	Kind         string        `validate:"empty=false" yaml:"kind"`
	PollInterval time.Duration `default:"30s" yaml:"poll_interval"`

	S3Config struct {
		BucketName string `default:"30s" validate:"empty=false" yaml:"bucket_name"`
	} `yaml:"s3_config"`
}

func (s *syncDaemonConfigSourceYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonConfigSourceYaml // declare new type to prevent stackoverflow
	t := (*T)(s)
	if err := unmarshal(t); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type syncDaemonConfigCommandEntryYaml struct {
	Command   []string `validate:"empty=false" yaml:"command"`
	OnFailure string   `default:"fail_sync" validate:"one_of=ignore,fail_sync,panic" yaml:"on_failure"` // one_of
}

func (s *syncDaemonConfigCommandEntryYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonConfigCommandEntryYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type syncDaemonConfigRuleYaml struct {
	ID          string                             `validate:"empty=false" yaml:"id"`
	Source      string                             `validate:"empty=false" yaml:"source"`
	Prefix      string                             `default:"" yaml:"prefix"`
	MaxFailures int                                `default:"-1" yaml:"max_failures"`
	AfterSync   []syncDaemonConfigCommandEntryYaml `yaml:"after_sync"`
	Includes    []syncDaemonConfigRuleIncludesYaml `validate:"empty=false" yaml:"includes"`
}

func (s *syncDaemonConfigRuleYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonConfigRuleYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

type syncDaemonConfigRuleIncludesYaml struct {
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Cleanup bool   `default:"false" yaml:"cleanup"`
}

func (s *syncDaemonConfigRuleIncludesYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonConfigRuleIncludesYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}
