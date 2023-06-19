package syncdaemon

import "time"

type syncDaemonConfig struct {
	sources []syncDaemonConfigSource
}

type syncDaemonConfigSourceKind int8

const (
	s3 syncDaemonConfigSourceKind = 0
)

var syncDaemonConfigSourceKindMap = map[string]syncDaemonConfigSourceKind{"s3": s3}

type syncDaemonConfigSource struct {
	id           string
	kind         syncDaemonConfigSourceKind
	pollInterval time.Duration
	rules        []syncDaemonConfigSyncRule

	s3BucketName string
}

type syncDaemonConfigSyncRule struct {
	id          string
	prefix      string
	maxFailures int
	AfterSync   []syncDaemonConfigCommandEntry
	Includes    []syncDaemonConfigIncludeRule
}

type syncDaemonConfigCommandOnFailureT int8

const (
	failSync syncDaemonConfigCommandOnFailureT = 0
	ignore   syncDaemonConfigCommandOnFailureT = 1
	panic_   syncDaemonConfigCommandOnFailureT = 2
)

var syncDaemonConfigOnFailureTMap = map[string]syncDaemonConfigCommandOnFailureT{"fail_sync": failSync, "ignore": ignore, "panic": panic_}

type syncDaemonConfigCommandEntry struct {
	command   []string
	onFailure syncDaemonConfigCommandOnFailureT
}

type syncDaemonConfigIncludeRule struct {
	from    string // TODO filematcher
	to      string
	cleanup bool
}
