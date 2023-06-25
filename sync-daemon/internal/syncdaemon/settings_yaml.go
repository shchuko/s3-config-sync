package syncdaemon

import (
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/syncdaemon/sources/s3_source"
	"gopkg.in/dealancer/validate.v2"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"time"
)

type syncDaemonSettingsYaml struct {
	Sources []syncDaemonSourceYaml     `validate:"empty=false" yaml:"sources"`
	Rules   []syncDaemonConfigRuleYaml `validate:"empty=false" yaml:"rules"`
}

func (s *syncDaemonSettingsYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonSettingsYaml // declare new type to prevent stackoverflow
	if err := unmarshal((*T)(s)); err != nil {
		return err
	}
	if err := validate.Validate(s); err != nil {
		return err
	}
	return nil
}

func (s *syncDaemonSettingsYaml) Validate() error {
	sourcesIdsUnique, nonUniqueSourceId := collections.CheckUnique(s.Sources, func(item syncDaemonSourceYaml) string {
		return item.ID
	})
	if !sourcesIdsUnique {
		return fmt.Errorf("invalid sources: non-unique id='%v'", *nonUniqueSourceId)
	}

	rulesIdsUnique, nonUniqueRuleId := collections.CheckUnique(s.Rules, func(item syncDaemonConfigRuleYaml) string {
		return item.ID
	})
	if !rulesIdsUnique {
		return fmt.Errorf("invalid rules: non-unique id='%v'", *nonUniqueRuleId)
	}
	return nil
}

type syncDaemonSourceYaml struct {
	ID           string        `validate:"empty=false"  yaml:"id"`
	Kind         string        `validate:"empty=false" yaml:"kind"`
	PollInterval time.Duration `default:"30s" yaml:"poll_interval"`

	S3Config struct {
		BucketName string `validate:"empty=false" yaml:"bucket_name"`
		Region     string `validate:"empty=false" yaml:"region"`
	} `yaml:"s3_config"`
}

func (s *syncDaemonSourceYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(s); err != nil {
		return err
	}
	type T syncDaemonSourceYaml // declare new type to prevent stackoverflow
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
	OnFailure string   `default:"fail_sync" validate:"one_of=ignore,fail_sync" yaml:"on_failure"` // one_of
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

func readConfig(filename string, out *syncDaemonSettings) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	str, err := resolveEnvVars(string(data))
	if err != nil {
		return err
	}

	var yamlConfig syncDaemonSettingsYaml
	if err := yaml.Unmarshal([]byte(str), &yamlConfig); err != nil {
		return err
	}

	if err := yamlConfig.toSyncConfig(out); err != nil {
		return err
	}
	return nil
}

func (s *syncDaemonSettingsYaml) toSyncConfig(out *syncDaemonSettings) error {
	sourceSettingsMap := make(map[string]*syncSourceSettings, len(s.Sources))
	sourceSettingsList := make([]*syncSourceSettings, 0, len(s.Sources))

	for _, yamlSource := range s.Sources {
		var source SyncConfigSource

		switch yamlSource.Kind {
		case "s3":
			source = s3_source.NewSyncSourceS3(yamlSource.S3Config.BucketName, yamlSource.S3Config.Region)
		default:
			return errors.New("Unknown source type: " + yamlSource.ID)
		}

		sourceSettings := syncSourceSettings{
			source:       source,
			id:           yamlSource.ID,
			pollInterval: yamlSource.PollInterval,
			rules:        make([]syncRule, 0),
		}

		sourceSettingsMap[sourceSettings.id] = &sourceSettings
		sourceSettingsList = append(sourceSettingsList, &sourceSettings)
	}

	for _, yamlRule := range s.Rules {
		sourceId := yamlRule.Source
		source, ok := sourceSettingsMap[sourceId]
		if !ok {
			return fmt.Errorf("rule id=%s: source with id=%s is not found", yamlRule.ID, sourceId)
		}

		includes := collections.Map(yamlRule.Includes, func(t syncDaemonConfigRuleIncludesYaml) syncRuleInclude {
			return *newSyncRuleInclude(t.From, t.To, t.Cleanup)
		})

		afterSync := collections.Map(yamlRule.AfterSync, func(t syncDaemonConfigCommandEntryYaml) syncRuleCommand {
			return syncRuleCommand{command: t.Command, onFailure: errorCardinalityMap[t.OnFailure]}
		})

		rule := syncRule{
			id:          yamlRule.ID,
			prefix:      yamlRule.Prefix,
			maxFailures: yamlRule.MaxFailures,
			afterSync:   afterSync,
			includes:    includes,
		}

		source.rules = append(source.rules, rule)
	}

	unusedSources := collections.Filter(sourceSettingsList, func(t *syncSourceSettings) bool { return len(t.rules) == 0 })
	if len(unusedSources) != 0 {
		unusedSourcesIds := collections.Map(unusedSources, func(t *syncSourceSettings) string { return t.id })
		return fmt.Errorf("unused sources found: %v", unusedSourcesIds)
	}

	out.sources = collections.Map(sourceSettingsList, func(s *syncSourceSettings) syncSourceSettings { return *s })
	return nil
}

func resolveEnvVars(s string) (string, error) {
	varRegex := `\$\{env\:([A-Za-z_][A-Za-z0-9_]*)\}`
	re := regexp.MustCompile(varRegex)

	matches := collections.DistinctBy(re.FindAllStringSubmatch(s, -1), func(t []string) string { return t[0] })
	if matches == nil {
		return s, nil
	}

	for _, match := range matches {
		varName := match[1]
		value, ok := os.LookupEnv(varName)
		if !ok {
			return "", fmt.Errorf("env variable '%s' is used in config but not set", varName)
		}
		s = strings.ReplaceAll(s, match[0], value)
	}

	return s, nil
}
