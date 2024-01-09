package main

import (
	"fmt"
)

var (
	version = "dev"
	date    = "notset"
	commit  = ""
	builtBy = ""
)

func init() {
	rootCmd.Version = fmt.Sprintf("%s [%s] (%s) <%s>", version, commit, date, builtBy)
}
