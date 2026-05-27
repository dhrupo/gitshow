// Package main is the gitshow CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/dhrupo/gitshow/cmd/gitshow/commands"
)

// Version is the gitshow build version.  Override at link time:
//
//	go build -ldflags "-X main.Version=0.1.0" ./cmd/gitshow
var Version = "1.0.0"

func main() {
	if err := commands.NewRootCmd(Version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gitshow:", err)
		os.Exit(1)
	}
}
