package config_parser

import (
	"fmt"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadConfig(filename string, out *SyncConfig) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var yamlConfig SyncConfigYaml
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return err
	}

	if err := yamlConfig.toSyncConfig(out); err != nil {
		return err
	}
	return nil
}

func (yamlConfig *SyncConfigYaml) toSyncConfig(out *SyncConfig) error {
	sourcesMap := make(map[string]*SyncSource, len(yamlConfig.Sources))
	sourcesList := make([]*SyncSource, 0, len(yamlConfig.Sources))

	for _, yamlSource := range yamlConfig.Sources {
		source := SyncSource{
			ID:           yamlSource.ID,
			Kind:         syncSourceKindMap[yamlSource.Kind],
			PollInterval: yamlSource.PollInterval,
			Rules:        make([]SyncRule, 0),
			S3BucketName: yamlSource.S3Config.BucketName,
		}
		sourcesMap[source.ID] = &source
		sourcesList = append(sourcesList, &source)
	}

	for _, yamlRule := range yamlConfig.Rules {
		sourceId := yamlRule.Source
		source, ok := sourcesMap[sourceId]
		if !ok {
			return fmt.Errorf("rule id=%s: source with id=%s is not found", yamlRule.ID, sourceId)
		}

		includes := collections.Map(yamlRule.Includes, func(t RuleIncludesYaml) IncludeRule {
			return IncludeRule{From: t.From, To: t.To, Cleanup: t.Cleanup}
		})

		afterSync := collections.Map(yamlRule.AfterSync, func(t CommandEntryYaml) CommandEntry {
			return CommandEntry{Command: t.Command, OnFailure: onFailureTMap[t.OnFailure]}
		})

		rule := SyncRule{
			ID:          yamlRule.ID,
			Prefix:      yamlRule.Prefix,
			MaxFailures: yamlRule.MaxFailures,
			AfterSync:   afterSync,
			Includes:    includes,
		}

		source.Rules = append(source.Rules, rule)
	}
	out.Sources = collections.Map(sourcesList, func(s *SyncSource) SyncSource { return *s })
	return nil
}
