package main

import (
	"github.com/gianarb/papertrail-archive/cmd"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	viper.SetEnvPrefix("PAPERTRAIL_ARCHIVE")
	viper.AutomaticEnv()
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.Execute()
}
