package config_parser

import "time"

type SyncConfig struct {
	Sources []SyncSource
}

type SyncSourceKind int8

const (
	S3 SyncSourceKind = 0
)

var syncSourceKindMap = map[string]SyncSourceKind{"s3": S3}

type SyncSource struct {
	ID           string
	Kind         SyncSourceKind
	PollInterval time.Duration
	Rules        []SyncRule

	S3BucketName string
}

type SyncRule struct {
	ID          string
	Prefix      string
	MaxFailures int
	AfterSync   []CommandEntry
	Includes    []IncludeRule
}

type OnFailureT int8

const (
	FailSync OnFailureT = 0
	Ignore   OnFailureT = 1
	Panic    OnFailureT = 2
)

var onFailureTMap = map[string]OnFailureT{"fail_sync": FailSync, "ignore": Ignore, "panic": Panic}

type CommandEntry struct {
	Command   []string
	OnFailure OnFailureT
}

type IncludeRule struct {
	From    string // TODO filematcher
	To      string
	Cleanup bool
}
