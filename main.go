package main

import (
	"fmt"
	"os"

	"kick/cmd"
)

// 通过 ldflags 注入的版本信息
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	// 添加版本信息
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built at: %s)", version, commit, date)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
