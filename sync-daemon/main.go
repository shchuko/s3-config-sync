package main

import (
	"flag"
	"fmt"
	loader "github.com/shchuko/s3-sync-config/sync-daemon/internal/config_loader"
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
	var config loader.SyncConfig

	if err := loader.LoadConfig(configPath, &config); err != nil {
		fmt.Println("Error loading syncConfig:", err)
		os.Exit(1)
	}
	fmt.Printf("%#v", config)
}
