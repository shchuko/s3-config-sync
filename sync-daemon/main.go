package main

import (
	"flag"
	"fmt"
	"github.com/shchuko/s3-sync-config/sync-daemon/internal/syncdaemon"
	"os"
)

// args
var (
	printHelp  bool
	configPath string
)

func init() {
	flag.BoolVar(&printHelp, "help", false, "Print this help")
	flag.StringVar(&configPath, "config", "/etc/sync-daemon-config.yaml", "Path to sync-daemon config")
}

func parseArgs() {
	flag.Parse()
	if printHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {
	parseArgs()

	daemon := syncdaemon.NewSyncDaemon(configPath)
	err := daemon.Run()
	if err != nil {
		fmt.Println("Sync Daemon error:", err)
		os.Exit(1)
	}
}
