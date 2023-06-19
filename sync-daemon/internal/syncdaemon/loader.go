package syncdaemon

import (
	"fmt"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/collections"
	"gopkg.in/yaml.v3"
	"os"
)

func loadConfig(filename string, out *syncDaemonConfig) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var yamlConfig syncDaemonConfigYaml
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return err
	}

	if err := yamlConfig.toSyncConfig(out); err != nil {
		return err
	}
	return nil
}

func (yamlConfig *syncDaemonConfigYaml) toSyncConfig(out *syncDaemonConfig) error {
	sourcesMap := make(map[string]*syncDaemonConfigSource, len(yamlConfig.Sources))
	sourcesList := make([]*syncDaemonConfigSource, 0, len(yamlConfig.Sources))

	for _, yamlSource := range yamlConfig.Sources {
		source := syncDaemonConfigSource{
			id:           yamlSource.ID,
			kind:         syncDaemonConfigSourceKindMap[yamlSource.Kind],
			pollInterval: yamlSource.PollInterval,
			rules:        make([]syncDaemonConfigSyncRule, 0),
			s3BucketName: yamlSource.S3Config.BucketName,
		}
		sourcesMap[source.id] = &source
		sourcesList = append(sourcesList, &source)
	}

	for _, yamlRule := range yamlConfig.Rules {
		sourceId := yamlRule.Source
		source, ok := sourcesMap[sourceId]
		if !ok {
			return fmt.Errorf("rule id=%s: source with id=%s is not found", yamlRule.ID, sourceId)
		}

		includes := collections.Map(yamlRule.Includes, func(t syncDaemonConfigRuleIncludesYaml) syncDaemonConfigIncludeRule {
			return syncDaemonConfigIncludeRule{from: t.From, to: t.To, cleanup: t.Cleanup}
		})

		afterSync := collections.Map(yamlRule.AfterSync, func(t syncDaemonConfigCommandEntryYaml) syncDaemonConfigCommandEntry {
			return syncDaemonConfigCommandEntry{command: t.Command, onFailure: syncDaemonConfigOnFailureTMap[t.OnFailure]}
		})

		rule := syncDaemonConfigSyncRule{
			id:          yamlRule.ID,
			prefix:      yamlRule.Prefix,
			maxFailures: yamlRule.MaxFailures,
			AfterSync:   afterSync,
			Includes:    includes,
		}

		source.rules = append(source.rules, rule)
	}
	out.sources = collections.Map(sourcesList, func(s *syncDaemonConfigSource) syncDaemonConfigSource { return *s })
	return nil
}
